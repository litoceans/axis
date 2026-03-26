package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/godlabs/axis/pkg/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Storage provides database access for Axis
type Storage struct {
	db *sql.DB
}

// New creates a new storage instance
func New(dbPath string) (*Storage, error) {
	// Handle sqlite:// URL format
	actualPath := dbPath
	if len(dbPath) > 9 && dbPath[:9] == "sqlite://" {
		actualPath = dbPath[9:]
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(actualPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", actualPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	s := &Storage{db: db}

	// Initialize schema
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

// initSchema creates the database schema
func (s *Storage) initSchema() error {
	schema := `
	-- Organizations
	CREATE TABLE IF NOT EXISTS orgs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		plan TEXT DEFAULT 'free',
		monthly_budget_usd REAL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Teams
	CREATE TABLE IF NOT EXISTS teams (
		id TEXT PRIMARY KEY,
		org_id TEXT REFERENCES orgs(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Members
	CREATE TABLE IF NOT EXISTS members (
		id TEXT PRIMARY KEY,
		org_id TEXT REFERENCES orgs(id) ON DELETE CASCADE,
		email TEXT UNIQUE NOT NULL,
		name TEXT,
		role TEXT CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- API Keys
	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		key_hash TEXT UNIQUE NOT NULL,
		key_prefix TEXT NOT NULL,
		key_name TEXT,
		org_id TEXT REFERENCES orgs(id) ON DELETE CASCADE,
		team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
		member_id TEXT REFERENCES members(id) ON DELETE SET NULL,
		scopes TEXT,
		models TEXT,
		rpm_limit INTEGER,
		tpm_limit INTEGER,
		monthly_budget_usd REAL,
		environments TEXT,
		expires_at TIMESTAMPTZ,
		last_used_at TIMESTAMPTZ,
		revoked_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Usage Logs
	CREATE TABLE IF NOT EXISTS usage_logs (
		id TEXT PRIMARY KEY,
		key_id TEXT,
		org_id TEXT,
		model TEXT NOT NULL,
		provider TEXT NOT NULL,
		endpoint TEXT NOT NULL,
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		cached_tokens INTEGER DEFAULT 0,
		cost_usd REAL DEFAULT 0,
		latency_ms INTEGER DEFAULT 0,
		status_code INTEGER,
		cached BOOLEAN DEFAULT 0,
		error BOOLEAN DEFAULT 0,
		error_type TEXT,
		trace_id TEXT,
		tokens_input_reported INTEGER DEFAULT 0,
		tokens_output_reported INTEGER DEFAULT 0,
		tokens_input_calculated INTEGER DEFAULT 0,
		tokens_output_calculated INTEGER DEFAULT 0,
		cost_reported REAL DEFAULT 0,
		cost_calculated REAL DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Routing Chains
	CREATE TABLE IF NOT EXISTS routing_chains (
		id TEXT PRIMARY KEY,
		org_id TEXT REFERENCES orgs(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		chains TEXT NOT NULL,
		is_default BOOLEAN DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Cache Entries (metadata only)
	CREATE TABLE IF NOT EXISTS cache_entries (
		id TEXT PRIMARY KEY,
		key_id TEXT,
		org_id TEXT,
		model TEXT NOT NULL,
		prompt_hash TEXT NOT NULL,
		response_hash TEXT NOT NULL,
		cost_usd REAL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMPTZ
	);

	-- Audit Log
	CREATE TABLE IF NOT EXISTS audit_log (
		id TEXT PRIMARY KEY,
		org_id TEXT,
		actor_id TEXT,
		actor_type TEXT,
		action TEXT NOT NULL,
		target_type TEXT,
		target_id TEXT,
		metadata TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Budget Alerts
	CREATE TABLE IF NOT EXISTS budget_alerts (
		id TEXT PRIMARY KEY,
		key_id TEXT NOT NULL,
		org_id TEXT NOT NULL,
		threshold_percent REAL NOT NULL,
		spent_usd REAL NOT NULL,
		limit_usd REAL NOT NULL,
		triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		acknowledged BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (key_id) REFERENCES api_keys(id)
	);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Info().Msg("database schema initialized")
	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// StoreKey stores an API key
func (s *Storage) StoreKey(ctx context.Context, key *APIKey) error {
	query := `
	INSERT INTO api_keys (id, key_hash, key_prefix, key_name, org_id, team_id, member_id, scopes, models, rpm_limit, tpm_limit, monthly_budget_usd, environments, expires_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	scopesJSON, _ := json.Marshal(key.Scopes)
	modelsJSON, _ := json.Marshal(key.Models)
	envsJSON, _ := json.Marshal(key.Environments)

	_, err := s.db.ExecContext(ctx, query,
		key.ID, key.KeyHash, key.KeyPrefix, key.KeyName, key.OrgID, key.TeamID, key.MemberID,
		string(scopesJSON), string(modelsJSON), key.RPMLimit, key.TPMLimit, key.MonthlyBudgetUSD,
		string(envsJSON), key.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to store key: %w", err)
	}

	return nil
}

// GetKeyByHash retrieves an API key by its hash
func (s *Storage) GetKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `
	SELECT id, key_hash, key_prefix, key_name, org_id, team_id, member_id, scopes, models, rpm_limit, tpm_limit, monthly_budget_usd, environments, expires_at, last_used_at, revoked_at, created_at
	FROM api_keys
	WHERE key_hash = ? AND revoked_at IS NULL
	`

	var key APIKey
	var scopesStr, modelsStr, envsStr string
	var lastUsedAt, revokedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, keyHash).Scan(
		&key.ID, &key.KeyHash, &key.KeyPrefix, &key.KeyName, &key.OrgID, &key.TeamID, &key.MemberID,
		&scopesStr, &modelsStr, &key.RPMLimit, &key.TPMLimit, &key.MonthlyBudgetUSD, &envsStr,
		&key.ExpiresAt, &lastUsedAt, &revokedAt, &key.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	json.Unmarshal([]byte(scopesStr), &key.Scopes)
	json.Unmarshal([]byte(modelsStr), &key.Models)
	json.Unmarshal([]byte(envsStr), &key.Environments)

	if lastUsedAt.Valid {
		key.LastUsedAt = lastUsedAt.Time
	}

	return &key, nil
}

// GetKeyByID retrieves an API key by its ID
func (s *Storage) GetKeyByID(ctx context.Context, keyID string) (*APIKey, error) {
	query := `
	SELECT id, key_hash, key_prefix, key_name, org_id, team_id, member_id, scopes, models, rpm_limit, tpm_limit, monthly_budget_usd, environments, expires_at, last_used_at, revoked_at, created_at
	FROM api_keys
	WHERE id = ? AND revoked_at IS NULL
	`

	var key APIKey
	var scopesStr, modelsStr, envsStr string
	var lastUsedAt, revokedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, keyID).Scan(
		&key.ID, &key.KeyHash, &key.KeyPrefix, &key.KeyName, &key.OrgID, &key.TeamID, &key.MemberID,
		&scopesStr, &modelsStr, &key.RPMLimit, &key.TPMLimit, &key.MonthlyBudgetUSD, &envsStr,
		&key.ExpiresAt, &lastUsedAt, &revokedAt, &key.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	json.Unmarshal([]byte(scopesStr), &key.Scopes)
	json.Unmarshal([]byte(modelsStr), &key.Models)
	json.Unmarshal([]byte(envsStr), &key.Environments)

	if lastUsedAt.Valid {
		key.LastUsedAt = lastUsedAt.Time
	}

	return &key, nil
}

// LogUsage logs a request
func (s *Storage) LogUsage(ctx context.Context, log *types.UsageLog) error {
	query := `
	INSERT INTO usage_logs (id, key_id, org_id, model, provider, endpoint, input_tokens, output_tokens, cached_tokens, cost_usd, latency_ms, status_code, cached, error, error_type, trace_id, tokens_input_reported, tokens_output_reported, tokens_input_calculated, tokens_output_calculated, cost_reported, cost_calculated, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		log.ID, log.KeyID, log.OrgID, log.Model, log.Provider, log.Endpoint,
		log.InputTokens, log.OutputTokens, log.CachedTokens, log.CostUSD, log.LatencyMs,
		log.StatusCode, log.Cached, log.Error, log.ErrorType, log.TraceID,
		log.TokensInputReported, log.TokensOutputReported, log.TokensInputCalculated,
		log.TokensOutputCalculated, log.CostReported, log.CostCalculated, log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to log usage: %w", err)
	}

	return nil
}

// GetUsage retrieves usage for a key within a time range
func (s *Storage) GetUsage(ctx context.Context, keyID, orgID string, from, to time.Time) ([]*types.UsageLog, error) {
	query := `
	SELECT id, key_id, org_id, model, provider, endpoint, input_tokens, output_tokens, cached_tokens, cost_usd, latency_ms, status_code, cached, error, error_type, trace_id, created_at
	FROM usage_logs
	WHERE key_id = ? AND created_at BETWEEN ? AND ?
	ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, keyID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage: %w", err)
	}
	defer rows.Close()

	var logs []*types.UsageLog
	for rows.Next() {
		var l types.UsageLog
		err := rows.Scan(&l.ID, &l.KeyID, &l.OrgID, &l.Model, &l.Provider, &l.Endpoint,
			&l.InputTokens, &l.OutputTokens, &l.CachedTokens, &l.CostUSD, &l.LatencyMs,
			&l.StatusCode, &l.Cached, &l.Error, &l.ErrorType, &l.TraceID, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, &l)
	}

	return logs, nil
}

// GetKeyCost retrieves total cost for a key within a time range
func (s *Storage) GetKeyCost(ctx context.Context, keyID string, from, to time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(cost_usd), 0) FROM usage_logs WHERE key_id = ? AND created_at BETWEEN ? AND ?`

	var cost float64
	err := s.db.QueryRowContext(ctx, query, keyID, from, to).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("failed to get key cost: %w", err)
	}

	return cost, nil
}

// GetKeySpend retrieves total spend for a key in the current calendar month
func (s *Storage) GetKeySpend(ctx context.Context, keyID string) (float64, error) {
	// Get current calendar month bounds
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	query := `SELECT COALESCE(SUM(cost_usd), 0) FROM usage_logs WHERE key_id = ? AND created_at >= ? AND created_at < ?`

	var spend float64
	err := s.db.QueryRowContext(ctx, query, keyID, monthStart, monthEnd).Scan(&spend)
	if err != nil {
		return 0, fmt.Errorf("failed to get key spend: %w", err)
	}

	return spend, nil
}

// UpdateKeyLastUsed updates the last_used_at timestamp for a key
func (s *Storage) UpdateKeyLastUsed(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to update key last used: %w", err)
	}
	return nil
}

// CreateOrg creates a new organization
func (s *Storage) CreateOrg(ctx context.Context, org *Org) error {
	query := `INSERT INTO orgs (id, name, plan, monthly_budget_usd) VALUES (?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, org.ID, org.Name, org.Plan, org.MonthlyBudgetUSD)
	if err != nil {
		return fmt.Errorf("failed to create org: %w", err)
	}
	return nil
}

// GetOrg retrieves an organization
func (s *Storage) GetOrg(ctx context.Context, orgID string) (*Org, error) {
	query := `SELECT id, name, plan, monthly_budget_usd, created_at, updated_at FROM orgs WHERE id = ?`

	var org Org
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&org.ID, &org.Name, &org.Plan, &org.MonthlyBudgetUSD, &org.CreatedAt, &org.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get org: %w", err)
	}

	return &org, nil
}

