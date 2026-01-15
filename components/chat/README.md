# Chat Components

This directory contains React/TypeScript components for the CLI Proxy API Plus chat interface.

## Components

### Chat Interface Component

A React/TypeScript chat interface component that integrates with the CLI Proxy API Plus backend. This component provides a user-friendly interface for interacting with multiple AI providers through a unified chat interface.

#### Features

- **Multi-Provider Support**: Seamlessly switch between GitHub Copilot, Kiro (AWS CodeWhisperer), Claude, Gemini, and OpenAI
- **Provider Selection**: Dropdown to choose between available providers with capability badges and status indicators
- **Provider-Specific API Integration**: Handles provider-specific request/response formats automatically
- **Real-Time Streaming**: Supports streaming responses from all providers
- **Code Syntax Highlighting**: Automatically detects and highlights code blocks with language-specific syntax
- **Code Preview**: View generated code with syntax highlighting using VS Code Dark+ theme
- **Copy to Clipboard**: One-click copy functionality for code blocks
- **Responsive Design**: Clean, modern interface with Tailwind CSS styling, hover states and animations
- **Persistent Storage**: Provider selection stored using Zustand with localStorage persistence
- **API Key Management**: Secure API key input with show/hide toggle
- **Status Badges**: Visual indicators for provider availability
- **Capability Display**: Shows provider features (streaming, code generation, multimodal, context window)

### Code Preview Component

A fully-featured code preview component with syntax highlighting, copy-to-clipboard functionality, and live iframe preview for HTML/React/JSX components. Designed for AI chat interfaces like Bolt.new and v0.dev.

#### Features

- **Syntax Highlighting**: Uses `react-syntax-highlighter` with Prism and OneDark theme
- **Copy to Clipboard**: One-click code copying with visual feedback
- **Live Preview**: Iframe-based preview for HTML, JSX, and TSX components
- **Tab Interface**: Toggle between code view and live preview
- **Refresh Preview**: Reload the preview without losing state
- **File Name Display**: Shows the code file name in the header
- **Responsive Design**: Mobile-friendly with adaptive layouts
- **Source Tracking**: Supports marking content from 'bolt.new', 'v0.dev', or default sources

## Installation

```bash
cd components/chat
npm install
```

## Usage

### Chat Interface

```tsx
import { ChatInterface } from './components/chat/chat-interface';
import { ProviderSelector } from './components/chat';

function App() {
  return (
    <div>
      <ProviderSelector className="w-64" />
      <ChatInterface />
    </div>
  );
}
```

### Code Preview

#### Basic Code Display

```tsx
import { CodePreview } from './components/chat/code-preview';

<CodePreview
  code={`console.log('Hello, World!');`}
  language="javascript"
  fileName="example.js"
/>
```

#### HTML with Live Preview

```tsx
<CodePreview
  code={htmlString}
  language="html"
  fileName="index.html"
  showPreview={true}
  source="bolt.new"
/>
```

#### React Component with Preview

```tsx
<CodePreview
  code={reactComponentCode}
  language="jsx"
  fileName="Counter.jsx"
  showPreview={true}
  source="v0.dev"
/>
```

## Code Preview Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `code` | `string` | Yes | - | The code content to display |
| `language` | `string` | Yes | - | Programming language for syntax highlighting |
| `fileName` | `string` | No | - | Optional file name to display in header |
| `showPreview` | `boolean` | No | `false` | Enable live preview tab (only for html/jsx/tsx) |
| `source` | `'bolt.new' \| 'v0.dev' \| 'default'` | No | `'default'` | Source identifier for the code |

## Supported Providers

The chat component supports the following AI providers:

### Provider Types

- **langgraph**: LangGraph provider with 128K context window
- **bolt**: Bolt.new provider with 200K context window  
- **v0**: v0.dev provider with 100K context window
- **GitHub Copilot**: `gpt-4`, `gpt-3.5-turbo` models via `/v1/chat/completions`
- **Kiro (AWS CodeWhisperer)**: `kiro-default` model via `/v1/chat/completions`
- **Claude**: `claude-3-5-sonnet-20241022`, `claude-3-opus-20240229`, `claude-3-sonnet-20240229` via `/v1/messages`
- **Gemini**: `gemini-1.5-pro`, `gemini-1.5-flash` via `/v1beta/models`
- **OpenAI**: `gpt-4`, `gpt-4-turbo`, `gpt-3.5-turbo` via `/v1/chat/completions`

