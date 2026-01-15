import React from 'react';
import ReactMarkdown from 'react-markdown';
import { CodePreview } from './code-preview';
import type { Components } from 'react-markdown';

export const MarkdownWithCodePreview: React.FC<{ content: string }> = ({ content }) => {
  const components: Components = {
    code({ className, children, ...props }) {
      const match = /language-(\w+)/.exec(className || '');
      const language = match ? match[1] : '';
      const codeString = String(children).replace(/\n$/, '');

      if (language) {
        const shouldShowPreview = ['html', 'jsx', 'tsx'].includes(language);

        return (
          <CodePreview
            code={codeString}
            language={language}
            showPreview={shouldShowPreview}
          />
        );
      }

      return (
        <code className={className} {...props}>
          {children}
        </code>
      );
    },
  };

  return <ReactMarkdown components={components}>{content}</ReactMarkdown>;
};

export const BoltNewResponseRenderer: React.FC<{ response: string }> = ({ response }) => {
  const detectCodeBlocks = (text: string) => {
    const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g;
    const blocks: Array<{
      type: 'code' | 'text';
      content: string;
      language?: string;
      fileName?: string;
    }> = [];

    let lastIndex = 0;
    let match;

    while ((match = codeBlockRegex.exec(text)) !== null) {
      if (match.index > lastIndex) {
        blocks.push({
          type: 'text',
          content: text.slice(lastIndex, match.index),
        });
      }

      const language = match[1] || 'plaintext';
      const code = match[2].trim();
      
      const fileNameMatch = code.match(/^\/\/\s*(.+?)$/m) || code.match(/^<!--\s*(.+?)\s*-->$/m);
      const fileName = fileNameMatch ? fileNameMatch[1].trim() : undefined;

      blocks.push({
        type: 'code',
        content: code,
        language,
        fileName,
      });

      lastIndex = match.index + match[0].length;
    }

    if (lastIndex < text.length) {
      blocks.push({
        type: 'text',
        content: text.slice(lastIndex),
      });
    }

    return blocks;
  };

  const blocks = detectCodeBlocks(response);

  return (
    <div className="bolt-response">
      {blocks.map((block, index) => {
        if (block.type === 'code') {
          const showPreview = ['html', 'jsx', 'tsx'].includes(block.language || '');
          return (
            <CodePreview
              key={index}
              code={block.content}
              language={block.language || 'plaintext'}
              fileName={block.fileName}
              showPreview={showPreview}
              source="bolt.new"
            />
          );
        }

        return (
          <ReactMarkdown key={index}>
            {block.content}
          </ReactMarkdown>
        );
      })}
    </div>
  );
};

export const V0DevResponseRenderer: React.FC<{ response: string }> = ({ response }) => {
  const parseV0Response = (text: string) => {
    const sections: Array<{
      type: 'description' | 'code';
      content: string;
      language?: string;
      fileName?: string;
    }> = [];

    const lines = text.split('\n');
    let currentSection: typeof sections[0] | null = null;
    let buffer: string[] = [];

    const flushBuffer = () => {
      if (currentSection && buffer.length > 0) {
        currentSection.content = buffer.join('\n').trim();
        if (currentSection.content) {
          sections.push(currentSection);
        }
        buffer = [];
      }
    };

    for (const line of lines) {
      if (line.startsWith('```')) {
        flushBuffer();
        
        if (currentSection?.type === 'code') {
          currentSection = null;
        } else {
          const language = line.slice(3).trim() || 'plaintext';
          currentSection = {
            type: 'code',
            content: '',
            language,
          };
        }
      } else if (currentSection?.type === 'code' && line.startsWith('// ') && !currentSection.fileName) {
        currentSection.fileName = line.slice(3).trim();
      } else {
        if (!currentSection) {
          currentSection = {
            type: 'description',
            content: '',
          };
        }
        buffer.push(line);
      }
    }

    flushBuffer();
    return sections;
  };

  const sections = parseV0Response(response);

  const descriptionStyle: React.CSSProperties = {
    padding: '16px',
    background: '#f9fafb',
    borderRadius: '8px',
    borderLeft: '4px solid #3b82f6',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
      {sections.map((section, index) => {
        if (section.type === 'code') {
          const showPreview = ['html', 'jsx', 'tsx'].includes(section.language || '');
          return (
            <CodePreview
              key={index}
              code={section.content}
              language={section.language || 'plaintext'}
              fileName={section.fileName}
              showPreview={showPreview}
              source="v0.dev"
            />
          );
        }

        return (
          <div key={index} style={descriptionStyle}>
            <ReactMarkdown>{section.content}</ReactMarkdown>
          </div>
        );
      })}
    </div>
  );
};

export default MarkdownWithCodePreview;
