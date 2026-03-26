import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { teams } from '@/lib/mockData'

export function Settings() {
  const [activeTab, setActiveTab] = useState('team')

  const tabs = [
    { id: 'team', label: 'Team' },
    { id: 'providers', label: 'Providers' },
    { id: 'alerts', label: 'Alert Channels' },
    { id: 'security', label: 'Security' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-title">Settings</h1>
        <p className="text-body text-text-muted mt-1">
          Manage your organization and preferences
        </p>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-border">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-body transition-colors ${
              activeTab === tab.id
                ? 'text-accent border-b-2 border-accent'
                : 'text-text-muted hover:text-text-primary'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Team Management */}
      {activeTab === 'team' && (
        <div className="space-y-6">
          <Card>
            <div className="flex items-center justify-between mb-4">
              <h3 className="section-title">Teams</h3>
              <Button variant="secondary" size="sm">Add Team</Button>
            </div>
            <div className="space-y-4">
              {teams.map((team) => (
                <div key={team.id} className="flex items-center justify-between py-3 border-b border-border/50 last:border-0">
                  <div>
                    <p className="text-body font-medium text-text-primary">{team.name}</p>
                    <p className="text-micro text-text-muted">{team.members} members · {team.keys} keys</p>
                  </div>
                  <Button variant="ghost" size="sm">Manage</Button>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {/* Provider Keys */}
      {activeTab === 'providers' && (
        <Card>
          <h3 className="section-title mb-4">Provider API Keys</h3>
          <div className="space-y-4">
            {['OpenAI', 'Anthropic', 'Google', 'Ollama'].map((provider) => (
              <div key={provider} className="flex items-center gap-4">
                <div className="w-32">
                  <p className="text-body font-medium text-text-primary">{provider}</p>
                </div>
                <Input
                  type="password"
                  placeholder="sk-..."
                  className="flex-1"
                  defaultValue="sk-xxxxxxxxxxxxxxxxxxxx"
                />
                <Button variant="secondary" size="sm">Update</Button>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Alert Channels */}
      {activeTab === 'alerts' && (
        <Card>
          <h3 className="section-title mb-4">Alert Channels</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-body font-medium text-text-primary">Slack</p>
                <p className="text-micro text-text-muted">#axis-alerts</p>
              </div>
              <Badge variant="success">Connected</Badge>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-body font-medium text-text-primary">PagerDuty</p>
                <p className="text-micro text-text-muted">Not configured</p>
              </div>
              <Button variant="secondary" size="sm">Connect</Button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-body font-medium text-text-primary">Webhook</p>
                <p className="text-micro text-text-muted">Not configured</p>
              </div>
              <Button variant="secondary" size="sm">Connect</Button>
            </div>
          </div>
        </Card>
      )}

      {/* Security */}
      {activeTab === 'security' && (
        <div className="space-y-6">
          <Card>
            <h3 className="section-title mb-4">IP Allowlist</h3>
            <div className="space-y-4">
              <Input placeholder="192.168.1.0/24" />
              <Button variant="secondary" size="sm">Add IP Range</Button>
            </div>
          </Card>

          <Card className="border-danger/30">
            <h3 className="text-body font-semibold text-danger mb-4">Danger Zone</h3>
            <p className="text-body text-text-muted mb-4">
              Once you delete your organization, there is no going back. Be certain.
            </p>
            <Button variant="danger">Delete Organization</Button>
          </Card>
        </div>
      )}
    </div>
  )
}
