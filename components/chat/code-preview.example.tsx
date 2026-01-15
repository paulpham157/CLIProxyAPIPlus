import React from 'react';
import { CodePreview } from './code-preview';

export function CodePreviewExamples() {
  const javascriptCode = `
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

console.log(fibonacci(10));
`;

  const htmlCode = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Hello World</title>
  <style>
    body {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      margin: 0;
      font-family: system-ui, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }
    .card {
      background: white;
      padding: 2rem;
      border-radius: 8px;
      box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    }
  </style>
</head>
<body>
  <div class="card">
    <h1>Hello, World! 👋</h1>
    <p>This is a preview from Bolt.new</p>
  </div>
</body>
</html>
`;

  const reactCode = `
function TodoApp() {
  const [todos, setTodos] = React.useState([]);
  const [input, setInput] = React.useState('');

  const addTodo = () => {
    if (input.trim()) {
      setTodos([...todos, { id: Date.now(), text: input, done: false }]);
      setInput('');
    }
  };

  const toggleTodo = (id) => {
    setTodos(todos.map(todo => 
      todo.id === id ? { ...todo, done: !todo.done } : todo
    ));
  };

  return (
    <div style={{ maxWidth: '400px', margin: '0 auto', padding: '20px' }}>
      <h1>Todo List</h1>
      <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && addTodo()}
          placeholder="Add a todo..."
          style={{ flex: 1, padding: '8px' }}
        />
        <button onClick={addTodo} style={{ padding: '8px 16px' }}>
          Add
        </button>
      </div>
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {todos.map(todo => (
          <li
            key={todo.id}
            onClick={() => toggleTodo(todo.id)}
            style={{
              padding: '12px',
              marginBottom: '8px',
              background: '#f5f5f5',
              borderRadius: '4px',
              cursor: 'pointer',
              textDecoration: todo.done ? 'line-through' : 'none',
              opacity: todo.done ? 0.6 : 1
            }}
          >
            {todo.text}
          </li>
        ))}
      </ul>
    </div>
  );
}

ReactDOM.render(<TodoApp />, document.getElementById('root'));
`;

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <h1>Code Preview Component Examples</h1>
      
      <section style={{ marginBottom: '40px' }}>
        <h2>Example 1: JavaScript Code (No Preview)</h2>
        <CodePreview 
          code={javascriptCode}
          language="javascript"
          fileName="fibonacci.js"
        />
      </section>

      <section style={{ marginBottom: '40px' }}>
        <h2>Example 2: HTML with Preview (Bolt.new Style)</h2>
        <CodePreview 
          code={htmlCode}
          language="html"
          fileName="index.html"
          showPreview={true}
          source="bolt.new"
        />
      </section>

      <section style={{ marginBottom: '40px' }}>
        <h2>Example 3: React Component with Preview (v0.dev Style)</h2>
        <CodePreview 
          code={reactCode}
          language="jsx"
          fileName="TodoApp.jsx"
          showPreview={true}
          source="v0.dev"
        />
      </section>
    </div>
  );
}

export default CodePreviewExamples;
