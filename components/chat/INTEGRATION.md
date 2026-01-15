# Chat Interface Integration Guide

This guide explains how to integrate the chat interface components with the CLI Proxy API Plus backend.

## Architecture Overview

```
┌─────────────────────┐
│  Chat Interface     │
│  (React Component)  │
└──────────┬──────────┘
           │ HTTP/SSE
           ▼
┌─────────────────────┐
│  CLI Proxy API Plus │
│  (Go Backend)       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  AI Providers       │
│  (OpenAI, Claude,   │
│   Gemini, etc.)     │
└─────────────────────┘
```

## Backend Requirements

### 1. Enable CORS

The Go backend needs to allow CORS for the frontend to make requests:

```go
// In internal/api/server.go or similar
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"}, // Add your frontend URL
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
}))
```

### 2. Verify Endpoints

Ensure these endpoints are available:

- `GET /v1/models` - Returns list of available models
- `POST /v1/chat/completions` - Chat completions (streaming and non-streaming)

### 3. Response Format

The backend should return OpenAI-compatible responses:

**Non-streaming:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }]
}
```

**Streaming:**
```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"!"}}]}

data: [DONE]
```

## Frontend Setup

### 1. Install Dependencies

```bash
npm install
# or
yarn install
# or
pnpm install
```

### 2. Configure Environment

Create `.env.local`:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

### 3. Use the Component

#### Basic Usage

```tsx
import ChatInterface from '@/components/chat/chat-interface';

export default function ChatPage() {
  return <ChatInterface />;
}
```

#### Advanced Usage

```tsx
import ChatInterfaceAdvanced from '@/components/chat/chat-interface-advanced';

export default function ChatPage() {
  return <ChatInterfaceAdvanced />;
}
```

## Customization

### Adding a New Provider

1. Update the `Provider` type:

```typescript
type Provider = 'openai' | 'claude' | 'gemini' | 'codex' | 'kiro' | 'custom';
```

2. Add provider configuration:

```typescript
const PROVIDERS: Record<Provider, ProviderConfig> = {
  // ... existing providers
  custom: {
    name: 'Custom Provider',
    endpoint: '/v1/custom/completions',
    supportsStreaming: true,
    icon: '🔧',
    color: 'teal',
  },
};
```

3. Handle provider-specific response parsing in `parseProviderResponse`:

```typescript
const parseProviderResponse = (data: any, provider: Provider): string | null => {
  // ... existing parsing logic
  
  if (provider === 'custom' && data.result?.text) {
    return data.result.text;
  }
  
  return null;
};
```

### Styling

The components use Tailwind CSS. Customize by:

1. Modifying `tailwind.config.js`
2. Updating class names in the component
3. Creating custom CSS in `chat-interface.css`

### API Base URL

If your backend runs on a different URL, configure it:

```typescript
// Option 1: Environment variable
const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

// Option 2: Direct configuration
const PROVIDERS: Record<Provider, ProviderConfig> = {
  openai: {
    name: 'OpenAI',
    endpoint: `${API_BASE}/v1/chat/completions`,
    // ...
  },
};
```

## Deployment

### Frontend Deployment

#### Next.js (Recommended)

```bash
npm run build
npm start
```

#### Static Export

```bash
npm run build
npm run export
```

Serve the `out` directory with any static host.

#### Docker

```dockerfile
FROM node:18-alpine

WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

EXPOSE 3000
CMD ["npm", "start"]
```

### Backend Configuration

Ensure the backend serves on a publicly accessible URL and update CORS:

```go
AllowOrigins: []string{"https://your-frontend-domain.com"}
```

## Testing

### Manual Testing

1. Start the backend:
```bash
go run ./cmd/server
```

2. Start the frontend:
```bash
npm run dev
```

3. Open http://localhost:3000

### Integration Tests

Create tests for the chat interface:

```typescript
// __tests__/chat-interface.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import ChatInterface from '@/components/chat/chat-interface';

test('sends message on button click', () => {
  render(<ChatInterface />);
  const input = screen.getByPlaceholderText(/type your message/i);
  const button = screen.getByText(/send/i);
  
  fireEvent.change(input, { target: { value: 'Hello' } });
  fireEvent.click(button);
  
  expect(screen.getByText('Hello')).toBeInTheDocument();
});
```

## Troubleshooting

### CORS Errors

**Problem:** Browser console shows CORS errors

**Solution:** 
- Add frontend URL to backend CORS allowlist
- Ensure preflight OPTIONS requests are handled

### Streaming Not Working

**Problem:** No real-time updates when streaming is enabled

**Solution:**
- Check that backend sends proper SSE format
- Verify `Content-Type: text/event-stream` header
- Check browser console for connection errors

### Code Blocks Not Rendering

**Problem:** Code appears as plain text

**Solution:**
- Ensure markdown code blocks use triple backticks
- Verify `react-syntax-highlighter` is installed
- Check that language identifier is supported

### Provider Not Responding

**Problem:** Selected provider doesn't return responses

**Solution:**
- Verify provider is configured in backend
- Check API keys/tokens are valid
- Review backend logs for errors
- Test endpoint directly with curl:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## Security Considerations

1. **API Keys:** Never expose API keys in frontend code
2. **Authentication:** Implement auth if deploying publicly
3. **Rate Limiting:** Add rate limits to prevent abuse
4. **Input Validation:** Sanitize user input on backend
5. **Content Security:** Use CSP headers

## Performance Optimization

1. **Code Splitting:** Lazy load syntax highlighter
```typescript
import dynamic from 'next/dynamic';
const SyntaxHighlighter = dynamic(() => import('react-syntax-highlighter'));
```

2. **Memoization:** Use React.memo for message components
```typescript
const Message = React.memo(({ message }) => {
  // ...
});
```

3. **Virtual Scrolling:** For large message histories
```typescript
import { FixedSizeList } from 'react-window';
```

## Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
- [Anthropic Claude API](https://docs.anthropic.com/claude/reference)
- [Google Gemini API](https://ai.google.dev/docs)
- [Server-Sent Events Specification](https://html.spec.whatwg.org/multipage/server-sent-events.html)
