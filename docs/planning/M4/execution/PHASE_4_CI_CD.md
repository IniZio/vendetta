# Phase 4: CI/CD & Release Pipeline

**Purpose**: Automated testing, building, and releasing  
**Priority**: Critical (blocks production deployment)  
**Effort**: 30 hours

---

## CI/CD Pipeline Overview

```
Push to main
    ↓
GitHub Actions Trigger
    ├─ Lint (golangci-lint, go fmt)
    ├─ Test (unit + integration)
    ├─ Coverage (>90% on new code)
    └─ Build (multi-platform binaries)
    ├─ All Pass?
    │  YES → Create Release
    │  NO → Fail + Notify
    ├─ Release Artifacts
    │  ├─ nexus-linux-amd64
    │  ├─ nexus-linux-arm64
    │  ├─ nexus-darwin-amd64
    │  └─ nexus-darwin-arm64
    └─ Upload to GitHub Releases
```

---

## GitHub Actions Workflow

**File**: `.github/workflows/ci.yml`

```yaml
name: CI/CD

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run gofmt
        run: |
          if gofmt -l . | grep -q .; then
            echo "Code formatting issues found. Run 'make fmt' to fix."
            gofmt -l .
            exit 1
          fi
      
      - name: Run go vet
        run: go vet ./...
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout 5m

  test:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y lxc docker.io
      
      - name: Run unit tests
        run: make test-unit
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: true
          minimum_coverage_percentage: 80

  build:
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Build multi-platform binaries
        run: make ci-build
      
      - name: Create checksums
        run: |
          cd dist
          sha256sum * > CHECKSUMS.txt
          cat CHECKSUMS.txt
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: dist/
          retention-days: 30

  release:
    runs-on: ubuntu-latest
    needs: build
    if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3
      
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: binaries
          path: dist/
      
      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          draft: false
          prerelease: ${{ contains(github.ref, 'rc') || contains(github.ref, 'beta') }}
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Release Process

### Version Numbering

Use Semantic Versioning: `MAJOR.MINOR.PATCH`

- **MAJOR** (v2.0.0): Breaking API changes
- **MINOR** (v1.1.0): New features (backward compatible)
- **PATCH** (v1.0.1): Bug fixes

### Release Steps

1. **Ensure main is stable**
   ```bash
   git status  # Clean working tree
   make ci-check  # All tests/lint pass
   ```

2. **Create release tag**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0: M4 complete"
   git push origin v1.0.0
   ```

3. **Wait for CI/CD pipeline**
   - Check: https://github.com/nexus/nexus/actions
   - Verify all jobs pass

4. **Verify release**
   - Download binaries from GitHub Releases
   - Test on target platform
   - Verify checksums

5. **Announce release**
   ```
   nexus v1.0.0 Released! 🎉
   
   Features:
   - One-liner install script
   - GitHub authentication
   - SSH key management
   - Workspace management
   - Load tested up to 10+ concurrent workspaces
   
   Install: curl -fsSL https://nexus.example.com/install.sh | bash
   ```

### Hotfix Process

For critical bugs:

```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-issue

# Fix and test
make test

# Commit and push
git add .
git commit -m "fix: critical issue description"
git push origin hotfix/critical-issue

# Create PR (fast-track review)
gh pr create --title "HOTFIX: critical issue" --body "Fixes #123"

# After merge, create patch release
git tag -a v1.0.1 -m "Hotfix: critical issue"
git push origin v1.0.1
```

---

## Artifact Management

### Binary Distribution

**GitHub Releases**
- All binaries hosted on: `https://github.com/nexus/nexus/releases`
- Automatic checksums
- Release notes auto-generated from commits

**Install Script**
- Reference latest release: `https://github.com/nexus/nexus/releases/latest/download/nexus-linux-amd64`
- One-liner: `curl -fsSL https://nexus.example.com/install.sh | bash`

### Archive Old Releases

After 6 months:
- Archive binaries to S3/storage
- Keep recent 3 versions on GitHub Releases
- Document in release notes

---

## Code Coverage Requirements

### Coverage Thresholds

| Component | Minimum | Target |
|-----------|---------|--------|
| Core logic | 80% | 95% |
| New code (Phase 4) | 90% | 100% |
| Handlers | 85% | 95% |
| Config parsing | 85% | 95% |
| Error handling | 90% | 100% |
| **Overall** | **85%** | **90%** |

