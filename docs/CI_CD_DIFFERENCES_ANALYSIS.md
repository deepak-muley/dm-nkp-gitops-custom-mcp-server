# CI/CD Workflow Differences Analysis

## Key Findings

### 1. CD Trigger Difference (PRIORITY: High)

**MCP Server (Correct):**
- Uses `workflow_run` trigger (CD waits for CI to complete)
- Quality gate implemented
- Tags can trigger immediately

**Custom App (Needs Update):**
- Uses direct `push` trigger (CD runs independently)
- No quality gate between CI and CD
- Could fail if CI hasn't passed

**Impact:** Security and reliability issue - CD can deploy untested code

---

### 2. CI/CD Jobs Comparison

#### MCP Server CI Jobs:
1. `test` - Unit tests + coverage
2. `build` - Build binaries
3. `helm` - Lint Helm chart
4. `kubesec` - K8s manifest security
5. `docker` - Build Docker image
6. Security scans - (separate workflow)

#### Custom App CI Jobs:
1. `test` - Unit + integration tests + coverage
2. `build` - Build binaries
3. `docker-build` - Build Docker image + push to dev
4. `helm` - Package Helm chart + push to dev
5. `kubesec` - K8s manifest security
6. `e2e` - E2E tests (depends on docker-build & helm)
7. Security scans - (separate workflow)

#### Differences Found:

| Feature | MCP Server | Custom App | Status |
|---------|-----------|-----------|--------|
| Docker push in CI | ❌ No | ✅ Yes (dev registry) | Custom App has feature |
| Helm push in CI | ❌ No | ✅ Yes (dev registry) | Custom App has feature |
| E2E in CI | ❌ No | ✅ Yes | Custom App has feature |
| Kubesec | ✅ Yes | ✅ Yes | ✅ Aligned |
| Integration tests | ⚠️ Optional | ✅ Yes | Custom App more robust |
| Coverage upload | ✅ Yes | ✅ Yes | ✅ Aligned |

---

### 3. CD/Deployment Flow

**MCP Server:**
- `docker` job: Build + sign + push to prod
- `helm` job: Sign + push to prod
- `update-catalog` job: Update external repo
- Sequential: docker → helm → catalog

**Custom App:**
- `build-and-push` job: Build + sign + push to prod
- `e2e` job: Test prod artifacts (depends on build-and-push)
- Combined approach: Single job does both docker + helm

---

## Recommended Actions

### 1. Update Custom App CD Trigger (MUST DO)
Change from direct `push` to `workflow_run` for quality gate

### 2. Add Dev Registry Publishing to MCP Server CI (OPTIONAL but Good)
Push to dev registry during CI for immediate testing

### 3. Align E2E Testing
- MCP Server: Has E2E in CD
- Custom App: Has E2E in CI
- Should decide: CI only (faster feedback) or both?

---

## Next Steps

1. Update Custom App CD to use `workflow_run`
2. Document why MCP Server doesn't push in CI (by design)
3. Decide on E2E placement strategy (CI vs CD)
4. Update alignment documentation
