'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, vs } from 'react-syntax-highlighter/dist/esm/styles/prism';

type Provider = 'openai' | 'claude' | 'gemini' | 'codex' | 'kiro';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  code?: CodeBlock[];
  provider?: Provider;
  model?: string;
  error?: boolean;
}

interface CodeBlock {
  language: string;
  code: string;
  filename?: string;
  startLine?: number;
}

interface ProviderConfig {
  name: string;
  endpoint: string;
  supportsStreaming: boolean;
  icon?: string;
  color?: string;
}

interface ChatSettings {
  temperature?: number;
  maxTokens?: number;
  topP?: number;
  frequencyPenalty?: number;
  presencePenalty?: number;
}

const PROVIDERS: Record<Provider, ProviderConfig> = {
  openai: {
    name: 'OpenAI',
    endpoint: '/v1/chat/completions',
    supportsStreaming: true,
    icon: '🤖',
    color: 'blue',
  },
  claude: {
    name: 'Claude',
    endpoint: '/v1/chat/completions',
    supportsStreaming: true,
    icon: '🧠',
    color: 'purple',
  },
  gemini: {
    name: 'Gemini',
    endpoint: '/v1/chat/completions',
    supportsStreaming: true,
    icon: '✨',
    color: 'green',
  },
  codex: {
    name: 'Codex',
    endpoint: '/v1/chat/completions',
    supportsStreaming: true,
    icon: '💻',
    color: 'indigo',
  },
  kiro: {
    name: 'Kiro',
    endpoint: '/v1/chat/completions',
    supportsStreaming: true,
    icon: '🚀',
    color: 'orange',
  },
};

