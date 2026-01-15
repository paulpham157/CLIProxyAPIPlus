import React, { useState } from 'react';
import { useProviderStore, type Provider, type ProviderInfo } from '../../store/provider-store';

interface ProviderSelectorProps {
  className?: string;
}

export const ProviderSelector: React.FC<ProviderSelectorProps> = ({ className = '' }) => {
  const { selectedProvider, providers, setSelectedProvider } = useProviderStore();
  const [isOpen, setIsOpen] = useState(false);

  const selectedProviderInfo = providers[selectedProvider];
  const providerList = Object.values(providers);

  const handleSelect = (provider: Provider) => {
    setSelectedProvider(provider);
    setIsOpen(false);
  };

  const getStatusBadgeColor = (available: boolean) => {
    return available ? 'bg-green-500' : 'bg-red-500';
  };

  const getStatusText = (available: boolean) => {
    return available ? 'Available' : 'Unavailable';
  };

  return (
    <div className={`relative inline-block text-left ${className}`}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="inline-flex w-full justify-between items-center gap-x-2 rounded-md bg-white px-4 py-2.5 text-sm font-medium text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        <div className="flex items-center gap-2">
          <span>{selectedProviderInfo.name}</span>
          <span
            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium text-white ${getStatusBadgeColor(
              selectedProviderInfo.status.available
            )}`}
          >
            {getStatusText(selectedProviderInfo.status.available)}
          </span>
        </div>
        <svg
          className={`h-5 w-5 text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden="true"
        >
          <path
            fillRule="evenodd"
            d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute right-0 z-20 mt-2 w-80 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="py-1" role="menu" aria-orientation="vertical">
              {providerList.map((provider) => (
                <ProviderOption
                  key={provider.id}
                  provider={provider}
                  isSelected={provider.id === selectedProvider}
                  onSelect={() => handleSelect(provider.id)}
                />
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
};

interface ProviderOptionProps {
  provider: ProviderInfo;
  isSelected: boolean;
  onSelect: () => void;
}

const ProviderOption: React.FC<ProviderOptionProps> = ({
  provider,
  isSelected,
  onSelect,
}) => {
  const { capabilities, status } = provider;

  return (
    <button
      onClick={onSelect}
      className={`w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors ${
        isSelected ? 'bg-indigo-50' : ''
      }`}
      role="menuitem"
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="font-medium text-gray-900">{provider.name}</span>
          {isSelected && (
            <svg
              className="h-4 w-4 text-indigo-600"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
                clipRule="evenodd"
              />
            </svg>
          )}
        </div>
        <span
          className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium text-white ${
            status.available ? 'bg-green-500' : 'bg-red-500'
          }`}
        >
          {status.available ? 'Available' : 'Unavailable'}
        </span>
      </div>

      <div className="space-y-1.5">
        <div className="flex flex-wrap gap-1.5">
          {capabilities.streaming && (
            <span className="inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10">
              Streaming
            </span>
          )}
          {capabilities.codeGeneration && (
            <span className="inline-flex items-center rounded-md bg-purple-50 px-2 py-0.5 text-xs font-medium text-purple-700 ring-1 ring-inset ring-purple-700/10">
              Code Generation
            </span>
          )}
          {capabilities.multimodal && (
            <span className="inline-flex items-center rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 ring-1 ring-inset ring-amber-700/10">
              Multimodal
            </span>
          )}
        </div>
        <div className="text-xs text-gray-500">
          Context: {(capabilities.contextWindow / 1000).toFixed(0)}K tokens
          {status.latency && ` • ${status.latency}ms`}
        </div>
        {status.errorMessage && (
          <div className="text-xs text-red-600 mt-1">
            {status.errorMessage}
          </div>
        )}
      </div>
    </button>
  );
};
