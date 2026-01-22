# CI/CD Best Practices

This document captures common CI/CD patterns and best practices learned from production projects, particularly from the reference implementation in `dm-nkp-gitops-custom-app`. These practices should be applied consistently across all projects.

## Table of Contents

1. [Image Signing with Cosign](#image-signing-with-cosign)
2. [Artifact Management](#artifact-management)
3. [Workflow Structure](#workflow-structure)
4. [Security Best Practices](#security-best-practices)
5. [Performance Optimization](#performance-optimization)
6. [Common Patterns](#common-patterns)

---

## Image Signing with Cosign

### Why Sign Images?

- **Supply Chain Security**: Verify image authenticity and integrity
- **Compliance**: Meet security requirements for production deployments
- **Trust**: Ensure images haven't been tampered with

### Keyless Signing (Recommended)

Use **keyless signing** with OIDC for public repositories. No key management required!

```yaml
permissions:
  id-token: write  # Required for keyless signing
  contents: read
  packages: write

steps:
  - name: Install Cosign
    uses: sigstore/cosign-installer@v3.0.2
  
  - name: Build and push image
    id: build
    uses: docker/build-push-action@v5
    with:
      push: true
      tags: ghcr.io/owner/image:tag
    outputs:
      digest: ${{ steps.build.outputs.digest }}
  
  - name: Sign container image
    if: steps.build.outputs.digest != ''
    env:
      IMAGE_TAG: ghcr.io/owner/image@${{ steps.build.outputs.digest }}
    run: |
      cosign sign --yes ${IMAGE_TAG}
  
  - name: Verify signature
    env:
      IMAGE_TAG: ghcr.io/owner/image@${{ steps.build.outputs.digest }}
    run: |
      cosign verify ${IMAGE_TAG} \
        --certificate-identity-regexp ".*" \
        --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

### Key-Based Signing (Alternative)

For private repos or when you need more control:

```yaml
- name: Sign with key-based signing
  env:
    COSIGN_PRIVATE_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
    COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}
  run: |
    cosign sign --key env://COSIGN_PRIVATE_KEY \
      ghcr.io/owner/image@${{ steps.build.outputs.digest }}
```

**Setup:**
```bash
# Generate key pair
cosign generate-key-pair

# Store in GitHub Secrets:
# COSIGN_PRIVATE_KEY: Contents of cosign.key
# COSIGN_PASSWORD: Password used during key generation
```

### Signing Helm Charts

```yaml
- name: Sign Helm chart
  env:
    CHART_PATH: chart.tgz
  run: |
    cosign sign-blob --yes ${CHART_PATH} \
      --output-signature ${CHART_PATH}.sig \
      --output-certificate ${CHART_PATH}.pem
```

---

## Artifact Management

### Build Artifacts on Every Commit

**Pattern**: Build and sign artifacts on every push to main/master, not just tags.

```yaml
on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]

jobs:
  docker:
    steps:
      - name: Set image tag
        id: image-tag
        run: |
          if [ "${{ github.event_name }}" = "pull_request" ]; then
            echo "tag=pr-${{ github.event.pull_request.number }}" >> $GITHUB_OUTPUT
          else
            SHORT_SHA=$(echo ${{ github.sha }} | cut -c1-7)
            echo "tag=${{ github.ref_name }}-${SHORT_SHA}" >> $GITHUB_OUTPUT
          fi
      
      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          push: ${{ github.event_name != 'pull_request' }}
          tags: |
            ghcr.io/owner/image:${{ steps.image-tag.outputs.tag }}
            ghcr.io/owner/image:${{ github.sha }}
```

### Tag Strategy

| Event | Tag Pattern | Example |
|-------|------------|---------|
| **PR** | `pr-{number}` | `pr-123` |
| **Branch push** | `{branch}-{short-sha}` | `master-a1b2c3d` |
| **Tag push** | `v{version}` | `v1.2.3` |
| **Tag push** | `latest` | `latest` |

### Artifact Retention

```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    name: build-artifacts
    path: dist/
    retention-days: 30  # PR artifacts
    # retention-days: 90  # Release artifacts
```

---

## Workflow Structure

### Standard CI Workflow

```yaml
name: CI
on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository_owner }}/app-name

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v ./...
  
  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make build
  
  docker:
    runs-on: ubuntu-latest
    needs: build
    permissions:
      contents: read
      packages: write
      id-token: write
    steps:
      # ... build, push, sign
```

### Standard CD Workflow

```yaml
name: CD
on:
  push:
    tags: ['v*']

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository_owner }}/app-name

jobs:
  docker:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write
    steps:
      - id: version
        run: echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT
      # ... build, push, sign with version tag
  
  helm:
    runs-on: ubuntu-latest
    needs: docker
    permissions:
      contents: read
      packages: write
      id-token: write
    steps:
      # ... package, push, sign chart
```

---

## Security Best Practices

### 1. Minimal Permissions

Always use the minimum required permissions:

```yaml
permissions:
  contents: read      # Read repository
  packages: write     # Push to container registry
  id-token: write     # For keyless signing
```

### 2. Use Secrets Properly

- Never hardcode secrets
- Use GitHub Secrets for sensitive data
- Use `${{ secrets.GITHUB_TOKEN }}` for registry auth (auto-provided)

### 3. Sign Everything

- Sign Docker images
- Sign Helm charts
- Sign binaries (if applicable)

### 4. Verify Signatures

Always verify signatures after signing:

```yaml
- name: Verify signature
  run: |
    cosign verify ${IMAGE_TAG} \
      --certificate-identity-regexp ".*" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

