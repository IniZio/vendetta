---
title: "Conventional Commits"
description: "Enforce conventional commit message format for consistent git history"
applies_to: ["**/*"]
priority: high
enabled: true
---

# Conventional Commits

Conventional commits provide a standardized format for commit messages that makes the git history more readable and enables automated tooling.

## Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc.)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **build**: Changes that affect the build system or external dependencies
- **ci**: Changes to our CI configuration files and scripts
- **chore**: Other changes that don't modify src or test files

## Examples

```
feat: add user authentication
fix: resolve memory leak in user service
docs: update API documentation
feat(ui): add dark mode toggle
refactor: simplify user validation logic

BREAKING CHANGE: remove deprecated API endpoints
```

## Scope (optional)

Scopes provide additional context and are typically related to a specific component or feature:

```
feat(auth): implement JWT token validation
fix(api): handle null pointer exception
docs(readme): update installation instructions
```

## Breaking Changes

For commits that introduce breaking changes, add a footer:

```
feat: change API response format

BREAKING CHANGE: The response now includes additional metadata fields
```

## Why Conventional Commits?

- **Automated tooling**: Enables automatic changelog generation and version bumping
- **Consistency**: Standardized format across the team
- **Readability**: Clear intent and impact of each commit
- **Tooling integration**: Works with tools like semantic-release, commitizen, etc.

## Validation Rules

- Type must be one of the allowed types (case sensitive)
- Description must be present and start with lowercase
- Body and footer are optional but recommended for complex changes
- Lines should not exceed 72 characters (except for the first line which can be 50-72)