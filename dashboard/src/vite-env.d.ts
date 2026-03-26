/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_AXIS_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
