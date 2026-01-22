# Catalog Update Automation Design

This document describes the automated catalog update system that keeps the app catalog repository (`dm-nkp-gitops-app-catalog`) synchronized with new image versions.

## Overview

When a new Docker image and Helm chart are built and pushed in the CD workflow, the system automatically:

1. **Detects** the catalog repository structure
2. **Updates** the appropriate version reference
3. **Commits and pushes** the change to the catalog repository

This ensures that the catalog always references the latest (or specified) version of the application.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  dm-nkp-gitops-custom-mcp-server (Application Repo)            │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  CD Workflow                                              │  │
│  │  1. Build Docker image                                    │  │
│  │  2. Sign image                                            │  │
│  │  3. Build Helm chart                                      │  │
│  │  4. Push chart to OCI registry                           │  │
│  │  5. Sign chart                                            │  │
│  └───────────────────────┬───────────────────────────────────┘  │
│                          │                                       │
│                          │ Triggers                              │
│                          ▼                                       │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  update-catalog Job                                       │  │
│  │  - Checkout catalog repo                                  │  │
│  │  - Detect catalog structure                               │  │
│  │  - Update version reference                               │  │
│  │  - Commit and push                                        │  │
│  └───────────────────────┬───────────────────────────────────┘  │
└──────────────────────────┼───────────────────────────────────────┘
                           │
                           │ Git Push
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  dm-nkp-gitops-app-catalog (Catalog Repo)                      │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Updated Files:                                           │  │
│  │  - apps/dm-nkp-gitops-a2a-server/helmrelease.yaml        │  │
│  │  - apps/dm-nkp-gitops-a2a-server/ocirepository.yaml      │  │
│  │  - catalog.yaml (if using catalog manifest)              │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Flux CD picks up changes and deploys new version        │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Supported Catalog Structures

The automation supports multiple catalog organization patterns:

### 1. HelmRelease with OCIRepository (Recommended)

**Structure:**
```
apps/
  dm-nkp-gitops-a2a-server/
    helmrelease.yaml
    ocirepository.yaml
```

**HelmRelease Example:**
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
      version: 1.2.3  # ← Updated automatically
      sourceRef:
        kind: OCIRepository
        name: dm-nkp-gitops-a2a-server-chart
```

**OCIRepository Example:**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: dm-nkp-gitops-a2a-server-chart
  namespace: gitops-agent
spec:
  url: oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server
  ref:
    tag: 1.2.3  # ← Updated automatically
```

### 2. Standalone OCIRepository

**Structure:**
```
apps/
  dm-nkp-gitops-a2a-server/
    ocirepository.yaml
```

**What Gets Updated:**
- `spec.ref.tag` - Chart version
- `spec.url` - OCI registry URL (if missing)

### 3. Catalog Manifest (YAML)

**Structure:**
```
catalog.yaml
# or
apps.yaml
```

**Example:**
```yaml
apps:
  dm-nkp-gitops-a2a-server:
    version: 1.2.3  # ← Updated automatically
    chart: oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server
```

### 4. Auto-Creation

If no catalog structure is found, the automation will:

1. Create `apps/dm-nkp-gitops-a2a-server/` directory
2. Generate a HelmRelease and OCIRepository
3. Set the version to the newly built chart version

---

## Workflow Details

### Trigger Conditions

The catalog update job runs when:

- ✅ Helm chart job succeeds
- ✅ Chart is pushed to OCI registry
- ✅ Chart is signed

**Does NOT run:**
- ❌ On pull requests
- ❌ If helm job fails
- ❌ If no changes are detected

### Version Strategy

| Trigger | Version Format | Example |
|---------|---------------|---------|
| **Tag (v*)** | `{version}` | `1.2.3` |
| **Master branch** | `{branch}-{short-sha}` | `master-a1b2c3d` |

### Update Process

1. **Checkout Catalog Repo**
   ```yaml
   - uses: actions/checkout@v4
     with:
       repository: owner/dm-nkp-gitops-app-catalog
       token: ${{ secrets.GITHUB_TOKEN }}
   ```

2. **Detect Structure**
   - Checks for HelmRelease files
   - Checks for OCIRepository files
   - Checks for catalog manifests
   - Falls back to auto-create if none found

3. **Update Version**
   - Uses `yq` if available (preferred)
   - Falls back to `sed` for basic updates
   - Updates version field in detected files

4. **Commit and Push**
   - Commits with message: `chore: update dm-nkp-gitops-a2a-server to {version} [skip ci]`
   - Pushes to master/main branch
   - `[skip ci]` prevents catalog repo CI from running

---

## Configuration

