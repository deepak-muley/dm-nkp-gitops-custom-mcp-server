# CI/CD Architecture

## Current Setup

### CI Workflow
- **Triggers**: Push to `master/main` branches, Pull Requests
- **Purpose**: Validate code quality, run tests, build binaries
- **Jobs**: `test` → `build` → `helm` → `docker`
- **Outputs**: Docker images (for PRs: not pushed, for branches: pushed)

### CD Workflow
- **Triggers**: Push to `master/main` branches, Tags (`v*`)
- **Purpose**: Build, sign, and deploy artifacts
- **Jobs**: `docker` → `helm` → `update-catalog`
- **Outputs**: Docker images, Helm charts, Catalog updates

## Best Practice: Sequential Execution

**Current State**: CI and CD run in parallel (both triggered by same push)

**Recommended**: CD should run **after** CI succeeds

### Why Sequential?

1. **Quality Gate**: Ensures code passes tests before deployment
2. **Avoid Duplicate Work**: CI already builds Docker images
3. **Fail Fast**: Don't waste resources on CD if CI fails
4. **Clear Separation**: CI = validation, CD = deployment

### Implementation Options

#### Option 1: `workflow_run` Trigger (Recommended)

CD workflow runs after CI workflow completes successfully:

```yaml
# CD workflow
on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]
    branches: [master, main]
  tags: ['v*']  # Tags can still trigger immediately
```

**Pros:**
- Clear separation of concerns
- CD only runs if CI passes
- Works well with PRs (CI runs, CD doesn't)

**Cons:**
- Slight delay (workflow_run has ~1-2 min delay)
- Need to checkout code again in CD

#### Option 2: Combined Workflow

Single workflow with job dependencies:

```yaml
jobs:
  ci-test:
    # ... test job
  ci-build:
    needs: ci-test
    # ... build job
  cd-docker:
    needs: ci-build
    # ... docker job
```

**Pros:**
- Faster (no workflow_run delay)
- Single workflow view
- Shared artifacts

**Cons:**
- Larger workflow file
- Less separation of concerns

#### Option 3: Hybrid Approach (Current + Recommended)

- **For branch pushes**: CD runs after CI (`workflow_run`)
- **For tags**: CD runs immediately (tags are typically created after CI passes)

This is the **recommended approach** for this repository.

## Recommended Changes

1. **CD workflow** should use `workflow_run` for branch pushes
2. **Tags** can still trigger CD immediately (assumed to be created after CI passes)
3. **CI workflow** should output artifacts that CD can reuse (optional optimization)

## Example: workflow_run Implementation

```yaml
# .github/workflows/cd.yaml
on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]
    branches: [master, main]
  tags:
    - 'v*'

jobs:
  docker:
    # Only run if CI succeeded
    if: ${{ github.event.workflow_run.conclusion == 'success' || github.event_name == 'push' }}
    steps:
      - uses: actions/checkout@v4
        with:
          # Use the commit that triggered CI
          ref: ${{ github.event.workflow_run.head_branch || github.ref }}
          sha: ${{ github.event.workflow_run.head_sha || github.sha }}
```

## Benefits

✅ **Quality Gate**: CD only runs if CI passes  
✅ **Efficiency**: No duplicate builds if CI fails  
✅ **Clarity**: Clear separation between validation and deployment  
✅ **Flexibility**: Tags can still trigger immediately  

## Migration Path

1. Update CD workflow to use `workflow_run` for branches
2. Keep tag triggers immediate
3. Test with a branch push
4. Verify CD only runs after CI succeeds
