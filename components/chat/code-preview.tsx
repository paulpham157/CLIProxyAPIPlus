import React, { useState, useRef, useEffect } from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

export interface CodePreviewProps {
  code: string;
  language: string;
  fileName?: string;
  showPreview?: boolean;
  source?: 'bolt.new' | 'v0.dev' | 'default';
}

const styles = {
  container: {
    border: '1px solid #e5e7eb',
    borderRadius: '8px',
    overflow: 'hidden',
    margin: '16px 0',
    background: '#1e1e1e',
  } as React.CSSProperties,
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '8px 12px',
    background: '#2d2d2d',
    borderBottom: '1px solid #3d3d3d',
    gap: '12px',
  } as React.CSSProperties,
  fileName: {
    fontSize: '13px',
    fontFamily: "'Menlo', 'Monaco', 'Courier New', monospace",
    color: '#9ca3af',
    flexShrink: 0,
  } as React.CSSProperties,
  tabs: {
    display: 'flex',
    gap: '4px',
    flex: 1,
    justifyContent: 'center',
  } as React.CSSProperties,
  tabButton: (active: boolean): React.CSSProperties => ({
    padding: '6px 16px',
    background: active ? '#4b5563' : 'transparent',
    border: 'none',
    color: active ? '#ffffff' : '#9ca3af',
    fontSize: '13px',
    fontWeight: 500,
    cursor: 'pointer',
    borderRadius: '4px',
    transition: 'all 0.2s',
  }),
  actions: {
    display: 'flex',
    gap: '8px',
    alignItems: 'center',
    flexShrink: 0,
  } as React.CSSProperties,
  actionButton: {
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
    padding: '6px 12px',
    background: '#3d3d3d',
    border: '1px solid #4b5563',
    borderRadius: '4px',
    color: '#e5e7eb',
    fontSize: '13px',
    cursor: 'pointer',
    transition: 'all 0.2s',
  } as React.CSSProperties,
  refreshButton: {
    padding: '6px 10px',
  } as React.CSSProperties,
  content: {
    position: 'relative',
    overflow: 'auto',
  } as React.CSSProperties,
  previewContainer: {
    background: '#ffffff',
    minHeight: '400px',
    position: 'relative',
  } as React.CSSProperties,
  previewIframe: {
    width: '100%',
    height: '500px',
    border: 'none',
    background: '#ffffff',
  } as React.CSSProperties,
};

export const CodePreview: React.FC<CodePreviewProps> = ({
  code,
  language,
  fileName,
  showPreview = false,
  source = 'default',
}) => {
  const [copied, setCopied] = useState(false);
  const [activeTab, setActiveTab] = useState<'code' | 'preview'>('code');
  const [iframeKey, setIframeKey] = useState(0);
  const iframeRef = useRef<HTMLIFrameElement>(null);

  const shouldShowPreview = showPreview && (language === 'html' || language === 'jsx' || language === 'tsx');

  useEffect(() => {
    if (activeTab === 'preview' && iframeRef.current) {
      const iframe = iframeRef.current;
      const iframeDoc = iframe.contentDocument || iframe.contentWindow?.document;

      if (iframeDoc) {
        iframeDoc.open();
        
        if (language === 'html') {
          iframeDoc.write(code);
        } else if (language === 'jsx' || language === 'tsx') {
          const previewHTML = generateReactPreview(code);
          iframeDoc.write(previewHTML);
        }
        
        iframeDoc.close();
      }
    }
  }, [activeTab, code, language, iframeKey]);

  const generateReactPreview = (reactCode: string): string => {
    return `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>React Preview</title>
  <script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
  <script src="https://unpkg.com/@babel/standalone/babel.min.js"></script>
  <style>
    body {
      margin: 0;
      padding: 16px;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
        'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
      -webkit-font-smoothing: antialiased;
      -moz-osx-font-smoothing: grayscale;
    }
    #root {
      min-height: 100vh;
    }
  </style>
</head>
<body>
  <div id="root"></div>
  <script type="text/babel">
    const { useState, useEffect, useRef } = React;
    
    ${reactCode}
    
    const rootElement = document.getElementById('root');
    const root = ReactDOM.createRoot(rootElement);
    
    try {
      root.render(<App />);
    } catch (error) {
      root.render(<div style={{ color: 'red', padding: '20px' }}>
        <h3>Preview Error:</h3>
        <pre>{error.toString()}</pre>
      </div>);
    }
  </script>
</body>
</html>
    `.trim();
  };

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy code:', err);
    }
  };

  const refreshPreview = () => {
    setIframeKey(prev => prev + 1);
  };

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        {fileName && (
          <div style={styles.fileName}>
            {fileName}
          </div>
        )}
        
        {shouldShowPreview && (
          <div style={styles.tabs}>
            <button
              style={styles.tabButton(activeTab === 'code')}
              onClick={() => setActiveTab('code')}
            >
              Code
            </button>
            <button
              style={styles.tabButton(activeTab === 'preview')}
              onClick={() => setActiveTab('preview')}
            >
              Preview
            </button>
          </div>
        )}

        <div style={styles.actions}>
          {activeTab === 'preview' && (
            <button
              style={{...styles.actionButton, ...styles.refreshButton}}
              onClick={refreshPreview}
              title="Refresh preview"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
              </svg>
            </button>
          )}
          
          {activeTab === 'code' && (
            <button
              style={styles.actionButton}
              onClick={copyToClipboard}
              title="Copy to clipboard"
            >
              {copied ? (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M20 6L9 17l-5-5"/>
                </svg>
              ) : (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                </svg>
              )}
              <span>{copied ? 'Copied!' : 'Copy'}</span>
            </button>
          )}
        </div>
      </div>

      <div style={styles.content}>
        {activeTab === 'code' ? (
          <SyntaxHighlighter
            language={language}
            style={oneDark}
            customStyle={{
              margin: 0,
              borderRadius: '0 0 8px 8px',
              fontSize: '14px',
            }}
            showLineNumbers
          >
            {code}
          </SyntaxHighlighter>
        ) : (
          <div style={styles.previewContainer}>
            <iframe
              key={iframeKey}
              ref={iframeRef}
              style={styles.previewIframe}
              sandbox="allow-scripts allow-same-origin"
              title="Code Preview"
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default CodePreview;
