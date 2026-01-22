# Catalog Update Setup Guide

Quick setup guide for the automated catalog update feature.

## Quick Start

### 1. Ensure Catalog Repository Exists

The catalog repository should be at:
- **Repository**: `{your-org}/dm-nkp-gitops-app-catalog`
- **Default branch**: `master` or `main`

### 2. Configure Catalog Structure

Choose one of these structures:

#### Option A: HelmRelease + OCIRepository (Recommended)

Create:
```
apps/dm-nkp-gitops-a2a-server/helmrelease.yaml
apps/dm-nkp-gitops-a2a-server/ocirepository.yaml
```

#### Option B: Catalog Manifest

Create:
```
catalog.yaml
# or
apps.yaml
```

#### Option C: Auto-Creation

The workflow will auto-create the structure if none exists.

### 3. Verify Workflow Configuration

The CD workflow is already configured with:

```yaml
env:
  CATALOG_REPO: dm-nkp-gitops-app-catalog
  CATALOG_OWNER: ${{ github.repository_owner }}
```

### 4. Test the Workflow

1. Push a change to master branch
2. CD workflow will:
   - Build and push image
   - Build and push Helm chart
   - **Automatically update catalog** with new version

### 5. Verify Catalog Update

Check the catalog repository:
- New commit with message: `chore: update dm-nkp-gitops-a2a-server to {version} [skip ci]`
- Version field updated in catalog files

---

## Troubleshooting

### Catalog update not running?

**Check:**
1. Helm job completed successfully
2. Catalog repo is accessible (same org or PAT configured)
3. Workflow has `contents: write` permission

### Permission denied?

**For same org:**
- Uses `GITHUB_TOKEN` automatically
- Should work out of the box

**For different org/private repo:**
1. Create Personal Access Token (PAT) with `repo` scope
2. Add as secret: `CATALOG_REPO_TOKEN`
3. Update workflow:
   ```yaml
   token: ${{ secrets.CATALOG_REPO_TOKEN }}
   ```

### Version not updating?

**Check:**
1. Catalog structure matches expected patterns
2. File paths are correct
3. YAML syntax is valid

---

## Example Catalog Files

### HelmRelease Example

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: dm-nkp-gitops-a2a-server
  namespace: gitops-agent
spec:
  chart:
    spec:
      chart: dm-nkp-gitops-a2a-server
      version: master-a1b2c3d  # ← Auto-updated
      sourceRef:
        kind: OCIRepository
        name: dm-nkp-gitops-a2a-server-chart
```

### OCIRepository Example

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: dm-nkp-gitops-a2a-server-chart
  namespace: gitops-agent
spec:
  url: oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server
  ref:
    tag: master-a1b2c3d  # ← Auto-updated
```

---

## What Gets Updated

| Trigger | Version Format | Example |
|---------|---------------|---------|
| **Tag (v*)** | `{version}` | `1.2.3` |
| **Master branch** | `{branch}-{short-sha}` | `master-a1b2c3d` |

---

## Next Steps

1. ✅ Push changes to trigger workflow
2. ✅ Monitor workflow execution
3. ✅ Verify catalog update
4. ✅ Flux CD will pick up changes automatically

For detailed design documentation, see [CATALOG_UPDATE_DESIGN.md](CATALOG_UPDATE_DESIGN.md).
