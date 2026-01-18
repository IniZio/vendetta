---
description: Create a conventional git commit
agent: build
model: opencode/grok-code-fast-1
template: |
  Create a git commit using the Conventional Commits format.
  
  Arguments: $ARGUMENTS
  
  If no arguments provided, analyze the staged changes and suggest an appropriate commit message.
  
  Recent git status:
  !`git status --short`
  
  Commit message format: <type>(<scope>): <description>
  
  Supported types:
  - feat: A new feature
  - fix: A bug fix
  - docs: Documentation only changes
  - style: Changes that don't affect code meaning
  - refactor: Code change that neither fixes a bug nor adds a feature
  - perf: Code change that improves performance
  - test: Adding or updating tests
  - build: Changes to build system or dependencies
  - ci: Changes to CI/CD configuration
  - chore: Other changes that don't modify source or test files
  
  Generate a concise, meaningful conventional commit message that accurately describes the staged changes. Only provide the commit message, ready to be used with: git commit -m "YOUR_MESSAGE"
---
