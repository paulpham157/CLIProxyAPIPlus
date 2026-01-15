/**
 * Provider type definitions for unified provider abstraction
 */

/**
 * Supported AI provider types
 */
export type ProviderType =
  | 'github-copilot'
  | 'claude'
  | 'gemini'
  | 'gemini-cli'
  | 'gemini-vertex'
  | 'aistudio'
  | 'codex'
  | 'kiro'
  | 'qwen'
  | 'iflow'
  | 'antigravity'
  | 'openai-compatibility'
  | 'vertex-compat';

/**
 * Provider authentication status
 */
export type ProviderStatus =
  | 'active'
  | 'inactive'
  | 'disabled'
  | 'unavailable'
  | 'expired'
  | 'error';

/**
 * Provider configuration
 */
export interface ProviderConfig {
  /**
   * Unique identifier for this provider instance
   */
  id: string;

  /**
   * Provider type
   */
  provider: ProviderType;

  /**
   * Optional prefix for model namespacing
   */
  prefix?: string;

  /**
   * Optional human-readable label
   */
  label?: string;

  /**
   * Current status of the provider
   */
  status: ProviderStatus;

  /**
   * Whether the provider is intentionally disabled
   */
  disabled?: boolean;

  /**
   * Whether the provider is temporarily unavailable
   */
  unavailable?: boolean;

  /**
   * Optional proxy URL override for this provider
   */
  proxyUrl?: string;

  /**
   * Provider-specific immutable configuration attributes
   */
  attributes?: Record<string, string>;

  /**
   * Provider-specific runtime mutable metadata
   */
  metadata?: Record<string, any>;

  /**
   * API key for key-based authentication
   */
  apiKey?: string;

  /**
   * OAuth credentials for token-based authentication
   */
  oauth?: {
    email?: string;
    accessToken?: string;
    refreshToken?: string;
    expiresAt?: string | number;
  };

  /**
   * Quota and rate limiting information
   */
  quota?: {
    exceeded: boolean;
    reason?: string;
    nextRecoverAt?: string | number;
    backoffLevel?: number;
  };

  /**
   * Timestamps
   */
  createdAt?: string | number;
  updatedAt?: string | number;
  lastRefreshedAt?: string | number;
}

/**
 * Provider request payload
 */
export interface ProviderRequest {
  /**
   * Target model name
   */
  model: string;

  /**
   * Request payload (provider-specific format)
   */
  payload: any;

  /**
   * Payload format identifier
   */
  format?: string;

  /**
   * Whether to stream the response
   */
  stream?: boolean;

  /**
   * HTTP headers to forward
   */
  headers?: Record<string, string | string[]>;

  /**
   * Query string parameters
   */
  query?: Record<string, string | string[]>;

  /**
   * Optional execution metadata
   */
  metadata?: Record<string, any>;
}

/**
 * Provider response payload
 */
export interface ProviderResponse {
  /**
   * Response payload (provider-specific format)
   */
  payload: any;

  /**
   * HTTP status code
   */
  statusCode?: number;

  /**
   * Response headers
   */
  headers?: Record<string, string | string[]>;

  /**
   * Optional response metadata
   */
  metadata?: Record<string, any>;

  /**
   * Usage statistics if available
   */
  usage?: {
    promptTokens?: number;
    completionTokens?: number;
    totalTokens?: number;
  };
}

/**
 * Streaming response chunk
 */
export interface ProviderStreamChunk {
  /**
   * Chunk payload (provider-specific format)
   */
  payload: any;

  /**
   * Error if streaming failed
   */
  error?: Error;

  /**
   * Whether this is the final chunk
   */
  done?: boolean;
}

/**
 * Provider error
 */
export interface ProviderError extends Error {
  /**
   * Error code
   */
  code?: string;

  /**
   * HTTP status code if applicable
   */
  statusCode?: number;

  /**
   * Whether the error is retryable
   */
  retryable?: boolean;

  /**
   * Provider-specific error details
   */
  details?: Record<string, any>;
}

/**
 * Provider interface defining the contract for all providers
 */
export interface IProvider {
  /**
   * Get provider identifier
   */
  identifier(): ProviderType;

  /**
   * Execute a non-streaming request
   */
  execute(
    config: ProviderConfig,
    request: ProviderRequest
  ): Promise<ProviderResponse>;

  /**
   * Execute a streaming request
   */
  executeStream(
    config: ProviderConfig,
    request: ProviderRequest
  ): AsyncIterable<ProviderStreamChunk>;

  /**
   * Refresh provider authentication
   */
  refresh(config: ProviderConfig): Promise<ProviderConfig>;

  /**
   * Validate provider configuration
   */
  validate(config: ProviderConfig): Promise<boolean>;

  /**
   * Count tokens for a request (if supported)
   */
  countTokens?(
    config: ProviderConfig,
    request: ProviderRequest
  ): Promise<number>;
}

/**
 * Provider factory function type
 */
export type ProviderFactory = (config: ProviderConfig) => IProvider;

/**
 * Provider registry for managing multiple providers
 */
export interface IProviderRegistry {
  /**
   * Register a provider factory
   */
  register(type: ProviderType, factory: ProviderFactory): void;

  /**
   * Get a provider instance
   */
  get(type: ProviderType, config: ProviderConfig): IProvider | undefined;

  /**
   * List all registered provider types
   */
  list(): ProviderType[];

  /**
   * Check if a provider type is registered
   */
  has(type: ProviderType): boolean;
}