export default function ChatInterfaceAdvanced() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [selectedProvider, setSelectedProvider] = useState<Provider>('openai');
  const [isLoading, setIsLoading] = useState(false);
  const [enableStreaming, setEnableStreaming] = useState(true);
  const [selectedModel, setSelectedModel] = useState('gpt-4');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [showSettings, setShowSettings] = useState(false);
  const [settings, setSettings] = useState<ChatSettings>({
    temperature: 0.7,
    maxTokens: 2000,
    topP: 1.0,
  });
  const [darkMode, setDarkMode] = useState(true);
  const [codePreviewMode, setCodePreviewMode] = useState<'inline' | 'modal'>('inline');
  const [selectedCodeBlock, setSelectedCodeBlock] = useState<CodeBlock | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const abortControllerRef = useRef<AbortController | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    fetchAvailableModels();
  }, [selectedProvider]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px';
    }
  }, [input]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const fetchAvailableModels = async () => {
    try {
      const response = await fetch('/v1/models');
      if (response.ok) {
        const data = await response.json();
        const models = data.data?.map((m: any) => m.id) || [];
        setAvailableModels(models);
        if (models.length > 0 && !models.includes(selectedModel)) {
          setSelectedModel(models[0]);
        }
      }
    } catch (error) {
      console.error('Failed to fetch models:', error);
    }
  };

  const extractCodeBlocks = useCallback((text: string): CodeBlock[] => {
    const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g;
    const blocks: CodeBlock[] = [];
    let match;

    while ((match = codeBlockRegex.exec(text)) !== null) {
      blocks.push({
        language: match[1] || 'text',
        code: match[2].trim(),
      });
    }

    return blocks;
  }, []);

  const handleSendMessage = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage: Message = {
      role: 'user',
      content: input,
      timestamp: new Date(),
      provider: selectedProvider,
      model: selectedModel,
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    const provider = PROVIDERS[selectedProvider];
    abortControllerRef.current = new AbortController();

    try {
      if (enableStreaming && provider.supportsStreaming) {
        await handleStreamingResponse(userMessage);
      } else {
        await handleNonStreamingResponse(userMessage);
      }
    } catch (error: any) {
      if (error.name === 'AbortError') {
        console.log('Request aborted');
      } else {
        console.error('Error sending message:', error);
        const errorMessage: Message = {
          role: 'assistant',
          content: `Error: ${error.message || 'Failed to get response'}`,
          timestamp: new Date(),
          provider: selectedProvider,
          error: true,
        };
        setMessages((prev) => [...prev, errorMessage]);
      }
    } finally {
      setIsLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleStreamingResponse = async (userMessage: Message) => {
    const provider = PROVIDERS[selectedProvider];
    const requestBody = {
      model: selectedModel,
      messages: [...messages, userMessage].map((m) => ({
        role: m.role,
        content: m.content,
      })),
      stream: true,
      ...settings,
    };

    const response = await fetch(provider.endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
      signal: abortControllerRef.current?.signal,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `HTTP ${response.status}`);
    }

    const reader = response.body?.getReader();
    const decoder = new TextDecoder();
    let accumulatedContent = '';

    const assistantMessage: Message = {
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      code: [],
      provider: selectedProvider,
      model: selectedModel,
    };

    setMessages((prev) => [...prev, assistantMessage]);

    if (reader) {
      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6);
              if (data === '[DONE]') continue;

              try {
                const parsed = JSON.parse(data);
                const delta = parseProviderResponse(parsed, selectedProvider);

                if (delta) {
                  accumulatedContent += delta;
                  setMessages((prev) => {
                    const updated = [...prev];
                    const lastMessage = updated[updated.length - 1];
                    lastMessage.content = accumulatedContent;
                    lastMessage.code = extractCodeBlocks(accumulatedContent);
                    return updated;
                  });
                }
              } catch (e) {
                console.warn('Failed to parse chunk:', e);
              }
            }
          }
        }
      } finally {
        reader.releaseLock();
      }
    }
  };

  const handleNonStreamingResponse = async (userMessage: Message) => {
    const provider = PROVIDERS[selectedProvider];
    const requestBody = {
      model: selectedModel,
      messages: [...messages, userMessage].map((m) => ({
        role: m.role,
        content: m.content,
      })),
      stream: false,
      ...settings,
    };

    const response = await fetch(provider.endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
      signal: abortControllerRef.current?.signal,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `HTTP ${response.status}`);
    }

    const data = await response.json();
    const content = parseProviderResponse(data, selectedProvider) || '';

    const assistantMessage: Message = {
      role: 'assistant',
      content,
      timestamp: new Date(),
      code: extractCodeBlocks(content),
      provider: selectedProvider,
      model: selectedModel,
    };

    setMessages((prev) => [...prev, assistantMessage]);
  };

  const parseProviderResponse = (data: any, provider: Provider): string | null => {
    if (!data) return null;

    if (data.choices && Array.isArray(data.choices) && data.choices.length > 0) {
      const choice = data.choices[0];

      if (choice.delta?.content) {
        return choice.delta.content;
      }

      if (choice.message?.content) {
        return choice.message.content;
      }
    }

    if (data.content) {
      return typeof data.content === 'string' ? data.content : JSON.stringify(data.content);
    }

    if (provider === 'gemini' && data.candidates?.[0]?.content?.parts?.[0]?.text) {
      return data.candidates[0].content.parts[0].text;
    }

    if (provider === 'claude' && data.completion) {
      return data.completion;
    }

    return null;
  };

  const handleStopGeneration = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      console.log('Copied to clipboard');
    });
  };

  const downloadCode = (code: string, filename?: string, language?: string) => {
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename || `code.${language || 'txt'}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const clearChat = () => {
    setMessages([]);
  };

  const exportChat = () => {
    const chatText = messages
      .map((m) => `[${m.timestamp.toLocaleString()}] ${m.role}: ${m.content}`)
      .join('\n\n');
    const blob = new Blob([chatText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `chat-export-${Date.now()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const renderCodeBlock = (block: CodeBlock, index: number) => {
    const syntaxStyle = darkMode ? vscDarkPlus : vs;

    return (
      <div key={index} className="relative group my-4">
        <div
          className={`flex items-center justify-between ${
            darkMode ? 'bg-gray-900' : 'bg-gray-100'
          } rounded-t-lg px-4 py-2 border-b ${
            darkMode ? 'border-gray-700' : 'border-gray-300'
          }`}
        >
          <span
            className={`text-sm font-mono ${
              darkMode ? 'text-gray-400' : 'text-gray-600'
            }`}
          >
            {block.language}
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => copyToClipboard(block.code)}
              className={`text-xs px-2 py-1 ${
                darkMode
                  ? 'bg-gray-700 hover:bg-gray-600'
                  : 'bg-gray-200 hover:bg-gray-300'
              } rounded transition-colors`}
              title="Copy code"
            >
              Copy
            </button>
            <button
              onClick={() => downloadCode(block.code, block.filename, block.language)}
              className={`text-xs px-2 py-1 ${
                darkMode
                  ? 'bg-gray-700 hover:bg-gray-600'
                  : 'bg-gray-200 hover:bg-gray-300'
              } rounded transition-colors`}
              title="Download code"
            >
              Download
            </button>
            {codePreviewMode === 'modal' && (
              <button
                onClick={() => setSelectedCodeBlock(block)}
                className={`text-xs px-2 py-1 ${
                  darkMode
                    ? 'bg-gray-700 hover:bg-gray-600'
                    : 'bg-gray-200 hover:bg-gray-300'
                } rounded transition-colors`}
                title="Open in modal"
              >
                Expand
              </button>
            )}
          </div>
        </div>
        <SyntaxHighlighter
          language={block.language}
          style={syntaxStyle}
          customStyle={{
            margin: 0,
            borderTopLeftRadius: 0,
            borderTopRightRadius: 0,
            borderBottomLeftRadius: '0.5rem',
            borderBottomRightRadius: '0.5rem',
          }}
          showLineNumbers
        >
          {block.code}
        </SyntaxHighlighter>
      </div>
    );
  };

  return (
    <div
      className={`flex flex-col h-screen ${
        darkMode ? 'bg-gray-900 text-gray-100' : 'bg-white text-gray-900'
      }`}
    >
      <header
        className={`${
          darkMode ? 'bg-gray-800 border-gray-700' : 'bg-gray-100 border-gray-300'
        } border-b p-4`}
      >
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <h1 className="text-2xl font-bold">AI Chat Interface</h1>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <label htmlFor="provider-select" className="text-sm font-medium">
                Provider:
              </label>
              <select
                id="provider-select"
                value={selectedProvider}
                onChange={(e) => setSelectedProvider(e.target.value as Provider)}
                className={`${
                  darkMode
                    ? 'bg-gray-700 border-gray-600'
                    : 'bg-white border-gray-300'
                } border rounded px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500`}
              >
                {Object.entries(PROVIDERS).map(([key, config]) => (
                  <option key={key} value={key}>
                    {config.icon} {config.name}
                  </option>
                ))}
              </select>
            </div>

            {availableModels.length > 0 && (
              <div className="flex items-center gap-2">
                <label htmlFor="model-select" className="text-sm font-medium">
                  Model:
                </label>
                <select
                  id="model-select"
                  value={selectedModel}
                  onChange={(e) => setSelectedModel(e.target.value)}
                  className={`${
                    darkMode
                      ? 'bg-gray-700 border-gray-600'
                      : 'bg-white border-gray-300'
                  } border rounded px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500`}
                >
                  {availableModels.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={enableStreaming}
                onChange={(e) => setEnableStreaming(e.target.checked)}
                className="rounded"
              />
              Streaming
            </label>

            <button
              onClick={() => setShowSettings(!showSettings)}
              className={`px-3 py-1 ${
                darkMode ? 'bg-gray-700 hover:bg-gray-600' : 'bg-gray-200 hover:bg-gray-300'
              } rounded text-sm transition-colors`}
            >
              ⚙️ Settings
            </button>

            <button
              onClick={() => setDarkMode(!darkMode)}
              className={`px-3 py-1 ${
                darkMode ? 'bg-gray-700 hover:bg-gray-600' : 'bg-gray-200 hover:bg-gray-300'
              } rounded text-sm transition-colors`}
            >
              {darkMode ? '☀️' : '🌙'}
            </button>

            <button
              onClick={exportChat}
              disabled={messages.length === 0}
              className={`px-3 py-1 ${
                darkMode ? 'bg-gray-700 hover:bg-gray-600' : 'bg-gray-200 hover:bg-gray-300'
              } rounded text-sm transition-colors disabled:opacity-50`}
            >
              📥 Export
            </button>

            <button
              onClick={clearChat}
              disabled={messages.length === 0}
              className={`px-3 py-1 ${
                darkMode ? 'bg-red-700 hover:bg-red-600' : 'bg-red-200 hover:bg-red-300'
              } rounded text-sm transition-colors disabled:opacity-50`}
            >
              🗑️ Clear
            </button>
          </div>
        </div>

        {showSettings && (
          <div
            className={`mt-4 p-4 ${
              darkMode ? 'bg-gray-700' : 'bg-gray-100'
            } rounded-lg max-w-7xl mx-auto`}
          >
            <h3 className="text-lg font-semibold mb-3">Settings</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <label className="text-sm">Temperature: {settings.temperature}</label>
                <input
                  type="range"
                  min="0"
                  max="2"
                  step="0.1"
                  value={settings.temperature}
                  onChange={(e) =>
                    setSettings({ ...settings, temperature: parseFloat(e.target.value) })
                  }
                  className="w-full"
                />
              </div>
              <div>
                <label className="text-sm">Max Tokens: {settings.maxTokens}</label>
                <input
                  type="range"
                  min="100"
                  max="4000"
                  step="100"
                  value={settings.maxTokens}
                  onChange={(e) =>
                    setSettings({ ...settings, maxTokens: parseInt(e.target.value) })
                  }
                  className="w-full"
                />
              </div>
              <div>
                <label className="text-sm">Top P: {settings.topP}</label>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.1"
                  value={settings.topP}
                  onChange={(e) =>
                    setSettings({ ...settings, topP: parseFloat(e.target.value) })
                  }
                  className="w-full"
                />
              </div>
              <div>
                <label className="text-sm flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={codePreviewMode === 'modal'}
                    onChange={(e) =>
                      setCodePreviewMode(e.target.checked ? 'modal' : 'inline')
                    }
                  />
                  Modal Code Preview
                </label>
              </div>
            </div>
          </div>
        )}
      </header>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="max-w-4xl mx-auto space-y-4">
          {messages.map((message, index) => (
            <div
              key={index}
              className={`flex ${
                message.role === 'user' ? 'justify-end' : 'justify-start'
              }`}
            >
              <div
                className={`max-w-3xl rounded-lg p-4 ${
                  message.role === 'user'
                    ? 'bg-blue-600 text-white'
                    : message.error
                    ? 'bg-red-600 text-white'
                    : darkMode
                    ? 'bg-gray-800 text-gray-100'
                    : 'bg-gray-100 text-gray-900'
                }`}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="font-semibold">
                    {message.role === 'user'
                      ? 'You'
                      : message.provider
                      ? PROVIDERS[message.provider].name
                      : 'Assistant'}
                  </span>
                  {message.model && (
                    <span className="text-xs opacity-70">({message.model})</span>
                  )}
                  <span className="text-xs opacity-70">
                    {message.timestamp.toLocaleTimeString()}
                  </span>
                </div>

                <div className="prose prose-invert max-w-none">
                  {message.code && message.code.length > 0 ? (
                    <div className="space-y-4">
                      {message.content.split(/```[\w]*\n[\s\S]*?```/).map((text, i) => (
                        <React.Fragment key={i}>
                          {text.trim() && (
                            <p className="whitespace-pre-wrap">{text.trim()}</p>
                          )}
                          {message.code && message.code[i] && renderCodeBlock(message.code[i], i)}
                        </React.Fragment>
                      ))}
                    </div>
                  ) : (
                    <p className="whitespace-pre-wrap">{message.content}</p>
                  )}
                </div>
              </div>
            </div>
          ))}
          {isLoading && (
            <div className="flex justify-start">
              <div
                className={`${
                  darkMode ? 'bg-gray-800' : 'bg-gray-100'
                } rounded-lg p-4`}
              >
                <div className="flex items-center gap-2">
                  <div className="animate-spin rounded-full h-4 w-4 border-2 border-blue-500 border-t-transparent"></div>
                  <span className={`text-sm ${darkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                    {PROVIDERS[selectedProvider].name} is thinking...
                  </span>
                </div>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      <div
        className={`${
          darkMode ? 'bg-gray-800 border-gray-700' : 'bg-gray-100 border-gray-300'
        } border-t p-4`}
      >
        <div className="max-w-4xl mx-auto">
          <div className="flex gap-2">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSendMessage();
                }
              }}
              placeholder="Type your message... (Shift+Enter for new line)"
              className={`flex-1 ${
                darkMode
                  ? 'bg-gray-700 border-gray-600'
                  : 'bg-white border-gray-300'
              } border rounded-lg px-4 py-3 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500`}
              rows={1}
              disabled={isLoading}
              style={{ maxHeight: '200px', minHeight: '60px' }}
            />
            {isLoading ? (
              <button
                onClick={handleStopGeneration}
                className="px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors"
              >
                ⏹️ Stop
              </button>
            ) : (
              <button
                onClick={handleSendMessage}
                disabled={!input.trim()}
                className="px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-colors"
              >
                ▶️ Send
              </button>
            )}
          </div>
        </div>
      </div>

      {selectedCodeBlock && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div
            className={`${
              darkMode ? 'bg-gray-800' : 'bg-white'
            } rounded-lg max-w-6xl max-h-[90vh] overflow-auto w-full`}
          >
            <div className="p-4 border-b border-gray-700 flex items-center justify-between">
              <h3 className="text-lg font-semibold">Code Preview</h3>
              <button
                onClick={() => setSelectedCodeBlock(null)}
                className="px-3 py-1 bg-gray-700 hover:bg-gray-600 rounded"
              >
                ✕ Close
              </button>
            </div>
            <div className="p-4">{renderCodeBlock(selectedCodeBlock, 0)}</div>
          </div>
        </div>
      )}
    </div>
  );
}