// ListKeys retrieves all API keys for an organization
func (s *Storage) ListKeys(ctx context.Context, orgID string) ([]*APIKey, error) {
	query := `
	SELECT id, key_hash, key_prefix, key_name, org_id, team_id, member_id, scopes, models, rpm_limit, tpm_limit, monthly_budget_usd, environments, expires_at, last_used_at, revoked_at, created_at
	FROM api_keys
	WHERE org_id = ? AND revoked_at IS NULL
	ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		var scopesStr, modelsStr, envsStr string
		var lastUsedAt, revokedAt sql.NullTime

		err := rows.Scan(
			&key.ID, &key.KeyHash, &key.KeyPrefix, &key.KeyName, &key.OrgID, &key.TeamID, &key.MemberID,
			&scopesStr, &modelsStr, &key.RPMLimit, &key.TPMLimit, &key.MonthlyBudgetUSD, &envsStr,
			&key.ExpiresAt, &lastUsedAt, &revokedAt, &key.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}

		json.Unmarshal([]byte(scopesStr), &key.Scopes)
		json.Unmarshal([]byte(modelsStr), &key.Models)
		json.Unmarshal([]byte(envsStr), &key.Environments)

		if lastUsedAt.Valid {
			key.LastUsedAt = lastUsedAt.Time
		}

		keys = append(keys, &key)
	}

	return keys, nil
}

// CreateKey stores a new API key
func (s *Storage) CreateKey(ctx context.Context, key *APIKey) error {
	return s.StoreKey(ctx, key)
}

// DeleteKey revokes an API key
func (s *Storage) DeleteKey(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("key not found")
	}

	return nil
}

// RotateKey rotates an API key by revoking the old one
func (s *Storage) RotateKey(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to rotate key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("key not found")
	}

	return nil
}

// GetUsageDetailed retrieves usage for an org within a time range
func (s *Storage) GetUsageDetailed(ctx context.Context, orgID string, from, to time.Time) ([]*types.UsageLog, error) {
	query := `
	SELECT id, key_id, org_id, model, provider, endpoint, input_tokens, output_tokens, cached_tokens, cost_usd, latency_ms, status_code, cached, error, error_type, trace_id, created_at
	FROM usage_logs
	WHERE org_id = ? AND created_at BETWEEN ? AND ?
	ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage: %w", err)
	}
	defer rows.Close()

	var logs []*types.UsageLog
	for rows.Next() {
		var l types.UsageLog
		err := rows.Scan(&l.ID, &l.KeyID, &l.OrgID, &l.Model, &l.Provider, &l.Endpoint,
			&l.InputTokens, &l.OutputTokens, &l.CachedTokens, &l.CostUSD, &l.LatencyMs,
			&l.StatusCode, &l.Cached, &l.Error, &l.ErrorType, &l.TraceID, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, &l)
	}

	return logs, nil
}

