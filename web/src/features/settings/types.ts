export interface AISettings {
  provider: string;
  apiKeyMasked: string;
  hasApiKey: boolean;
  model: string;
  ollamaBaseURL: string;
  customBaseURL: string;
  isConfigured: boolean;
}

export interface UpdateAISettingsPayload {
  provider?: string;
  apiKey?: string;
  model?: string;
  ollamaBaseURL?: string;
  customBaseURL?: string;
}

export interface ProviderInfo {
  id: string;
  name: string;
  defaultModel: string;
  models: string[];
  needsApiKey: boolean;
  needsBaseUrl: boolean;
}
