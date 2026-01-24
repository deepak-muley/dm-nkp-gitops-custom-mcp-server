# Workflow Verification Guide

Quick guide to verify that CI and CD workflows are working correctly.

## Quick Verification

### 1. Check Workflow Runs

Go to: https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/actions

**Expected:**
- ✅ CI workflow should be running/complete
- ✅ CD workflow should be running/complete (on master push)

### 2. CI Workflow Checklist

**Jobs that should pass:**
- [ ] **test** - Code compilation and tests
- [ ] **build** - Builds MCP and A2A servers
- [ ] **helm** - Lints Helm chart
- [ ] **docker** - Builds, pushes, and signs Docker image

**What to check:**
1. **test job:**
   - ✅ Code compiles successfully
   - ✅ Tests run (or skip gracefully if none)

2. **build job:**
   - ✅ `make build` succeeds
   - ✅ `make build-a2a` succeeds

3. **helm job:**
   - ✅ Helm chart lints without errors

4. **docker job:**
   - ✅ Docker image builds successfully
   - ✅ Image is pushed to `ghcr.io`
   - ✅ Image is signed with Cosign (may show warnings, but non-blocking)
   - ✅ Artifacts are uploaded

### 3. CD Workflow Checklist

**Jobs that should pass:**
- [ ] **docker** - Builds, pushes, and signs Docker image
- [ ] **helm** - Packages, pushes, and signs Helm chart
- [ ] **update-catalog** - Updates catalog repository (may skip if repo not accessible)

**What to check:**
1. **docker job:**
   - ✅ Image builds with version tag
   - ✅ Image is pushed with tags: `{version}` and `{branch}-latest` or `latest`
   - ✅ Image is signed (with retry logic)

2. **helm job:**
   - ✅ Chart is packaged
   - ✅ Chart is pushed to OCI registry
   - ✅ Chart is signed
   - ✅ Job outputs are set correctly

3. **update-catalog job:**
   - ✅ Catalog repo is checked out (or shows clear error if not accessible)
   - ✅ Version is updated in catalog files
   - ✅ PR is created (if using PR method) or direct push succeeds

## Common Issues and Solutions

### Issue: "Image digest is empty"

**Cause:** Image wasn't pushed or build failed

**Check:**
1. Look at docker build step logs
2. Verify `push: true` is set
3. Check registry permissions

**Solution:** Build step should show the digest in logs

### Issue: "Signing failed"

**Cause:** Image not available in registry yet, or permission issues

**What happens now:**
- ✅ Workflow continues (non-blocking)
- ✅ Image is still available without signature
- ✅ Retry logic attempts 3 times with delays

**If it fails:**
- Check if image exists: `docker manifest inspect ghcr.io/owner/image@digest`
- Verify `id-token: write` permission is set
- Check Cosign logs for specific error

### Issue: "Catalog update skipped"

**Cause:** Catalog repo not accessible

**What happens:**
- ✅ Workflow continues (non-blocking)
- ✅ Clear error message with setup instructions

**Solution:**
- If same-org: Should work automatically
- If cross-org: Add `CATALOG_REPO_TOKEN` secret

### Issue: "Helm chart packaging failed"

**Cause:** Chart.yaml not found or invalid

**Check:**
1. Verify `chart/dm-nkp-gitops-a2a-server/Chart.yaml` exists
2. Check Chart.yaml syntax
3. Look at helm package logs

## Verification Commands

### Check if image was pushed:

```bash
# List images in your registry
gh api user/packages?package_type=container | jq '.[] | select(.name | contains("dm-nkp-gitops-a2a-server"))'

# Or use docker
docker manifest inspect ghcr.io/deepak-muley/dm-nkp-gitops-a2a-server:master-{sha}
```

### Check if image is signed:

```bash
cosign verify ghcr.io/deepak-muley/dm-nkp-gitops-a2a-server@<digest> \
  --certificate-identity-regexp ".*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

### Check Helm chart:

```bash
# List charts in OCI registry
helm show chart oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server --version <version>
```

## Success Indicators

### ✅ CI Workflow Success

- All jobs show green checkmarks
- Docker image appears in GitHub Packages
- Artifacts are uploaded

### ✅ CD Workflow Success

- Docker image pushed with correct tags
- Helm chart pushed to OCI registry
- Catalog PR created (or direct push succeeded)
- All signatures verified

## Monitoring

### Real-time Monitoring

1. Go to: https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/actions
2. Click on the latest workflow run
3. Watch jobs execute in real-time
4. Check logs for any warnings or errors

### Key Metrics to Watch

- **Build time:** Should be < 5 minutes
- **Image size:** Check in GitHub Packages
- **Signing success rate:** Should be > 90%
- **Catalog update success:** Depends on repo access

## Next Steps After Verification

1. ✅ **If all green:** Workflows are working correctly!
2. ⚠️ **If warnings:** Check logs, but workflow should still complete
3. ❌ **If failures:** Check specific job logs for details

## Quick Links

- **CI Workflow:** https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/actions/workflows/ci.yaml
- **CD Workflow:** https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/actions/workflows/cd.yaml
- **Packages:** https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkgs/container/dm-nkp-gitops-a2a-server