// GetCostsByModel retrieves cost breakdown by model for an org
func (s *Storage) GetCostsByModel(ctx context.Context, orgID string, from, to time.Time) ([]*types.CostBreakdown, error) {
	query := `
	SELECT model, provider,
		COUNT(*) as requests,
		COALESCE(SUM(input_tokens), 0) as tokens_in,
		COALESCE(SUM(output_tokens), 0) as tokens_out,
		COALESCE(SUM(cost_usd), 0) as cost_usd
	FROM usage_logs
	WHERE org_id = ? AND created_at BETWEEN ? AND ?
	GROUP BY model, provider
	ORDER BY cost_usd DESC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query costs: %w", err)
	}
	defer rows.Close()

	var costs []*types.CostBreakdown
	for rows.Next() {
		var c types.CostBreakdown
		err := rows.Scan(&c.Model, &c.Provider, &c.Requests, &c.TokensIn, &c.TokensOut, &c.CostUsd)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost: %w", err)
		}
		costs = append(costs, &c)
	}

	return costs, nil
}

// ListRoutingChains retrieves routing chains for an organization
func (s *Storage) ListRoutingChains(ctx context.Context, orgID string) ([]*types.RoutingChain, error) {
	query := `
	SELECT id, org_id, name, chains, is_default
	FROM routing_chains
	WHERE org_id = ?
	ORDER BY is_default DESC, created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list routing chains: %w", err)
	}
	defer rows.Close()

	var chains []*types.RoutingChain
	for rows.Next() {
		var chain types.RoutingChain
		var chainsJSON string
		var orgIDTmp string

		err := rows.Scan(&chain.ID, &orgIDTmp, &chain.Name, &chainsJSON, &chain.IsDefault)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routing chain: %w", err)
		}

		// Parse chains JSON into ChainModel slice
		var chainSteps []types.ChainStep
		if err := json.Unmarshal([]byte(chainsJSON), &chainSteps); err != nil {
			return nil, fmt.Errorf("failed to parse chains JSON: %w", err)
		}

		// Convert ChainStep to ChainModel
		chain.Models = make([]types.ChainModel, len(chainSteps))
		for i, step := range chainSteps {
			chain.Models[i] = types.ChainModel{
				Model:      step.Model,
				Provider:   step.Provider,
				MaxLatency: float64(step.MaxLatencyMs),
				Retries:    step.MaxRetries,
				Weight:     step.Weight,
				FailOpen:   step.FailOpen,
			}
		}

		chains = append(chains, &chain)
	}

	return chains, nil
}

