# Provider Selector Component

A React component for selecting between different AI providers (LangGraph, Bolt.new, v0.dev) with capability badges and status indicators.

## Features

- **Provider Selection**: Dropdown to choose between available providers
- **Status Badges**: Visual indicators for provider availability
- **Capability Display**: Shows provider features (streaming, code generation, multimodal, context window)
- **Persistent Storage**: Selection is stored using Zustand with localStorage persistence
- **Responsive Design**: Tailwind CSS styling with hover states and animations

## Usage

```tsx
import { ProviderSelector } from './components/chat';

function App() {
  return (
    <div>
      <ProviderSelector className="w-64" />
    </div>
  );
}
```

## Store Access

Access the provider store directly for advanced usage:

```tsx
import { useProviderStore } from './store';

function MyComponent() {
  const { selectedProvider, providers, setSelectedProvider, updateProviderStatus } = useProviderStore();

  // Update provider status
  useEffect(() => {
    updateProviderStatus('langgraph', {
      available: true,
      latency: 150,
    });
  }, []);

  return <div>Selected: {providers[selectedProvider].name}</div>;
}
```

## Provider Types

- `langgraph`: LangGraph provider with 128K context window
- `bolt`: Bolt.new provider with 200K context window
- `v0`: v0.dev provider with 100K context window

## Capabilities

Each provider displays:
- **Streaming**: Real-time response capability
- **Code Generation**: Code generation support
- **Multimodal**: Support for images and other media
- **Context Window**: Maximum token context size
- **Latency**: Response time (if available)

## Styling

The component uses Tailwind CSS classes. Ensure Tailwind is configured in your project with the following colors:
- `gray`, `indigo`, `green`, `red`, `blue`, `purple`, `amber`
