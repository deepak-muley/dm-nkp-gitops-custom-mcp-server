# Catalog Update Security Guide

This document explains the security model for automated catalog updates and when you need a Personal Access Token (PAT).

## Security Model Overview

### Current Implementation

The workflow uses **two authentication methods**:

1. **GITHUB_TOKEN** (default) - Automatically provided by GitHub Actions
2. **PAT** (optional) - Personal Access Token stored as secret

### Authentication Scenarios

| Scenario | GITHUB_TOKEN | PAT Required? | Notes |
|----------|--------------|--------------|-------|
| **Same org, direct push** | ✅ Works | ❌ No | Limited to same organization |
| **Same org, create PR** | ✅ Works | ❌ No | Can create PRs in same org |
| **Cross-org, direct push** | ❌ Fails | ✅ Yes | Need PAT with `repo` scope |
| **Cross-org, create PR** | ❌ Fails | ✅ Yes | Need PAT with `repo` scope |
| **Private repo** | ⚠️ Limited | ✅ Recommended | PAT provides better control |

---

## GITHUB_TOKEN Limitations

### What GITHUB_TOKEN Can Do

✅ **Same organization repositories:**
- Read/write access to repos in the same org
- Create branches
- Create pull requests
- Push commits (with proper permissions)

✅ **Permissions:**
- Automatically scoped to the repository running the workflow
- Can be extended with `permissions:` block

### What GITHUB_TOKEN Cannot Do

❌ **Cross-organization:**
- Cannot access repos in different organizations
- Cannot create PRs across orgs

❌ **Enterprise restrictions:**
- May be restricted by enterprise policies
- Limited by organization settings

---

## When You Need a PAT

### Required Scenarios

1. **Cross-organization access**
   - Catalog repo is in a different GitHub organization
   - Example: `org1/app-repo` → `org2/catalog-repo`

2. **Enterprise restrictions**
   - Organization policies restrict GITHUB_TOKEN
   - Enterprise SSO requirements

3. **Better security control**
   - Want to use a service account
   - Need fine-grained permissions
   - Audit trail requirements

### PAT Setup

#### Step 1: Create Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Set expiration (recommended: 90 days, renewable)
4. Select scopes:
   - ✅ `repo` - Full control of private repositories
   - ✅ `workflow` - Update GitHub Action workflows (if needed)

#### Step 2: Add as Secret

1. Go to your repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `CATALOG_REPO_TOKEN`
4. Value: Paste your PAT
5. Click "Add secret"

#### Step 3: Update Workflow

The workflow automatically uses PAT if available:

```yaml
token: ${{ secrets.CATALOG_REPO_TOKEN || secrets.GITHUB_TOKEN }}
```

This means:
- Uses PAT if `CATALOG_REPO_TOKEN` secret exists
- Falls back to `GITHUB_TOKEN` if PAT not available

---

## Security Best Practices

### 1. Use Pull Requests (Recommended)

**Why PRs are better:**
- ✅ Review before merge
- ✅ CI/CD validation
- ✅ Audit trail
- ✅ Rollback capability
- ✅ No direct write access needed

**Current workflow supports both:**
- Direct push (faster, less secure)
- PR creation (slower, more secure)

### 2. Least Privilege Principle

**PAT scopes:**
- Minimum: `repo` scope (read/write to repos)
- Avoid: `admin:repo` (unnecessary permissions)

**Workflow permissions:**
```yaml
permissions:
  contents: write  # Only what's needed
  pull-requests: write  # For PR creation
```

### 3. Token Rotation

**Best practices:**
- Rotate PATs every 90 days
- Use service accounts for automation
- Monitor token usage
- Revoke unused tokens

### 4. Secret Management

**Never:**
- ❌ Commit tokens to code
- ❌ Log tokens in workflow output
- ❌ Use tokens in public repos (unless read-only)

**Always:**
- ✅ Store in GitHub Secrets
- ✅ Use environment-specific secrets
- ✅ Rotate regularly

---

## Workflow Configuration

### Option 1: Direct Push (Current Default)

**Use case:** Same org, fast updates, trusted automation

```yaml
- name: Commit and push catalog update
  run: |
    git commit -m "chore: update version"
    git push origin HEAD:master
```

