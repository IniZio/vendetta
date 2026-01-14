#!/bin/bash
# E2E Test Setup Script
# This script sets up the environment for running comprehensive e2e tests

set -e

echo "🚀 Setting up vendetta E2E Test Environment"

# Check prerequisites
echo "📋 Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Aborting."; exit 1; }
command -v git >/dev/null 2>&1 || { echo "❌ Git is required but not installed. Aborting."; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Aborting."; exit 1; }

echo "✅ Prerequisites check passed"

# Create test directories
TEST_ROOT="/tmp/vendetta-e2e-$(date +%s)"
TEST_REPO="$TEST_ROOT/test-repo"
TEST_REMOTE="$TEST_ROOT/test-remote"

echo "📁 Creating test directories..."
mkdir -p "$TEST_REPO"
mkdir -p "$TEST_REMOTE"

# Set up test git remote repository
echo "🔗 Setting up test git remote..."
cd "$TEST_REMOTE"
git init --bare
cd "$TEST_REPO"

# Initialize test repository
echo "📦 Initializing test repository..."
git init
git config user.name "Test User"
git config user.email "test@example.com"

# Add remote
git remote add origin "$TEST_REMOTE"

# Create realistic project structure
echo "🏗️  Creating realistic project structure..."

# package.json for a Node.js project
cat > package.json << 'EOF'
{
  "name": "test-fullstack-app",
  "version": "1.0.0",
  "description": "Test fullstack application for vendetta e2e testing",
  "scripts": {
    "dev": "concurrently \"npm run server\" \"npm run client\"",
    "server": "cd server && node index.js",
    "client": "cd client && npm run dev",
    "test": "jest",
    "build": "npm run build:client && npm run build:server",
    "build:client": "cd client && npm run build",
    "build:server": "cd server && npm run build"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "concurrently": "^8.2.2",
    "jest": "^29.7.0",
    "supertest": "^6.3.3"
  }
}
EOF

# Server code
mkdir -p server
cat > server/index.js << 'EOF'
const express = require('express');
const cors = require('cors');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 5000;

app.use(cors());
app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

app.get('/api/users', (req, res) => {
  res.json([
    { id: 1, name: 'Alice', email: 'alice@example.com' },
    { id: 2, name: 'Bob', email: 'bob@example.com' }
  ]);
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running on port ${PORT}`);
});
EOF

cat > server/package.json << 'EOF'
{
  "name": "test-server",
  "version": "1.0.0",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "nodemon": "^3.0.1",
    "jest": "^29.7.0",
    "supertest": "^6.3.3"
  }
}
EOF

# Client code (simple React-like structure)
mkdir -p client/src
cat > client/package.json << 'EOF'
{
  "name": "test-client",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "vitest": "^1.0.0",
    "@vitejs/plugin-react": "^4.0.0"
  }
}
EOF

cat > client/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Test App</title>
</head>
<body>
  <div id="root">
    <h1>Test Fullstack Application</h1>
    <p>This is a test application for vendetta e2e testing.</p>
  </div>
  <script type="module" src="/src/main.jsx"></script>
</body>
</html>
EOF

cat > client/src/main.jsx << 'EOF'
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.jsx';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
EOF

cat > client/src/App.jsx << 'EOF'
import { useState, useEffect } from 'react';

function App() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/users')
      .then(res => res.json())
      .then(data => {
        setUsers(data);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch users:', err);
        setLoading(false);
      });
  }, []);

  return (
    <div>
      <h1>Test Fullstack Application</h1>
      <p>This is a test application for vendetta e2e testing.</p>

      <h2>Users</h2>
      {loading ? (
        <p>Loading...</p>
      ) : (
        <ul>
          {users.map(user => (
            <li key={user.id}>{user.name} ({user.email})</li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default App;
EOF

# Docker Compose for services
cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser -d testdb"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  db_data:
EOF

# README
cat > README.md << 'EOF'
# Test Fullstack Application

This is a comprehensive test application for vendetta e2e testing. It includes:

- **Frontend**: React + Vite client application
- **Backend**: Node.js + Express API server
- **Database**: PostgreSQL database
- **Cache**: Redis instance
- **Containerization**: Docker Compose setup

## Local Development

```bash
# Install dependencies
npm install
cd server && npm install
cd ../client && npm install

# Start all services
docker-compose up -d
npm run dev
```

## Services

- **Web Client**: http://localhost:3000
- **API Server**: http://localhost:5000
- **Database**: localhost:5432
- **Redis**: localhost:6379
EOF

# Create initial commit
echo "📝 Creating initial commit..."
git add .
git commit -m "Initial commit: Test fullstack application setup"

# Create feature branches with realistic changes
echo "🌿 Creating feature branches..."

# Feature branch 1: API improvements
git checkout -b feature/api-improvements
cat >> server/index.js << 'EOF'

app.get('/api/health', (req, res) => {
  res.json({
    status: 'healthy',
    uptime: process.uptime(),
    timestamp: new Date().toISOString()
  });
});

app.post('/api/users', (req, res) => {
  const { name, email } = req.body;
  const newUser = {
    id: Date.now(),
    name,
    email
  };
  res.status(201).json(newUser);
});
EOF
git add server/index.js
git commit -m "Add health check and user creation endpoints"

# Feature branch 2: UI enhancements
git checkout main
git checkout -b feature/ui-enhancements
cat >> client/src/App.jsx << 'EOF'

function UserForm() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email })
      });
      if (response.ok) {
        alert('User created successfully!');
        setName('');
        setEmail('');
        // Refresh user list
        window.location.reload();
      }
    } catch (err) {
      alert('Failed to create user');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        required
      />
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
      />
      <button type="submit">Add User</button>
    </form>
  );
}
EOF

# Insert the UserForm component into the App
sed -i 's/      <\/div>/      <UserForm \/>\n      <\/div>/g' client/src/App.jsx
git add client/src/App.jsx
git commit -m "Add user creation form to UI"

# Feature branch 3: Database integration
git checkout main
git checkout -b feature/database-integration
cat >> server/index.js << 'EOF'

// Database connection (mock for testing)
const { Pool } = require('pg');
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 5432,
  database: process.env.DB_NAME || 'testdb',
  user: process.env.DB_USER || 'testuser',
  password: process.env.DB_PASSWORD || 'testpass',
});

app.get('/api/db-status', async (req, res) => {
  try {
    const client = await pool.connect();
    const result = await client.query('SELECT NOW()');
    client.release();
    res.json({ status: 'connected', timestamp: result.rows[0].now });
  } catch (err) {
    res.status(500).json({ status: 'error', message: err.message });
  }
});
EOF

# Add pg dependency
cat >> server/package.json << 'EOF'
    "pg": "^8.11.3",
EOF
git add server/
git commit -m "Add PostgreSQL database integration"

# Push all branches to remote
echo "⬆️  Pushing branches to remote..."
git checkout main
git push -u origin main
git checkout feature/api-improvements
git push -u origin feature/api-improvements
git checkout feature/ui-enhancements
git push -u origin feature/ui-enhancements
git checkout feature/database-integration
git push -u origin feature/database-integration
git checkout main

echo "✅ Test repository setup complete!"
echo "📂 Test repository: $TEST_REPO"
echo "🔗 Test remote: $TEST_REMOTE"
echo "🌿 Available branches: main, feature/api-improvements, feature/ui-enhancements, feature/database-integration"

# Export environment variables for tests
cat > "$TEST_ROOT/test-env.sh" << EOF
export TEST_ROOT="$TEST_ROOT"
export TEST_REPO="$TEST_REPO"
export TEST_REMOTE="$TEST_REMOTE"
export TEST_REPO_URL="file://$TEST_REMOTE"
EOF

echo "📄 Environment variables saved to: $TEST_ROOT/test-env.sh"
echo "🔧 Source this file in your test scripts: source $TEST_ROOT/test-env.sh"
