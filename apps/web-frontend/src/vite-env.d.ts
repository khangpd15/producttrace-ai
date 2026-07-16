/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_GO_CORE_API_URL: string;
  readonly VITE_NEST_AI_API_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