**Security:** ⚠️ Medium - Direct write access

### Option 2: Pull Request (Recommended)

**Use case:** Cross-org, review required, better audit

```yaml
- name: Create Pull Request
  uses: peter-evans/create-pull-request@v5
  with:
    token: ${{ secrets.CATALOG_REPO_TOKEN || secrets.GITHUB_TOKEN }}
    commit-message: "chore: update dm-nkp-gitops-a2a-server to ${VERSION}"
    title: "chore: update dm-nkp-gitops-a2a-server to ${VERSION}"
    body: |
      Automated catalog update from CD workflow.
      
      - Chart version: ${VERSION}
      - Chart URL: oci://ghcr.io/owner/charts/dm-nkp-gitops-a2a-server
    branch: update/dm-nkp-gitops-a2a-server-${VERSION}
    delete-branch: true
```

**Security:** ✅ High - Requires review before merge

---

## Configuration Matrix

| Scenario | Token | Method | Security Level |
|----------|-------|--------|----------------|
| Same org, same repo | GITHUB_TOKEN | Direct push | ⚠️ Medium |
| Same org, different repo | GITHUB_TOKEN | Direct push | ⚠️ Medium |
| Same org, different repo | GITHUB_TOKEN | PR creation | ✅ High |
| Cross org | PAT | Direct push | ⚠️ Medium |
| Cross org | PAT | PR creation | ✅ High |
| Private repo, same org | GITHUB_TOKEN | PR creation | ✅ High |
| Private repo, cross org | PAT | PR creation | ✅ High |

---

## Recommended Setup

### For Production (Recommended)

```yaml
# Use PR creation with PAT
env:
  CATALOG_UPDATE_METHOD: "pr"  # or "push"
  CATALOG_REPO_TOKEN: ${{ secrets.CATALOG_REPO_TOKEN || secrets.GITHUB_TOKEN }}
```

**Benefits:**
- Review before merge
- CI/CD validation
- Better audit trail
- Works cross-org with PAT

### For Development/Testing

```yaml
# Use direct push with GITHUB_TOKEN
env:
  CATALOG_UPDATE_METHOD: "push"
```

**Benefits:**
- Faster updates
- No manual intervention
- Good for same-org scenarios

---

## Troubleshooting

### Error: "Resource not accessible by integration"

**Cause:** GITHUB_TOKEN doesn't have access to target repo

**Solution:**
1. Check if repos are in same org
2. If cross-org, add PAT as `CATALOG_REPO_TOKEN` secret
3. Verify PAT has `repo` scope

### Error: "Permission denied"

**Cause:** Insufficient permissions

**Solution:**
1. Add `contents: write` permission to workflow
2. For PRs, add `pull-requests: write`
3. Verify PAT has correct scopes

### Error: "Repository not found"

**Cause:** Token can't access the repository

**Solution:**
1. Verify repository name is correct
2. Check if repo is private (may need PAT)
3. Verify token has access to the repo

---

## Security Checklist

Before enabling catalog updates:

- [ ] Determine if PAT is needed (cross-org?)
- [ ] Choose update method (PR vs direct push)
- [ ] Create PAT with minimal scopes (`repo` only)
- [ ] Store PAT as GitHub Secret
- [ ] Test with a test catalog repo first
- [ ] Enable branch protection on catalog repo (if using PRs)
- [ ] Set up required reviews (if using PRs)
- [ ] Monitor workflow runs for security issues
- [ ] Rotate PATs regularly (every 90 days)

---

## Summary

### Quick Decision Tree

```
Is catalog repo in same org?
├─ Yes → GITHUB_TOKEN works
│   └─ Want review? → Use PR method
│   └─ Want speed? → Use direct push
│
└─ No → Need PAT
    └─ Always use PR method (recommended)
    └─ Add CATALOG_REPO_TOKEN secret
```

### Recommended Configuration

**For most cases:**
- ✅ Use PR creation method
- ✅ Add PAT as `CATALOG_REPO_TOKEN` (even for same-org, for flexibility)
- ✅ Enable branch protection on catalog repo
- ✅ Require at least one review (or auto-merge for trusted workflows)

This provides the best balance of security, auditability, and automation.
