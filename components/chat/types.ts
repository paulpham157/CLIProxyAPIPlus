export interface CodePreviewProps {
  code: string;
  language: string;
  fileName?: string;
  showPreview?: boolean;
  source?: 'bolt.new' | 'v0.dev' | 'default';
}

export interface CodeBlock {
  type: 'code' | 'text';
  content: string;
  language?: string;
  fileName?: string;
}

export interface ResponseSection {
  type: 'description' | 'code';
  content: string;
  language?: string;
  fileName?: string;
}

export type SupportedLanguage = 
  | 'javascript'
  | 'typescript'
  | 'jsx'
  | 'tsx'
  | 'html'
  | 'css'
  | 'python'
  | 'go'
  | 'java'
  | 'c'
  | 'cpp'
  | 'rust'
  | 'ruby'
  | 'php'
  | 'sql'
  | 'bash'
  | 'json'
  | 'yaml'
  | 'markdown'
  | 'plaintext';

export type AISource = 'bolt.new' | 'v0.dev' | 'default';