### Environment Variables

Set in the CD workflow:

```yaml
env:
  CATALOG_REPO: dm-nkp-gitops-app-catalog
  CATALOG_OWNER: ${{ github.repository_owner }}
```

### Permissions Required

The workflow needs:

```yaml
permissions:
  contents: write  # To push to catalog repo
```

### Authentication

Uses `GITHUB_TOKEN` (automatically provided) to:
- Checkout catalog repository
- Commit changes
- Push to catalog repository

**Note:** The catalog repository must be accessible with the same token. For private repos or cross-org, you may need a Personal Access Token (PAT) stored as a secret.

---

## Example Catalog Structures

### Example 1: Flux CD with HelmRelease

**File:** `apps/dm-nkp-gitops-a2a-server/helmrelease.yaml`

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: dm-nkp-gitops-a2a-server
  namespace: gitops-agent
spec:
  interval: 5m
  chart:
    spec:
      chart: dm-nkp-gitops-a2a-server
      version: master-a1b2c3d  # ← Auto-updated
      sourceRef:
        kind: OCIRepository
        name: dm-nkp-gitops-a2a-server-chart
        namespace: gitops-agent
```

**File:** `apps/dm-nkp-gitops-a2a-server/ocirepository.yaml`

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: dm-nkp-gitops-a2a-server-chart
  namespace: gitops-agent
spec:
  interval: 5m
  url: oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server
  ref:
    tag: master-a1b2c3d  # ← Auto-updated
```

### Example 2: Simple Catalog Manifest

**File:** `catalog.yaml`

```yaml
applications:
  dm-nkp-gitops-a2a-server:
    name: dm-nkp-gitops-a2a-server
    version: master-a1b2c3d  # ← Auto-updated
    chart: oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server
    namespace: gitops-agent
```

---

## Troubleshooting

### Issue: Catalog update fails with "permission denied"

**Solution:** Ensure the workflow has `contents: write` permission and the catalog repo is accessible.

**For cross-org or private repos:**
1. Create a Personal Access Token (PAT) with `repo` scope
2. Add it as a secret: `CATALOG_REPO_TOKEN`
3. Update workflow to use:
   ```yaml
   token: ${{ secrets.CATALOG_REPO_TOKEN }}
   ```

### Issue: Version not updating correctly

**Solution:** Check the catalog structure detection. The workflow looks for:
1. `apps/{app-name}/helmrelease.yaml`
2. `apps/{app-name}/ocirepository.yaml`
3. `catalog.yaml` or `apps.yaml`

Ensure your catalog follows one of these patterns.

### Issue: yq not found, using sed fallback

**Solution:** This is normal. The workflow uses `sed` as a fallback if `yq` is not available. For better YAML handling, you can add:

```yaml
- name: Install yq
  run: |
    sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
    sudo chmod +x /usr/local/bin/yq
```

### Issue: Changes not being committed

**Check:**
1. Is the file path correct?
2. Does the version already match?
3. Check workflow logs for "No changes detected"

---

## Advanced Configuration

### Custom Catalog Path

If your catalog uses a different structure, you can customize the detection:

```yaml
- name: Custom catalog detection
  run: |
    # Add your custom detection logic
    if [ -f "custom/path/to/app.yaml" ]; then
      echo "strategy=custom" >> $GITHUB_OUTPUT
      echo "file_path=custom/path/to/app.yaml" >> $GITHUB_OUTPUT
    fi
```

### Multiple Apps in One Catalog

The workflow is designed for one app per run. For multiple apps:

1. **Option A:** Run separate workflows for each app
2. **Option B:** Modify the workflow to update multiple apps
3. **Option C:** Use a catalog manifest that lists all apps

### Conditional Updates

Only update on releases (not branch builds):

```yaml
if: always() && needs.helm.result == 'success' && needs.helm.outputs.is_release == 'true'
```

---

## Best Practices

1. **Use `[skip ci]` in commit messages** - Prevents catalog repo CI from running
2. **Version format consistency** - Use semantic versioning for releases
3. **Catalog structure** - Prefer HelmRelease + OCIRepository pattern
4. **Testing** - Test catalog updates in a separate branch first
5. **Monitoring** - Monitor catalog repo for successful updates

---

## Summary

✅ **Automated**: No manual intervention needed  
✅ **Flexible**: Supports multiple catalog structures  
✅ **Safe**: Only updates on successful builds  
✅ **Traceable**: Clear commit messages with versions  
✅ **GitOps-friendly**: Works seamlessly with Flux CD  

The catalog update automation ensures your application catalog always references the latest built and signed artifacts, maintaining a true GitOps workflow.
