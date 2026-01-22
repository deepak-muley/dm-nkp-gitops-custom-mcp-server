# Catalog Update - Secret Configuration

## Quick Answer: Do You Need to Set Secrets?

### ✅ **No Secrets Needed** (Default - Works Out of the Box)

If your catalog repository is in the **same GitHub organization**, the workflow works automatically using `GITHUB_TOKEN`. No setup required!

### ⚠️ **Secret Needed** (Only for Cross-Organization)

If your catalog repository is in a **different organization**, you need to add one secret: `CATALOG_REPO_TOKEN`

---

## Setup Instructions

### Step 1: Check Your Setup

**Question:** Is `dm-nkp-gitops-app-catalog` in the same organization as this repo?

- ✅ **Yes** → No action needed! It will work automatically.
- ❌ **No** → Continue to Step 2

### Step 2: Create Personal Access Token (PAT)

**Only needed if catalog repo is in a different org.**

1. Go to: https://github.com/settings/tokens
2. Click **"Generate new token (classic)"**
3. **Token name**: `Catalog Update Automation`
4. **Expiration**: 90 days (or your preference)
5. **Select scopes**: ✅ `repo` (Full control of private repositories)
6. Click **"Generate token"**
7. **Copy the token** (you won't see it again!)

### Step 3: Add Secret to Repository

1. Go to your repository: `https://github.com/{org}/dm-nkp-gitops-custom-mcp-server`
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **"New repository secret"**
4. **Name**: `CATALOG_REPO_TOKEN`
5. **Secret**: Paste your PAT
6. Click **"Add secret"**

✅ **Done!** The workflow will automatically use the PAT.

---

## How It Works

The workflow uses **automatic fallback**:

```yaml
# Automatically chooses the right token
token: ${{ secrets.CATALOG_REPO_TOKEN || secrets.GITHUB_TOKEN }}
```

**Behavior:**
- If `CATALOG_REPO_TOKEN` exists → Uses PAT (works cross-org)
- If `CATALOG_REPO_TOKEN` doesn't exist → Uses `GITHUB_TOKEN` (same-org only)

**You don't need to configure anything** - it just works!

---

## Verification

### Test the Workflow

1. Push a change to master branch
2. Go to: **Actions** → **CD** workflow
3. Check the `update-catalog` job
4. Look at the summary - it will show:
   - ✅ "Using PAT" (if secret is set)
   - ℹ️ "Using GITHUB_TOKEN" (if no secret)

### Expected Results

**Same-org (no secret):**
```
✅ Catalog updated successfully
ℹ️ Using GITHUB_TOKEN (same-org only)
```

**Cross-org (with secret):**
```
✅ Pull request created successfully
✅ Using PAT (cross-org support enabled)
```

---

## Troubleshooting

### Error: "Resource not accessible by integration"

**Cause:** `GITHUB_TOKEN` can't access catalog repo (likely cross-org)

**Solution:**
1. Add `CATALOG_REPO_TOKEN` secret (see Step 2-3 above)
2. Verify PAT has `repo` scope
3. Verify catalog repo name is correct

### Error: "Repository not found"

**Cause:** Token can't access the repository

**Solution:**
1. Check repository name: `${{ env.CATALOG_OWNER }}/${{ env.CATALOG_REPO }}`
2. Verify PAT has access to the catalog repo
3. If private repo, ensure PAT has `repo` scope

### Error: "Permission denied"

**Cause:** Insufficient permissions

**Solution:**
1. Verify PAT has `repo` scope (not just `public_repo`)
2. Check workflow has `contents: write` permission
3. For PRs, ensure `pull-requests: write` permission

---

## Security Best Practices

### ✅ Recommended

- Use PR method (default) - requires review
- Rotate PATs every 90 days
- Use minimal scopes (`repo` only)
- Monitor workflow runs

### ❌ Avoid

- Don't commit tokens to code
- Don't use admin scopes unnecessarily
- Don't share PATs between services

---

## Summary

| Scenario | Secret Needed? | Setup Time |
|----------|----------------|------------|
| Same org | ❌ No | 0 minutes |
| Cross org | ✅ Yes (PAT) | 2 minutes |

**Most common case (same org):** ✅ **No setup needed - it just works!**

---

## Quick Reference

**Check if secret is needed:**
```bash
# If catalog repo is: {your-org}/dm-nkp-gitops-app-catalog
# And this repo is: {your-org}/dm-nkp-gitops-custom-mcp-server
# → Same org = No secret needed ✅
```

**Add secret (if needed):**
1. Create PAT with `repo` scope
2. Add as `CATALOG_REPO_TOKEN` secret
3. Done!

The workflow handles everything else automatically.
