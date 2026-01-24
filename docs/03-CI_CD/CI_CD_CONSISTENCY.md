# CI/CD Consistency Guide

Quick reference for maintaining consistent CI/CD across all repositories.

## Repositories Using This Template

- `dm-nkp-gitops-custom-mcp-server` ✓ (Template source)
- `dm-nkp-gitops-custom-app` (Update in progress)
- Future projects (Copy this template)

## File Structure

Every repository should have:

```
.github/
├── workflows/
│   ├── ci.yaml (or ci.yml)          # CI workflow
│   ├── cd.yaml (or cd.yml)          # CD workflow
│   └── security.yaml (or security.yml)  # Security scans
└── (other files)

docs/
├── CI_CD_ARCHITECTURE.md            # Architecture documentation
├── CI_CD_TEMPLATE.md                # This template
└── (other docs)
```

## Environment Variables

### Standard Across All Repos

```yaml
env:
  REGISTRY: ghcr.io                   # Never change
  APP_NAME: your-app-name             # MUST customize
  IMAGE_NAME: ${{ github.repository_owner }}/{app-name}
```

### For CD Workflow

```yaml
env:
  CATALOG_REPO: your-catalog-repo     # MUST customize if using
  CATALOG_OWNER: ${{ github.repository_owner }}
  CATALOG_UPDATE_METHOD: pr           # or 'push'
```

### For Security Workflow

```yaml
env:
  REGISTRY: ghcr.io
  IMAGE_NAME: (from CI/CD workflow)
```

## Triggers

### CI Workflow
```yaml
on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]
```

### CD Workflow (Sequential: After CI)
```yaml
on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]
    branches: [master, main]
  push:
    tags:
      - 'v*'
```

### Security Workflow
```yaml
on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
```

## Version Numbering

All repos must follow the same version strategy:

| Trigger | Format | Example | Use Case |
|---------|--------|---------|----------|
| Tag | MAJOR.MINOR.PATCH | 1.0.0 | Production releases |
| Branch | 0.0.0-{branch}-{sha} | 0.0.0-master-abc1234 | Development builds |

**Why?**
- Semver compliance (Helm requirement)
- Prevents version conflicts
- Tracks commit SHA in version

## Cosign Installation

All repos must use the same Cosign version:

```yaml
- uses: sigstore/cosign-installer@v3.10.1  # Latest stable v3.x
```

**Current version:** v3.10.1 (installs Cosign v2.6.1)

Note: v4.x is for Cosign v3.x, but v3.10.1 is more stable.

## CVE Scanners

All repos should have all 7 scanners in `security.yaml`:

1. **CodeQL** - SAST (GitHub native)
2. **Trivy FS** - Filesystem scanning
3. **Dependency-Check** - OWASP vulnerabilities
4. **Gosec** - Go-specific issues
5. **Grype** - Comprehensive scanning
6. **Container Scan** - Image vulnerabilities
7. **License Check** - Compliance

## Customization Checklist

When setting up a new repository:

- [ ] Copy `.github/workflows/` from template
- [ ] Update `APP_NAME` in all workflows
- [ ] Update `IMAGE_NAME` in all workflows
- [ ] Verify Go version matches `go.mod`
- [ ] Check `Dockerfile` exists
- [ ] Check `chart/{app-name}/Chart.yaml` exists
- [ ] Add `CATALOG_REPO_TOKEN` secret (if using catalog updates)
- [ ] Run first workflow and verify success
- [ ] Check Security tab for scan results

## Differences from dm-nkp-gitops-custom-app

`dm-nkp-gitops-custom-app` uses:
- Buildpacks instead of Dockerfile
- Different versioning with `+` for charts
- Direct `push` trigger (not `workflow_run`)

To align with template:
1. Update CD trigger to use `workflow_run`
2. Align versioning with `-` format
3. Add unified `security.yaml`

## Common Issues & Solutions

### Issue: Version mismatch between image and chart

**Solution:** Use same format for both:
- Image: `1.0.0-sha-abc1234` (use `-`)
- Chart: `1.0.0-sha-abc1234` (use same format)

### Issue: CD not triggering after CI

**Solution:** Verify workflow trigger:
```yaml
workflow_run:
  workflows: ["CI"]  # Match CI workflow name exactly
  types: [completed]
```

### Issue: Cosign TUF errors

**Solution:** Ensure:
1. Using `cosign-installer@v3.10.1`
2. TUF cache cleared before signing
3. `SIGSTORE_TUF_ROOT` environment variable set

## Adding New Scanners

Edit `security.yaml` and add new job:

```yaml
  your-scanner:
    name: Your Scanner Name
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # Add scanner steps
```

Then add to `security-summary` job needs:
```yaml
needs: [codeql, trivy-fs, ..., your-scanner]
```

## Verification

After setting up a new repo, verify:

1. **CI runs on PR**
   - [ ] Push to feature branch
   - [ ] Create PR
   - [ ] Verify CI workflow runs

2. **CD runs after CI**
   - [ ] Push to master
   - [ ] Verify CI completes
   - [ ] Verify CD starts automatically
   - [ ] Verify images/charts pushed to GHCR

3. **Security scans**
   - [ ] Push to master
   - [ ] Go to Security tab → Code scanning
   - [ ] Verify all 7 scanners have results

## Template Update Process

When updating the template:

1. Update in `dm-nkp-gitops-custom-mcp-server` (source)
2. Document change in this file
3. Update other repos to match
4. Update version/date in `CI_CD_TEMPLATE.md`

## Resources

- [CI/CD Architecture](./CI_CD_ARCHITECTURE.md)
- [CI/CD Template](./CI_CD_TEMPLATE.md)
- [Workflow Verification](./WORKFLOW_VERIFICATION.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
