# Catalog Update Quick Start

## Security Setup

### Option 1: Same Organization (No PAT Needed)

If your catalog repo is in the **same GitHub organization**, `GITHUB_TOKEN` works automatically:

✅ **No setup required** - Works out of the box!

### Option 2: Cross-Organization (PAT Required)

If your catalog repo is in a **different organization**, you need a Personal Access Token:

#### Step 1: Create PAT

1. Go to: https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Name: `Catalog Update Automation`
4. Expiration: 90 days (renewable)
5. Scopes: ✅ `repo` (Full control of private repositories)
6. Click "Generate token"
7. **Copy the token immediately** (you won't see it again!)

#### Step 2: Add as Secret

1. Go to your repository → **Settings** → **Secrets and variables** → **Actions**
2. Click **"New repository secret"**
3. Name: `CATALOG_REPO_TOKEN`
4. Value: Paste your PAT
5. Click **"Add secret"**

✅ **Done!** The workflow will automatically use the PAT.

---

## Update Method Configuration

### Method 1: Pull Request (Recommended - Default)

**Benefits:**
- ✅ Review before merge
- ✅ CI/CD validation
- ✅ Better audit trail
- ✅ More secure

**Configuration:**
- Default: Already enabled (`CATALOG_UPDATE_METHOD: pr`)
- No action needed

### Method 2: Direct Push

**Benefits:**
- ✅ Faster updates
- ✅ No manual intervention

**Configuration:**
1. Go to repository → **Settings** → **Secrets and variables** → **Actions** → **Variables**
2. Click **"New repository variable"**
3. Name: `CATALOG_UPDATE_METHOD`
4. Value: `push`
5. Click **"Add variable"**

Or edit `.github/workflows/cd.yaml`:
```yaml
CATALOG_UPDATE_METHOD: push  # Change from 'pr' to 'push'
```

---

## Quick Decision Guide

```
Is catalog repo in same org?
├─ Yes
│   └─ Want review? → Use PR (default) ✅
│   └─ Want speed? → Use direct push
│
└─ No (cross-org)
    └─ Need PAT → Add CATALOG_REPO_TOKEN secret
    └─ Always use PR (recommended) ✅
```

---

## Testing

1. **Push a change to master branch**
2. **Monitor the workflow:**
   - Go to: Actions → CD workflow
   - Check `update-catalog` job
3. **Verify result:**
   - **PR method**: Check catalog repo for new PR
   - **Push method**: Check catalog repo for new commit

---

## Troubleshooting

### "Resource not accessible by integration"

**Cause:** GITHUB_TOKEN can't access catalog repo

**Solution:**
- If cross-org: Add `CATALOG_REPO_TOKEN` secret (PAT)
- If same-org: Check repository permissions

### "Permission denied"

**Cause:** Insufficient permissions

**Solution:**
- Verify PAT has `repo` scope
- Check workflow has `contents: write` permission

### PR not created

**Cause:** PR creation failed

**Solution:**
- Check if branch already exists (workflow will update existing PR)
- Verify `pull-requests: write` permission
- Check catalog repo settings (branch protection, etc.)

---

## Summary

| Setup | PAT Needed? | Method | Security |
|-------|-------------|--------|----------|
| Same org | ❌ No | PR (default) | ✅ High |
| Same org | ❌ No | Push | ⚠️ Medium |
| Cross org | ✅ Yes | PR (recommended) | ✅ High |
| Cross org | ✅ Yes | Push | ⚠️ Medium |

**Recommended:** Use PR method with PAT (even for same-org) for maximum security and flexibility.
