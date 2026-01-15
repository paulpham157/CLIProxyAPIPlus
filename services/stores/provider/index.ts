import { create } from 'zustand';

export interface ProviderConfig {
  apiKey?: string;
  apiUrl?: string;
  model?: string;
  temperature?: number;
  maxTokens?: number;
  [key: string]: any;
}

export interface ConnectionState {
  isConnected: boolean;
  isConnecting: boolean;
  error?: string;
  lastConnected?: Date;
}

export interface ProviderState {
  selectedProvider: string | null;
  providerConfigurations: Record<string, ProviderConfig>;
  connectionState: ConnectionState;
  setProvider: (provider: string, config?: ProviderConfig) => void;
  getProvider: () => { provider: string | null; config?: ProviderConfig };
  resetProvider: () => void;
  setConnectionState: (state: Partial<ConnectionState>) => void;
  updateProviderConfig: (provider: string, config: Partial<ProviderConfig>) => void;
}

const initialConnectionState: ConnectionState = {
  isConnected: false,
  isConnecting: false,
};

export const useProviderStore = create<ProviderState>((set, get) => ({
  selectedProvider: null,
  providerConfigurations: {},
  connectionState: initialConnectionState,

  setProvider: (provider: string, config?: ProviderConfig) => {
    set((state) => ({
      selectedProvider: provider,
      providerConfigurations: config
        ? {
            ...state.providerConfigurations,
            [provider]: config,
          }
        : state.providerConfigurations,
    }));
  },

  getProvider: () => {
    const state = get();
    return {
      provider: state.selectedProvider,
      config: state.selectedProvider
        ? state.providerConfigurations[state.selectedProvider]
        : undefined,
    };
  },

  resetProvider: () => {
    set({
      selectedProvider: null,
      providerConfigurations: {},
      connectionState: initialConnectionState,
    });
  },

  setConnectionState: (newState: Partial<ConnectionState>) => {
    set((state) => ({
      connectionState: {
        ...state.connectionState,
        ...newState,
      },
    }));
  },

  updateProviderConfig: (provider: string, config: Partial<ProviderConfig>) => {
    set((state) => ({
      providerConfigurations: {
        ...state.providerConfigurations,
        [provider]: {
          ...state.providerConfigurations[provider],
          ...config,
        },
      },
    }));
  },
}));