### Coverage Check

```bash
make test-coverage

# View HTML report
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

---

## Performance Benchmarking

### Benchmark Suite

**File**: `e2e/load_test.go`

```bash
go test -bench=. -benchmem ./e2e
```

**Targets**:
- Workspace creation: <500ms
- Workspace listing: <100ms
- SSH connection: <50ms

### Performance Regression Detection

If benchmark regresses >10%:
1. Investigation required
2. Optimization needed before merge
3. Document changes

### Sample Output

```
BenchmarkWorkspaceCreation-8       100     10485575 ns/op    5.2 MB/op
BenchmarkWorkspaceStatusQuery-8   5000      245123 ns/op    2.1 MB/op
```

---

## Dependency Management

### Go Modules

Keep dependencies current:

```bash
# Check for updates
go list -u -m all

# Update all
go get -u ./...

# Tidy unused
go mod tidy

# Verify
go mod verify
```

### Security Scanning

```bash
# Install security scanners
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/aquasecurity/trivy/cmd/trivy@latest

# Run scans
make security-scan

# Review findings
# Fix any HIGH/CRITICAL issues before release
```

---

## Testing Requirements

### Pre-Release Checklist

- [ ] Unit tests: 100% pass
- [ ] Integration tests: 100% pass
- [ ] Load tests: 10+ concurrent workspaces
- [ ] Coverage: >90% on new code
- [ ] Lint: 0 issues
- [ ] Format: `make fmt` clean
- [ ] No security issues (gosec/trivy)
- [ ] Performance benchmarks: No regression
- [ ] Documentation: Updated
- [ ] Changelog: Updated

### Manual Testing (Before Release)

```bash
# 1. Test installation
curl -fsSL https://nexus.example.com/install.sh | bash

# 2. Test auth flow
nexus auth github
nexus auth status

# 3. Test SSH setup
nexus ssh setup

# 4. Create test workspace
nexus workspace create oursky/epson-eshop

# 5. Verify services running
nexus workspace services test-ws

# 6. Connect via SSH
ssh -p 2236 dev@localhost

# 7. Clean up
nexus workspace delete test-ws
```

---

## Monitoring & Observability

### Health Check Endpoint

```
GET /health

Response:
{
  "status": "healthy",
  "version": "v1.0.0",
  "uptime_seconds": 3600,
  "workspaces": 5,
  "providers": {
    "lxc": "available",
    "docker": "unavailable"
  }
}
```

### Metrics Exposed

- Workspace creation count (total/day)
- Workspace creation duration (avg/p95)
- SSH connection count
- API request count by endpoint
- Error rate by type
- Provider availability

---

## Rollback Plan

If release has critical issues:

```bash
# 1. Stop deploying new version
#    (disable in install script)

# 2. Revert to previous version
git revert HEAD
git tag -a v1.0.1-rollback -m "Rollback due to issue #123"
git push origin v1.0.1-rollback

# 3. Announce issue
#    Notify users to use: NEXUS_VERSION=v1.0.0 in install script

# 4. Fix and re-release
#    (as hotfix - see above)
```

---

## Release Timeline

**Monthly cadence** (end of month):

| Week | Activity |
|------|----------|
| Week 1-3 | Development & testing |
| Week 4 | Release candidate (RC) |
| Week 4+ | Final testing & fixes |
| Month-end | Final release |

**Off-cycle hotfixes**: As needed for critical issues

---

## Documentation for Release

**Update before each release:**

1. **CHANGELOG.md** - What changed
2. **README.md** - Quick start
3. **Installation guide** - Setup instructions
4. **API docs** - If breaking changes
5. **Migration guide** - If version bump

### Changelog Format

```markdown
## v1.0.0 - 2026-02-28

### Features
- One-liner install script
- GitHub authentication
- SSH key management

### Fixes
- Workspace creation race condition
- SSH port allocation bug

### Breaking Changes
- None

### Migration
- None needed
```

---

## Success Criteria (Phase 4)

- ✅ CI/CD pipeline working
- ✅ All tests passing
- ✅ Coverage >90% on new code
- ✅ Multi-platform binaries building
- ✅ GitHub Releases populated
- ✅ Install script downloads binaries
- ✅ Security scans clean
- ✅ Performance benchmarks documented

---

**Done with Phase 4!** Ready for production launch → M4 Complete
