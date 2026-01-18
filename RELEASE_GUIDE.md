# Release Guide - Nexus

Complete guide for creating production releases.

---

## Quick Release (5 min)

```bash
# Ensure clean working tree
git status
make ci-check  # All tests/lint pass

# Create release tag
git tag -a v1.0.0 -m "Release v1.0.0: M4 Complete

Features:
- One-liner install
- GitHub auth
- SSH key management

Fixes:
- Race condition in workspace creation

Breaking: None
Migration: None"

# Push to GitHub
git push origin v1.0.0

# Wait for CI/CD (GitHub Actions)
# Verify: https://github.com/nexus/nexus/actions

# Check release page
# https://github.com/nexus/nexus/releases/tag/v1.0.0
```

Done! Binaries uploaded automatically.

---

## Manual Release (If Needed)

```bash
# 1. Build locally
make ci-build

# 2. Verify binaries
ls -lh dist/
file dist/nexus-*

# 3. Create checksums
cd dist && sha256sum * > CHECKSUMS.txt && cat CHECKSUMS.txt

# 4. Create GitHub Release manually
gh release create v1.0.0 \
  --title "v1.0.0: M4 Complete" \
  --notes "See CHANGELOG.md for details" \
  dist/*

# 5. Verify release
gh release view v1.0.0
```

---

## Version Numbers

Use Semantic Versioning: `MAJOR.MINOR.PATCH`

Examples:
- `v1.0.0` - Initial release
- `v1.0.1` - Bug fix
- `v1.1.0` - New feature
- `v2.0.0` - Breaking change

---

## Hotfix Process

For critical bugs:

```bash
# 1. Create hotfix branch
git checkout main && git pull
git checkout -b hotfix/description

# 2. Fix and test
# ... make changes ...
make test

# 3. Commit and push
git add .
git commit -m "fix(component): description"
git push origin hotfix/description

# 4. Create PR (mark as HOTFIX)
gh pr create --title "HOTFIX: description"

# 5. After merge, tag and release
git checkout main && git pull
git tag -a v1.0.1 -m "Hotfix: description"
git push origin v1.0.1
```

---

## Pre-Release Checklist

- [ ] All tests passing: `make ci-check`
- [ ] Coverage >90%: Check codecov.io
- [ ] No linting issues: `make lint`
- [ ] Code formatted: `make fmt-check`
- [ ] CHANGELOG.md updated
- [ ] README.md updated
- [ ] Version bumped in code (if applicable)
- [ ] No security issues: `make security-scan`
- [ ] Manual testing on macOS and Linux
- [ ] Performance benchmarks stable

---

## Testing Before Release

```bash
# Full CI simulation locally
make clean
make dev-setup
make ci-check
make ci-build

# Test installation
./scripts/install.sh latest
nexus version

# Test workflows
nexus auth status
nexus workspace --help
```

---

## Release Notes Template

```markdown
## v1.0.0 - 2026-02-28

### ✨ Features
- One-liner install script: `curl -fsSL ... | bash`
- GitHub authentication: `nexus auth github`
- SSH key management: `nexus ssh setup`
- Workspace management: `nexus workspace create`
- 10+ concurrent workspace support

### 🐛 Bug Fixes
- Fixed race condition in port allocation (#45)
- Fixed SSH key fingerprint calculation (#42)

### 📖 Documentation
- Complete installation guide
- M4 execution plan
- Error handling documentation
- CI/CD pipeline setup

### 🚀 Performance
- Workspace creation: <500ms
- SSH connection: <50ms
- Load tested: 10+ concurrent workspaces

### 🔒 Security
- All API errors sanitized
- SSH keys never logged
- GitHub tokens via gh CLI (credential manager)
- No hardcoded secrets

### 📦 Artifacts
- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Checksums: CHECKSUMS.txt

### ⚠️ Breaking Changes
- None

### 🔄 Migration
- No migration needed from v0.x

### 👥 Contributors
- @alice (Lead backend)
- @bob (Systems engineer)
- @charlie (QA)

### 📝 Changelog
See [CHANGELOG.md](CHANGELOG.md) for full list
```

---

## Announcing Release

```bash
# Social media / Slack
@channel 🚀 Nexus v1.0.0 Released!

One-liner install:
curl -fsSL https://nexus.example.com/install.sh | bash

Highlights:
✅ One-liner bootstrap
✅ GitHub authentication
✅ SSH key management
✅ 10+ concurrent workspaces

Install: https://github.com/nexus/nexus/releases/tag/v1.0.0
Docs: https://nexus.dev/install
```

---

## Rollback (If Critical Issue)

```bash
# 1. Stop distributing new version
#    Remove from website, announce issue

# 2. Create rollback release
git tag -a v1.0.0-rollback -m "Rollback: Critical issue #123"
git push origin v1.0.0-rollback

# 3. Notify users
#    Install previous: NEXUS_VERSION=v0.9.0 bash install.sh

# 4. Fix and re-release (hotfix process above)
```

---

## Monitoring After Release

```bash
# Check GitHub Releases page
gh release list

# Monitor issues
gh issue list --state=open --label=bug

# Check CI health
gh workflow list
gh run list --workflow=ci.yml --status=failure

# View download stats
# GitHub Release page shows download counts per asset
```

---

## Checklist - Release Day

- [ ] CI/CD pipeline green
- [ ] All artifacts downloaded and verified
- [ ] Release notes published
- [ ] Installation tested on test machine
- [ ] Announcement posted
- [ ] Documentation updated
- [ ] Monitoring configured
- [ ] Support team notified

---

## FAQ

**Q: What if tests fail before release?**
A: Fix issues on `main` first. Tag and release only green builds.

**Q: How often should we release?**
A: Monthly cadence. Hotfixes as needed for critical issues.

**Q: Can we skip a version number?**
A: Avoid it. Use semantic versioning: skip from v1.0.0 to v1.0.2 only for patch fixes.

**Q: What about pre-release versions?**
A: Use rc/beta tags: `v1.0.0-rc.1`, `v1.0.0-beta.1`

**Q: How do users upgrade?**
A: Re-run install script or use `nexus upgrade`

---

**Ready to release?** Follow the Quick Release section above!
