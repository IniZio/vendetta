# COR-02: Fix Service Discovery Environment Variables

## 🚨 CRITICAL ISSUE

Service discovery environment variables (e.g., `vendetta_SERVICE_WEB_URL`, `vendetta_SERVICE_API_URL`) are not available in running containers, breaking a core advertised feature.

## Root Cause

In `pkg/provider/docker/docker.go`, the Docker provider collects service URL environment variables during container creation but does not set them as persistent container environment variables. The variables are only passed to individual exec calls (like setup hooks) but not persisted in the container environment.

## Impact

Users cannot access service URLs through environment variables in their development sessions, making the service discovery feature completely non-functional.

## Evidence

- **E2E Test Failure**: `TestvendettaServiceDiscovery` fails because expected environment variables are missing
- **Code Location**: `pkg/provider/docker/docker.go:104-113`
- **Test Results**: Environment variables collected but never applied to container config

## Fix Required

**File**: `pkg/provider/docker/docker.go`

**Current Code** (lines ~104-113):
```go
env := []string{}
for name, svc := range cfg.Services {
    if svc.Port > 0 {
        pStr := fmt.Sprintf("%d/tcp", svc.Port)
        if bindings, ok := json.NetworkSettings.Ports[nat.Port(pStr)]; ok && len(bindings) > 0 {
            url := fmt.Sprintf("http://localhost:%s", bindings[0].HostPort)
            env = append(env, fmt.Sprintf("vendetta_SERVICE_%s_URL=%s", name, url))
        }
    }
}
```

**Problem**: The `env` slice is created but never used in the container configuration.

**Fix**: Modify the container creation to include the environment variables:

```go
resp, err := p.cli.ContainerCreate(ctx, &container.Config{
    Image: imgName,
    Tty:   true,
    Labels: map[string]string{
        "vendetta.session.id": sessionID,
    },
    Cmd:          []string{"/bin/bash"},
    Env:          env,  // ← ADD THIS LINE
    ExposedPorts: exposedPorts,
}, &container.HostConfig{
    Mounts:       mounts,
    PortBindings: portBindings,
    Privileged:   cfg.Docker.DinD,
}, nil, nil, sessionID)
```

## Testing

After fix:
1. Run `TestvendettaServiceDiscovery` - should pass
2. Manual verification: Create session with services, exec into container, check `env | grep vendetta_SERVICE`

## Priority
🚨 **CRITICAL** - Breaks core advertised functionality

## Dependencies
None - isolated fix to Docker provider

## Risk Assessment
- **Risk**: Low - adding environment variables to container config is standard Docker practice
- **Testing**: Well covered by existing E2E test
- **Rollback**: Easy - remove the Env field addition