### 5. Use Distroless Images

Prefer distroless or minimal base images:

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
```

### 6. Non-Root Containers

Always run as non-root:

```dockerfile
USER 65532:65532
```

---

## Performance Optimization

### 1. Use Build Cache

Enable GitHub Actions cache for Docker builds:

```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### 2. Parallel Jobs

Run independent jobs in parallel:

```yaml
jobs:
  test:
    # ...
  build:
    # ...
  helm-lint:
    # ... (runs in parallel with test/build)
```

### 3. Conditional Steps

Skip unnecessary steps:

```yaml
- name: Sign image
  if: github.event_name != 'pull_request' && steps.build.outputs.digest != ''
  # ...
```

### 4. Matrix Builds

Use matrix strategy for multi-platform builds:

```yaml
strategy:
  matrix:
    platform: [linux/amd64, linux/arm64]
```

Or use Docker Buildx (recommended):

```yaml
platforms: linux/amd64,linux/arm64
```

---

## Common Patterns

### Pattern 1: Version Extraction

```yaml
- id: version
  run: echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT
```

### Pattern 2: Dynamic Tagging

```yaml
- name: Set image tag
  id: image-tag
  run: |
    if [ "${{ github.event_name }}" = "pull_request" ]; then
      echo "tag=pr-${{ github.event.pull_request.number }}" >> $GITHUB_OUTPUT
    else
      SHORT_SHA=$(echo ${{ github.sha }} | cut -c1-7)
      echo "tag=${{ github.ref_name }}-${SHORT_SHA}" >> $GITHUB_OUTPUT
    fi
```

### Pattern 3: Build Args

```yaml
build-args: |
  VERSION=${{ steps.version.outputs.version }}
  GIT_COMMIT=${{ github.sha }}
  BUILD_TIME=${{ github.event.head_commit.timestamp }}
```

### Pattern 4: Multi-Platform Builds

```yaml
platforms: linux/amd64,linux/arm64
```

### Pattern 5: Conditional Push

```yaml
push: ${{ github.event_name != 'pull_request' }}
```

### Pattern 6: Digest-Based Signing

Always sign by digest, not tag:

```yaml
IMAGE_TAG: ghcr.io/owner/image@${{ steps.build.outputs.digest }}
```

### Pattern 7: Helm Chart Versioning

```yaml
- name: Update Chart version
  run: |
    sed -i "s/^version:.*/version: ${{ steps.version.outputs.version }}/" chart/Chart.yaml
    sed -i "s/^appVersion:.*/appVersion: \"${{ steps.version.outputs.version }}\"/" chart/Chart.yaml
```

---

## Workflow Checklist

### For Every New Project

- [ ] Set up CI workflow with test, build, docker jobs
- [ ] Set up CD workflow for tag-based releases
- [ ] Configure image signing with Cosign (keyless)
- [ ] Set up artifact uploads
- [ ] Configure proper permissions
- [ ] Use build caching
- [ ] Set up Helm chart packaging (if applicable)
- [ ] Configure multi-platform builds (if needed)
- [ ] Add signature verification
- [ ] Document workflow in README

### For Every Workflow Update

- [ ] Verify permissions are minimal
- [ ] Ensure signing is configured
- [ ] Test on PR before merging
- [ ] Update documentation if workflow changes
- [ ] Verify artifact retention settings

---

## Reference Implementations

### Example Projects

1. **dm-nkp-gitops-custom-app**
   - Full CI/CD with signing
   - Multi-platform builds
   - Helm chart packaging
   - Artifact management

2. **dm-nkp-gitops-custom-mcp-server** (this repo)
   - CI workflow with signing on every commit
   - CD workflow for releases
   - Helm chart signing

### Key Files to Reference

- `.github/workflows/ci.yaml` - Continuous Integration
- `.github/workflows/cd.yaml` - Continuous Deployment
- `Dockerfile` - Container build configuration
- `chart/` - Helm chart structure

---

## Troubleshooting

### Issue: Signing Fails

**Error**: `Error: no matching credentials`

**Solution**: Ensure `id-token: write` permission is set:

```yaml
permissions:
  id-token: write
```

### Issue: Cannot Push to Registry

**Error**: `unauthorized: authentication required`

**Solution**: Ensure `packages: write` permission and correct login:

```yaml
permissions:
  packages: write

- uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### Issue: Build Cache Not Working

**Solution**: Ensure cache is configured:

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

---

## Additional Resources

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Buildx Documentation](https://docs.docker.com/build/buildx/)
- [Helm OCI Registry](https://helm.sh/docs/topics/registries/)

---

## Summary

**Key Takeaways:**

1. ✅ **Sign all artifacts** using Cosign (keyless for public repos)
2. ✅ **Build on every commit** to master/main, not just tags
3. ✅ **Use minimal permissions** - only what's needed
4. ✅ **Enable build caching** for faster builds
5. ✅ **Sign by digest** not tag for security
6. ✅ **Verify signatures** after signing
7. ✅ **Use multi-platform builds** when needed
8. ✅ **Upload artifacts** for traceability

These practices ensure secure, efficient, and maintainable CI/CD pipelines across all projects.
