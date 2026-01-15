import React, { useState, useRef, useEffect } from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import './chat-interface.css';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  codeBlocks?: CodeBlock[];
}

interface CodeBlock {
  language: string;
  code: string;
  filename?: string;
}

interface Provider {
  id: string;
  name: string;
  endpoint: string;
  models: string[];
}

const PROVIDERS: Provider[] = [
  {
    id: 'copilot',
    name: 'GitHub Copilot',
    endpoint: '/v1/chat/completions',
    models: ['gpt-4', 'gpt-3.5-turbo']
  },
  {
    id: 'kiro',
    name: 'Kiro (AWS CodeWhisperer)',
    endpoint: '/v1/chat/completions',
    models: ['kiro-default']
  },
  {
    id: 'claude',
    name: 'Claude',
    endpoint: '/v1/messages',
    models: ['claude-3-5-sonnet-20241022', 'claude-3-opus-20240229', 'claude-3-sonnet-20240229']
  },
  {
    id: 'gemini',
    name: 'Gemini',
    endpoint: '/v1beta/models',
    models: ['gemini-1.5-pro', 'gemini-1.5-flash']
  },
  {
    id: 'openai',
    name: 'OpenAI',
    endpoint: '/v1/chat/completions',
    models: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo']
  }
];

export const ChatInterface: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<Provider>(PROVIDERS[0]);
  const [selectedModel, setSelectedModel] = useState<string>(PROVIDERS[0].models[0]);
  const [apiKey, setApiKey] = useState('');
  const [showApiKeyInput, setShowApiKeyInput] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    setSelectedModel(selectedProvider.models[0]);
  }, [selectedProvider]);

  const extractCodeBlocks = (content: string): CodeBlock[] => {
    const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g;
    const blocks: CodeBlock[] = [];
    let match;

    while ((match = codeBlockRegex.exec(content)) !== null) {
      blocks.push({
        language: match[1] || 'text',
        code: match[2].trim()
      });
    }

    return blocks;
  };

  const handleClaudeResponse = (data: any): string => {
    if (data.content && Array.isArray(data.content)) {
      return data.content
        .filter((item: any) => item.type === 'text')
        .map((item: any) => item.text)
        .join('\n');
    }
    return data.content || '';
  };

  const handleOpenAIResponse = (data: any): string => {
    if (data.choices && data.choices.length > 0) {
      return data.choices[0].message?.content || data.choices[0].text || '';
    }
    return '';
  };

  const handleGeminiResponse = (data: any): string => {
    if (data.candidates && data.candidates.length > 0) {
      const candidate = data.candidates[0];
      if (candidate.content?.parts) {
        return candidate.content.parts
          .map((part: any) => part.text)
          .join('\n');
      }
    }
    return '';
  };

  const handleStreamResponse = async (response: Response): Promise<string> => {
    const reader = response.body?.getReader();
    const decoder = new TextDecoder();
    let fullContent = '';

    if (!reader) return '';

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') continue;

            try {
              const parsed = JSON.parse(data);
              
              if (selectedProvider.id === 'claude') {
                if (parsed.type === 'content_block_delta' && parsed.delta?.text) {
                  fullContent += parsed.delta.text;
                }
              } else if (selectedProvider.id === 'gemini') {
                if (parsed.candidates?.[0]?.content?.parts) {
                  fullContent += parsed.candidates[0].content.parts
                    .map((p: any) => p.text)
                    .join('');
                }
              } else {
                if (parsed.choices?.[0]?.delta?.content) {
                  fullContent += parsed.choices[0].delta.content;
                }
              }
            } catch (e) {
              console.error('Error parsing stream chunk:', e);
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }

    return fullContent;
  };

  const sendMessage = async () => {
    if (!input.trim() || !apiKey.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      let requestBody: any;
      let endpoint = selectedProvider.endpoint;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${apiKey}`
      };

      if (selectedProvider.id === 'claude') {
        requestBody = {
          model: selectedModel,
          messages: [{ role: 'user', content: input.trim() }],
          max_tokens: 4096,
          stream: true
        };
        headers['anthropic-version'] = '2023-06-01';
      } else if (selectedProvider.id === 'gemini') {
        endpoint = `${selectedProvider.endpoint}/${selectedModel}:streamGenerateContent`;
        requestBody = {
          contents: [{
            parts: [{ text: input.trim() }]
          }]
        };
      } else {
        requestBody = {
          model: selectedModel,
          messages: [{ role: 'user', content: input.trim() }],
          stream: true
        };
      }

      const response = await fetch(endpoint, {
        method: 'POST',
        headers,
        body: JSON.stringify(requestBody)
      });

      if (!response.ok) {
        throw new Error(`API request failed: ${response.statusText}`);
      }

      let assistantContent = '';

      if (response.headers.get('content-type')?.includes('text/event-stream') || 
          requestBody.stream) {
        assistantContent = await handleStreamResponse(response);
      } else {
        const data = await response.json();
        
        if (selectedProvider.id === 'claude') {
          assistantContent = handleClaudeResponse(data);
        } else if (selectedProvider.id === 'gemini') {
          assistantContent = handleGeminiResponse(data);
        } else {
          assistantContent = handleOpenAIResponse(data);
        }
      }

      const codeBlocks = extractCodeBlocks(assistantContent);

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: assistantContent,
        timestamp: new Date(),
        codeBlocks
      };

      setMessages(prev => [...prev, assistantMessage]);
    } catch (error) {
      console.error('Error sending message:', error);
      
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `Error: ${error instanceof Error ? error.message : 'Failed to send message'}`,
        timestamp: new Date()
      };

      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const renderCodeBlock = (block: CodeBlock, index: number) => {
    return (
      <div key={index} className="code-block-container">
        <div className="code-block-header">
          <span className="code-block-language">{block.language}</span>
          <button
            className="code-block-copy-btn"
            onClick={() => copyToClipboard(block.code)}
            aria-label="Copy code"
          >
            Copy
          </button>
        </div>
        <SyntaxHighlighter
          language={block.language}
          style={vscDarkPlus}
          customStyle={{
            margin: 0,
            borderRadius: '0 0 4px 4px',
            fontSize: '14px'
          }}
        >
          {block.code}
        </SyntaxHighlighter>
      </div>
    );
  };

  const renderMessageContent = (message: Message) => {
    if (message.codeBlocks && message.codeBlocks.length > 0) {
      const parts = message.content.split(/```\w*\n[\s\S]*?```/g);
      const result: React.ReactNode[] = [];

      parts.forEach((part, index) => {
        if (part.trim()) {
          result.push(
            <div key={`text-${index}`} className="message-text">
              {part.trim()}
            </div>
          );
        }
        if (index < message.codeBlocks!.length) {
          result.push(renderCodeBlock(message.codeBlocks![index], index));
        }
      });

      return <div className="message-content-parts">{result}</div>;
    }

    return <div className="message-text">{message.content}</div>;
  };

  return (
    <div className="chat-interface">
      <div className="chat-header">
        <h1>AI Chat Interface</h1>
        
        <div className="provider-controls">
          <div className="control-group">
            <label htmlFor="provider-select">Provider:</label>
            <select
              id="provider-select"
              value={selectedProvider.id}
              onChange={(e) => {
                const provider = PROVIDERS.find(p => p.id === e.target.value);
                if (provider) setSelectedProvider(provider);
              }}
              className="provider-select"
            >
              {PROVIDERS.map(provider => (
                <option key={provider.id} value={provider.id}>
                  {provider.name}
                </option>
              ))}
            </select>
          </div>

          <div className="control-group">
            <label htmlFor="model-select">Model:</label>
            <select
              id="model-select"
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="model-select"
            >
              {selectedProvider.models.map(model => (
                <option key={model} value={model}>
                  {model}
                </option>
              ))}
            </select>
          </div>

          {showApiKeyInput && (
            <div className="control-group">
              <label htmlFor="api-key-input">API Key:</label>
              <input
                id="api-key-input"
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter API key"
                className="api-key-input"
              />
              <button
                onClick={() => setShowApiKeyInput(false)}
                className="hide-api-key-btn"
                aria-label="Hide API key input"
              >
                Hide
              </button>
            </div>
          )}

          {!showApiKeyInput && (
            <button
              onClick={() => setShowApiKeyInput(true)}
              className="show-api-key-btn"
            >
              Show API Key Input
            </button>
          )}
        </div>
      </div>

      <div className="chat-messages">
        {messages.map(message => (
          <div
            key={message.id}
            className={`chat-message ${message.role}`}
          >
            <div className="message-header">
              <span className="message-role">
                {message.role === 'user' ? 'You' : 'Assistant'}
              </span>
              <span className="message-timestamp">
                {message.timestamp.toLocaleTimeString()}
              </span>
            </div>
            {renderMessageContent(message)}
          </div>
        ))}
        {isLoading && (
          <div className="chat-message assistant loading">
            <div className="message-header">
              <span className="message-role">Assistant</span>
            </div>
            <div className="loading-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-container">
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              sendMessage();
            }
          }}
          placeholder="Type your message... (Shift+Enter for new line)"
          className="chat-input"
          rows={3}
          disabled={isLoading}
        />
        <button
          onClick={sendMessage}
          disabled={isLoading || !input.trim() || !apiKey.trim()}
          className="send-button"
        >
          {isLoading ? 'Sending...' : 'Send'}
        </button>
      </div>
    </div>
  );
};

export default ChatInterface;
