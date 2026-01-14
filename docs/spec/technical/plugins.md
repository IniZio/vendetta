# Plugin System Architecture

## 1. Overview

> **Note:** This document describes the **plugin system**. For **extends** (remote template inheritance), see [Configuration Reference](../product/configuration.md).

vendetta's plugin system follows ESLint's model: load plugins as sources of capabilities, then enable/disable specific rules, skills, and commands. This provides fine-grained control while maintaining composability.

## 2. Plugin Loading & Resolution

### **Plugin Sources**
Plugins can be loaded from:
- **Remote Git repositories**: `url` field with optional `branch`
- **Local directories**: `path` field for project-specific plugins
- **Default registry**: Short names like `"vibegear/standard"` resolve to known repositories

### **Namespace Resolution**
Each plugin gets a namespace based on its `name` field. Capabilities within plugins are prefixed with this namespace:
- Plugin `"vibegear/standard"` provides `vibegear/standard/code-quality`
- Plugin `"company/templates"` provides `company/templates/security`

## 3. Automatic Enablement

When you load plugins, all their capabilities are automatically enabled. This provides a "batteries included" experience:

```yaml
# Load plugin sources - all capabilities automatically enabled
plugins:
  - name: "vibegear/standard"
    url: "https://github.com/IniZio/vendetta-config.git"
  - name: "company/internal"
    path: "./.vendetta/plugins/internal"
```

For customization, use local overrides in `.vendetta/templates/` to modify or remove specific capabilities.

## 4. Local Overrides

Local templates in `.vendetta/templates/` can override or extend plugin capabilities:

```
.vendetta/templates/
├── rules/
│   └── custom-quality.md    # Overrides vibegear/standard/code-quality
├── skills/
│   └── local-web-search.yaml # Adds local/web-search
└── commands/
    └── custom-build.yaml    # Overrides vibegear/standard/build
```

## 4. Local Overrides

Local templates in `.vendetta/templates/` can override or disable plugin capabilities:

```
.vendetta/templates/
├── rules/
│   └── custom-quality.md    # Override vibegear/standard/code-quality
├── skills/
│   └── local-web-search.yaml # Add custom web search
└── commands/
    └── custom-build.yaml    # Override vibegear/standard/build
```

To disable a capability entirely, create an empty or minimal override file that effectively removes it.

## 5. Topological DAG Resolver
To resolve plugin dependencies, we implement a **Topological Sort** using a Depth-First Search (DFS) algorithm with state tracking for cycle detection.

### **Cycle Detection Algorithm**
Each node (plugin) in the graph can be in one of three states:
1.  **UNVISITED**: Node has not been processed.
2.  **VISITING**: Node is currently in the recursion stack.
3.  **VISITED**: Node and all its dependencies have been fully processed.

**Formal Verification**: If the resolver encounters a node in the **VISITING** state, a circular dependency exists. The build must abort with a `DependencyCycleError` showing the full path (e.g., `A -> B -> C -> A`).

## 2. Parallel Fetcher (The "uv" Engine)
Parallel fetching must be high-performance but resource-aware.

### **Implementation Pattern: `errgroup` + Semaphore**
We use `golang.org/x/sync/errgroup` to manage parallel Go routines with shared context cancellation.

```go
// Implementation logic
g, ctx := errgroup.WithContext(mainCtx)
sem := make(chan struct{}, 10) // Limit to 10 concurrent git clones

for _, repo := range repos {
    repo := repo
    g.Go(func() error {
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        
        return git.Clone(ctx, repo)
    })
}
return g.Wait()
```

### **Nested Path Handling**
To handle one repository providing multiple plugins (e.g., `vendetta-config/plugins/core` and `vendetta-config/plugins/extra`):
1.  **Normalization**: Map all plugin URLs to a unique repository identifier.
2.  **Deduplication**: Only one `git clone` or `git pull` is executed per unique repository.
3.  **Symlinking/Copying**: After cloning, the specific subpaths defined in `plugin.yaml` are mapped into the workspace's plugin registry.

## 3. Deterministic Output Verification
To ensure that two different developers get the exact same environment, we implement a **Build Checksum**.

### **Hashing Strategy**
1.  **Canonicalization**: Sort all active plugins alphabetically by namespace.
2.  **Content Hashing**: Create a SHA256 hash of the "Merged Rule State":
    - Canonical JSON representation of all rules, skills, and commands.
    - Version strings of all plugins from `vendetta.lock`.
3.  **Verification**: The hash is stored in `vendetta.lock` as `metadata.content_hash`. If `vendetta workspace create` results in a different hash, the process fails with a `DeterminismWarning`.

## 4. Error Handling & Recovery
- **Network Failures**: Implement an exponential backoff (3 retries) for remote clones.
- **Lockfile Mismatch**: If `config.yaml` changes but `vendetta.lock` is not updated, the CLI must suggest running `vendetta plugin update`.