## Provider-Specific Response Handling

The component automatically handles different response formats:

### OpenAI/Copilot/Kiro Format
```json
{
  "choices": [{
    "message": {
      "content": "Response text"
    }
  }]
}
```

### Claude Format
```json
{
  "content": [{
    "type": "text",
    "text": "Response text"
  }]
}
```

### Gemini Format
```json
{
  "candidates": [{
    "content": {
      "parts": [{
        "text": "Response text"
      }]
    }
  }]
}
```

## Streaming Support

The component supports streaming responses from all providers:

- **OpenAI/Copilot/Kiro**: Server-sent events with `delta.content`
- **Claude**: Server-sent events with `content_block_delta.delta.text`
- **Gemini**: Server-sent events with `candidates[0].content.parts[].text`

## Code Block Detection

The component automatically detects code blocks in markdown format:

````markdown
```javascript
function hello() {
  console.log("Hello, world!");
}
```
````

Supported features:
- Language detection from markdown fence
- Syntax highlighting using Prism
- Copy to clipboard functionality
- VS Code Dark+ theme

## Supported Languages

The code preview component supports all languages supported by `react-syntax-highlighter`, including:
- JavaScript/TypeScript
- JSX/TSX
- HTML/CSS
- Python
- Go
- Java
- C/C++
- And many more...

## Preview Mode

Preview mode is automatically enabled when:
- `showPreview` is set to `true`
- Language is `html`, `jsx`, or `tsx`

For React components (JSX/TSX), the preview:
- Loads React 18 and ReactDOM from CDN
- Uses Babel Standalone for JSX transformation
- Assumes the component exports an `App` function
- Provides error handling with helpful messages

## Security

The iframe preview uses sandboxing with:
- `allow-scripts`: Enables JavaScript execution
- `allow-same-origin`: Allows same-origin access for iframe document manipulation

API keys are stored in component state (not persisted). Use password input type for API key entry. HTTPS recommended for production deployments.

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

## Capabilities

Each provider displays:
- **Streaming**: Real-time response capability
- **Code Generation**: Code generation support
- **Multimodal**: Support for images and other media
- **Context Window**: Maximum token context size
- **Latency**: Response time (if available)

## Component API

### Props

Currently, the chat interface component doesn't accept props but manages its own state internally.

### Internal State

- `messages`: Array of chat messages
- `input`: Current input text
- `isLoading`: Loading state during API calls
- `selectedProvider`: Currently selected AI provider
- `selectedModel`: Currently selected model
- `apiKey`: User's API key for authentication
- `showApiKeyInput`: Toggle for API key input visibility

## Keyboard Shortcuts

- `Enter`: Send message
- `Shift + Enter`: New line in message input

## Styling

The components use Tailwind CSS classes and built-in dark themes. Ensure Tailwind is configured in your project with the following colors:
- `gray`, `indigo`, `green`, `red`, `blue`, `purple`, `amber`

Dark theme inspired by VS Code:
- Dark background: `#1e1e1e`
- Accent color: `#007acc`
- Code background: `#2d2d30`
- Syntax highlighting: VS Code Dark+ theme / OneDark theme

## Development

### Building

```bash
npm run build
```

### Type Checking

```bash
npx tsc --noEmit
```

## Integration with Backend

The component expects the CLI Proxy API Plus backend to be running and accessible. Configure your backend to:

1. Enable CORS for your frontend origin
2. Accept Bearer token authentication
3. Support the provider-specific endpoints listed above

## Example Backend Configuration

```yaml
host: 0.0.0.0
port: 8080
```

## Example Integration

See `code-preview.example.tsx` for complete usage examples including:
- HTML preview with custom styling
- Interactive React counter component
- TypeScript code display

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

Modern browsers with ES6+ support, Clipboard API, and iframe sandbox support required.

## License

MIT License - See LICENSE file in the root directory
