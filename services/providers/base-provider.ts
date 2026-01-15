export interface ProviderConfig {
  apiKey?: string;
  apiUrl?: string;
  timeout?: number;
  maxRetries?: number;
  headers?: Record<string, string>;
  [key: string]: any;
}

export interface RequestOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: any;
  timeout?: number;
  signal?: AbortSignal;
}

export interface StreamChunk {
  data: any;
  raw?: string;
  done: boolean;
}

export interface ProviderResponse<T = any> {
  data: T;
  status: number;
  headers: Record<string, string>;
  raw?: any;
}

export class ProviderError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public code?: string,
    public provider?: string,
    public originalError?: any
  ) {
    super(message);
    this.name = 'ProviderError';
    Object.setPrototypeOf(this, ProviderError.prototype);
  }
}

export abstract class BaseProvider {
  protected config: ProviderConfig;
  protected initialized: boolean = false;
  protected defaultTimeout: number = 30000;
  protected defaultMaxRetries: number = 3;

  constructor(config: ProviderConfig) {
    this.config = {
      timeout: this.defaultTimeout,
      maxRetries: this.defaultMaxRetries,
      ...config,
    };
  }

  public abstract initialize(): Promise<void>;

  public abstract sendRequest<T = any>(
    endpoint: string,
    options?: RequestOptions
  ): Promise<ProviderResponse<T>>;

  public abstract streamResponse(
    endpoint: string,
    options?: RequestOptions
  ): AsyncGenerator<StreamChunk, void, unknown>;

  protected ensureInitialized(): void {
    if (!this.initialized) {
      throw new ProviderError(
        'Provider not initialized. Call initialize() first.',
        500,
        'NOT_INITIALIZED',
        this.constructor.name
      );
    }
  }

  protected buildAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {};

    if (this.config.apiKey) {
      headers['Authorization'] = `Bearer ${this.config.apiKey}`;
    }

    if (this.config.headers) {
      Object.assign(headers, this.config.headers);
    }

    return headers;
  }

  protected mergeHeaders(
    ...headerSets: (Record<string, string> | undefined)[]
  ): Record<string, string> {
    return headerSets.reduce((merged, headers) => {
      if (headers) {
        Object.assign(merged, headers);
      }
      return merged;
    }, {} as Record<string, string>);
  }

  protected async retryWithBackoff<T>(
    operation: () => Promise<T>,
    maxRetries: number = this.config.maxRetries || this.defaultMaxRetries,
    initialDelay: number = 1000
  ): Promise<T> {
    let lastError: any;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error;

        if (attempt === maxRetries) {
          break;
        }

        if (!this.isRetryableError(error)) {
          throw error;
        }

        const delay = initialDelay * Math.pow(2, attempt);
        await this.sleep(delay);
      }
    }

    throw this.wrapError(lastError, 'Max retries exceeded');
  }

  protected isRetryableError(error: any): boolean {
    if (error instanceof ProviderError) {
      const retryableStatusCodes = [408, 429, 500, 502, 503, 504];
      return error.statusCode
        ? retryableStatusCodes.includes(error.statusCode)
        : false;
    }

    if (error.code === 'ECONNRESET' || error.code === 'ETIMEDOUT') {
      return true;
    }

    return false;
  }

  protected wrapError(error: any, context?: string): ProviderError {
    if (error instanceof ProviderError) {
      return error;
    }

    const message = context
      ? `${context}: ${error.message || String(error)}`
      : error.message || String(error);

    return new ProviderError(
      message,
      error.statusCode || error.status,
      error.code,
      this.constructor.name,
      error
    );
  }

  protected handleHttpError(status: number, body: any): ProviderError {
    let message = `HTTP ${status} error`;
    let code = `HTTP_${status}`;

    if (body) {
      if (typeof body === 'string') {
        message = body;
      } else if (body.error) {
        message = typeof body.error === 'string' ? body.error : body.error.message || message;
        code = body.error.code || body.error.type || code;
      } else if (body.message) {
        message = body.message;
      }
    }

    return new ProviderError(message, status, code, this.constructor.name);
  }

  protected transformRequest(data: any): any {
    if (!data) return data;

    if (typeof data === 'object') {
      return JSON.parse(JSON.stringify(data));
    }

    return data;
  }

  protected transformResponse<T>(data: any): T {
    if (!data) return data;

    if (typeof data === 'string') {
      try {
        return JSON.parse(data);
      } catch {
        return data as T;
      }
    }

    return data;
  }

  protected validateConfig(requiredFields: string[]): void {
    const missingFields = requiredFields.filter(
      (field) => !this.config[field]
    );

    if (missingFields.length > 0) {
      throw new ProviderError(
        `Missing required configuration fields: ${missingFields.join(', ')}`,
        500,
        'INVALID_CONFIG',
        this.constructor.name
      );
    }
  }

  protected sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  protected createAbortTimeout(
    timeoutMs: number = this.config.timeout || this.defaultTimeout
  ): { signal: AbortSignal; cleanup: () => void } {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

    return {
      signal: controller.signal,
      cleanup: () => clearTimeout(timeoutId),
    };
  }

  protected sanitizeForLogging(data: any): any {
    if (!data) return data;

    const sensitiveKeys = [
      'apiKey',
      'api_key',
      'authorization',
      'token',
      'password',
      'secret',
      'credentials',
    ];

    const sanitized = JSON.parse(JSON.stringify(data));

    const redactSensitiveData = (obj: any): void => {
      if (typeof obj !== 'object' || obj === null) return;

      for (const key in obj) {
        const lowerKey = key.toLowerCase();
        if (sensitiveKeys.some((sensitive) => lowerKey.includes(sensitive))) {
          obj[key] = '[REDACTED]';
        } else if (typeof obj[key] === 'object') {
          redactSensitiveData(obj[key]);
        }
      }
    };

    redactSensitiveData(sanitized);
    return sanitized;
  }

  public getConfig(): Readonly<ProviderConfig> {
    return Object.freeze({ ...this.config });
  }

  public isInitialized(): boolean {
    return this.initialized;
  }

  protected setInitialized(value: boolean): void {
    this.initialized = value;
  }
}
