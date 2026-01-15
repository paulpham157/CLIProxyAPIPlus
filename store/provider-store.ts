import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Provider = 'langgraph' | 'bolt' | 'v0';

export interface ProviderCapabilities {
  streaming: boolean;
  codeGeneration: boolean;
  contextWindow: number;
  multimodal: boolean;
}

export interface ProviderStatus {
  available: boolean;
  latency?: number;
  errorMessage?: string;
}

export interface ProviderInfo {
  id: Provider;
  name: string;
  capabilities: ProviderCapabilities;
  status: ProviderStatus;
}

interface ProviderStore {
  selectedProvider: Provider;
  providers: Record<Provider, ProviderInfo>;
  setSelectedProvider: (provider: Provider) => void;
  updateProviderStatus: (provider: Provider, status: ProviderStatus) => void;
}

const initialProviders: Record<Provider, ProviderInfo> = {
  langgraph: {
    id: 'langgraph',
    name: 'LangGraph',
    capabilities: {
      streaming: true,
      codeGeneration: true,
      contextWindow: 128000,
      multimodal: true,
    },
    status: {
      available: true,
    },
  },
  bolt: {
    id: 'bolt',
    name: 'Bolt.new',
    capabilities: {
      streaming: true,
      codeGeneration: true,
      contextWindow: 200000,
      multimodal: false,
    },
    status: {
      available: true,
    },
  },
  v0: {
    id: 'v0',
    name: 'v0.dev',
    capabilities: {
      streaming: true,
      codeGeneration: true,
      contextWindow: 100000,
      multimodal: true,
    },
    status: {
      available: true,
    },
  },
};

export const useProviderStore = create<ProviderStore>()(
  persist(
    (set) => ({
      selectedProvider: 'langgraph',
      providers: initialProviders,
      setSelectedProvider: (provider) =>
        set({ selectedProvider: provider }),
      updateProviderStatus: (provider, status) =>
        set((state) => ({
          providers: {
            ...state.providers,
            [provider]: {
              ...state.providers[provider],
              status,
            },
          },
        })),
    }),
    {
      name: 'provider-storage',
    }
  )
);
