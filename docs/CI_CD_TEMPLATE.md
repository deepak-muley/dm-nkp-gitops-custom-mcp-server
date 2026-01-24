# CI/CD Pipeline Template

Universal CI/CD template for consistency across all repositories. This template is used in:
- `dm-nkp-gitops-custom-mcp-server`
- `dm-nkp-gitops-custom-app`

## Overview

The CI/CD pipeline consists of three main workflows:

1. **CI Workflow** (`ci.yaml/yml`) - Validation and testing on every push/PR
2. **CD Workflow** (`cd.yaml/yml`) - Build, sign, and publish artifacts
3. **Security Workflow** (`security.yaml/yml`) - Vulnerability scanning and compliance

## CI Workflow

### Trigger
- On every push to `master`, `main` branches
- On every pull request to `master`, `main` branches

### Jobs
1. **test** - Code compilation and unit tests
2. **build** - Build binaries/artifacts
3. **helm** - Lint and prepare Helm charts
4. **docker** - Build and push container images (only for branches, not PRs)

### Key Features
- Go module caching for faster builds
- Graceful handling of missing tests
- Conditional image push (only on branch pushes, not PRs)
- Improved Docker image building with multi-platform support

## CD Workflow

### Trigger
- `workflow_run` trigger (runs after CI succeeds) for branch pushes
- Immediate trigger for tags (`v*`)

### Jobs
1. **docker** - Build, sign, and push Docker image
2. **helm** - Package, sign, and push Helm chart
3. **update-catalog** - Update external catalog repository (optional)

### Key Features
- Quality gate: Only runs if CI passes
- Semantic versioning for artifacts
- Docker image and Helm chart signing with Cosign
- Keyless signing using OIDC (GitHub's identity token)
- Automatic catalog updates via PR

### Version Strategy
- **Tags** (`v1.2.3`) → Use as-is
- **Branch pushes** → Prerelease format: `0.0.0-{branch}-{sha}`

Example:
- Tag: `v1.0.0` → `1.0.0` (release)
- Branch: `master-abc1234` → `0.0.0-master-abc1234` (prerelease)

## Security Workflow

### Trigger
- On every push to `master`, `main` branches
- On every pull request
- Daily at 2 AM UTC (scheduled)

### Scanners

| Scanner | Type | Coverage |
|---------|------|----------|
| CodeQL | SAST | Go source code analysis |
| Trivy FS | SCA | Filesystem + Go modules |
| Dependency-Check | SCA | OWASP dependency vulnerabilities |
| Gosec | SAST | Go-specific security issues |
| Grype | SCA | Comprehensive vulnerability detection |
| Container Scan | Image Scanning | Runtime vulnerabilities |
| License Check | Compliance | License compliance |

### Results
All results are uploaded to GitHub's Code Scanning tab for centralized viewing.

## Setup Instructions

### 1. Copy Workflows to New Repository

```bash
# From template repo
cp .github/workflows/ci.yaml /path/to/new-repo/.github/workflows/ci.yaml
cp .github/workflows/cd.yaml /path/to/new-repo/.github/workflows/cd.yaml
cp .github/workflows/security.yaml /path/to/new-repo/.github/workflows/security.yaml
```

### 2. Customize Environment Variables

Edit each workflow and update:

**In all workflows:**
```yaml
env:
  APP_NAME: your-app-name                           # Change this
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository_owner }}/your-app-name
```

**In CD workflow:**
```yaml
env:
  CATALOG_REPO: your-catalog-repo                   # Change if needed
  CATALOG_OWNER: ${{ github.repository_owner }}
  CATALOG_UPDATE_METHOD: pr                         # or 'push'
```

### 3. Create Required Secrets (Optional)

For cross-organization catalog updates:
```
CATALOG_REPO_TOKEN: <PAT with repo scope>
```

### 4. Configure Dockerfile and Chart

Ensure your repo has:
- `Dockerfile` - For container image building
- `chart/{app-name}/` - Helm chart directory with `Chart.yaml`
- `cmd/` or `src/` - Application source code

### 5. Verify Workflows

Push to a feature branch and verify:
1. CI workflow runs and passes
2. All jobs complete successfully
3. Security scans appear in Code Scanning tab

## Best Practices

### Version Management
- Use semantic versioning for tags: `v1.2.3`
- Prerelease versions for branch pushes allow multiple builds without conflicts
- Always create tags after successful CI/CD

### Image Signing
- Uses keyless signing (Cosign + OIDC)
- No secrets required for signing
- Signatures can be verified with: `cosign verify {image}`

### Security
- Run security scans on all pushes and PRs
- Fix CVEs before merging to master
- Review CodeQL alerts weekly

### Artifact Management
- Docker images tagged with version and SHA
- Helm charts use immutable versioning
- Artifacts retention: 90 days

## Troubleshooting

### CI Fails During Build
1. Check Go version matches `go.mod` spec
2. Run locally: `go build ./...`
3. Check for missing dependencies: `go mod download`

### CD Fails During Image Push
1. Verify `GITHUB_TOKEN` has `packages:write` permission
2. Check `Dockerfile` exists and builds locally
3. Review GHCR registry settings

### Security Scan Fails
1. Review CodeQL alerts in GitHub Security tab
2. Fix vulnerabilities in dependencies: `go get -u`
3. For Gosec issues, mark false positives with comments

## Customization Guide

### Add Additional Security Scanners
Edit `security.yaml` and add new job:
```yaml
  snyk-scan:
    name: Snyk Vulnerability Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: snyk/actions/golang@master
        with:
          args: --file=go.mod
```

### Add Additional Build Steps
Edit `ci.yaml` build job:
```yaml
  - name: Run integration tests
    run: make integration-tests
```

### Change Versioning Strategy
Edit `cd.yaml` version step to use custom logic.

## Files to Customize per Repository

| File | Items to Change |
|------|-----------------|
| `ci.yaml` | `APP_NAME`, `IMAGE_NAME`, `go-version` |
| `cd.yaml` | `APP_NAME`, `IMAGE_NAME`, `CATALOG_REPO`, `CATALOG_OWNER` |
| `security.yaml` | Language (if not Go), severity levels |

## Support

For issues or improvements, refer to:
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cosign Documentation](https://docs.sigstore.dev/cosign)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