// CreateRoutingChain creates a new routing chain
func (s *Storage) CreateRoutingChain(ctx context.Context, chain *types.RoutingChain) error {
	query := `
	INSERT INTO routing_chains (id, org_id, name, chains, is_default)
	VALUES (?, ?, ?, ?, ?)
	`

	// Convert ChainModel back to ChainStep for storage
	chainSteps := make([]types.ChainStep, len(chain.Models))
	for i, m := range chain.Models {
		chainSteps[i] = types.ChainStep{
			Model:        m.Model,
			Provider:     m.Provider,
			MaxLatencyMs: int(m.MaxLatency),
			MaxRetries:   m.Retries,
			Weight:       m.Weight,
			FailOpen:     m.FailOpen,
		}
	}

	chainsJSON, err := json.Marshal(chainSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal chains: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, chain.ID, chain.Models[0].Model, chain.Name, string(chainsJSON), chain.IsDefault)
	if err != nil {
		return fmt.Errorf("failed to create routing chain: %w", err)
	}

	return nil
}

// CreateBudgetAlert creates a new budget alert
func (s *Storage) CreateBudgetAlert(ctx context.Context, alert *BudgetAlert) error {
	query := `
	INSERT INTO budget_alerts (id, key_id, org_id, threshold_percent, spent_usd, limit_usd, triggered_at, acknowledged)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		alert.ID, alert.KeyID, alert.OrgID, alert.ThresholdPercent,
		alert.SpentUSD, alert.LimitUSD, alert.TriggeredAt, alert.Acknowledged)

	if err != nil {
		return fmt.Errorf("failed to create budget alert: %w", err)
	}

	return nil
}

// ListBudgetAlerts retrieves budget alerts for an org and optionally a key
func (s *Storage) ListBudgetAlerts(ctx context.Context, orgID, keyID string) ([]*BudgetAlert, error) {
	var query string
	var args []interface{}

	if keyID != "" {
		query = `
		SELECT id, key_id, org_id, threshold_percent, spent_usd, limit_usd, triggered_at, acknowledged
		FROM budget_alerts
		WHERE org_id = ? AND key_id = ?
		ORDER BY triggered_at DESC
		`
		args = []interface{}{orgID, keyID}
	} else {
		query = `
		SELECT id, key_id, org_id, threshold_percent, spent_usd, limit_usd, triggered_at, acknowledged
		FROM budget_alerts
		WHERE org_id = ?
		ORDER BY triggered_at DESC
		`
		args = []interface{}{orgID}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list budget alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*BudgetAlert
	for rows.Next() {
		var alert BudgetAlert
		err := rows.Scan(&alert.ID, &alert.KeyID, &alert.OrgID, &alert.ThresholdPercent,
			&alert.SpentUSD, &alert.LimitUSD, &alert.TriggeredAt, &alert.Acknowledged)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget alert: %w", err)
		}
		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// AcknowledgeAlert marks a budget alert as acknowledged
func (s *Storage) AcknowledgeAlert(ctx context.Context, alertID string) error {
	query := `UPDATE budget_alerts SET acknowledged = TRUE WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, alertID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// GetAlertForThreshold checks if an alert already exists for a key/threshold in current month
func (s *Storage) GetAlertForThreshold(ctx context.Context, keyID string, thresholdPercent float64) (*BudgetAlert, error) {
	// Get current calendar month bounds
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := `
	SELECT id, key_id, org_id, threshold_percent, spent_usd, limit_usd, triggered_at, acknowledged
	FROM budget_alerts
	WHERE key_id = ? AND threshold_percent = ? AND triggered_at >= ?
	LIMIT 1
	`

	var alert BudgetAlert
	err := s.db.QueryRowContext(ctx, query, keyID, thresholdPercent, monthStart).Scan(
		&alert.ID, &alert.KeyID, &alert.OrgID, &alert.ThresholdPercent,
		&alert.SpentUSD, &alert.LimitUSD, &alert.TriggeredAt, &alert.Acknowledged)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return &alert, nil
}

// APIKey represents an API key in the database
type APIKey struct {
	ID               string     `json:"id"`
	KeyHash          string     `json:"key_hash"`
	KeyPrefix        string     `json:"key_prefix"`
	KeyName          string     `json:"key_name,omitempty"`
	OrgID            string     `json:"org_id,omitempty"`
	TeamID           string     `json:"team_id,omitempty"`
	MemberID         string     `json:"member_id,omitempty"`
	Scopes           []string   `json:"scopes,omitempty"`
	Models           []string   `json:"models,omitempty"`
	RPMLimit         int        `json:"rpm_limit,omitempty"`
	TPMLimit         int        `json:"tpm_limit,omitempty"`
	MonthlyBudgetUSD float64    `json:"monthly_budget_usd,omitempty"`
	Environments     []string   `json:"environments,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	LastUsedAt       time.Time  `json:"last_used_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Org represents an organization in the database
type Org struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Plan             string    `json:"plan"`
	MonthlyBudgetUSD float64   `json:"monthly_budget_usd,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BudgetAlert represents a budget alert in the database
type BudgetAlert struct {
	ID                 string    `json:"id"`
	KeyID              string    `json:"key_id"`
	OrgID              string    `json:"org_id"`
	ThresholdPercent   float64   `json:"threshold_percent"`
	SpentUSD           float64   `json:"spent_usd"`
	LimitUSD           float64   `json:"limit_usd"`
	TriggeredAt        time.Time `json:"triggered_at"`
	Acknowledged       bool      `json:"acknowledged"`
}
