# GitHub Runners - Concept and Usage Guide

## Overview

A **GitHub Runner** is a server that executes GitHub Actions workflows. It's the compute environment that runs your CI/CD jobs, executing the steps defined in your workflow files.

## Core Concept

When a workflow is triggered:

```
Workflow Triggered (push, PR, schedule, etc.)
    ↓
GitHub Queues Job
    ↓
Runner Picks Up Job
    ↓
Runner Checks Out Code
    ↓
Runner Executes Steps
    ↓
Runner Reports Results
    ↓
Runner Cleans Up (ephemeral)
```

## Types of Runners

### 1. GitHub-Hosted Runners (Default)

**Managed by GitHub** - Zero setup required

- **Pre-installed tools**: Docker, Git, Node.js, Python, Go, etc.
- **Available on-demand**: Automatically provisioned when needed
- **Free tier**: 2,000 minutes/month for private repos (unlimited for public)
- **Billing**: Consumes minutes from your quota
- **Isolation**: Each job runs on a fresh, clean runner

**Available runner types:**
- `ubuntu-latest` (Ubuntu 22.04)
- `ubuntu-22.04`
- `ubuntu-20.04`
- `windows-latest` (Windows Server 2022)
- `windows-2022`
- `macos-latest` (macOS 12)
- `macos-13`
- `macos-14`

**Example:**
```yaml
jobs:
  test:
    runs-on: ubuntu-latest  # ← GitHub-hosted Ubuntu runner
    steps:
      - run: make test
```

### 2. Self-Hosted Runners

**Managed by you** - Full control

- **Custom hardware/software**: Install any tools you need
- **No per-minute charges**: Only infrastructure costs
- **Network access**: Can access private networks/resources
- **Security**: Your responsibility
- **Maintenance**: You manage updates, scaling, etc.

**Use cases:**
- Custom hardware requirements (ARM64, GPUs, etc.)
- Security/compliance requirements (air-gapped, on-premise)
- Cost savings at scale (high usage)
- Access to private resources (databases, internal APIs)

**Example:**
```yaml
jobs:
  test:
    runs-on: self-hosted  # ← Your own runner
    steps:
      - run: make test
```

## Runner Lifecycle

### 1. Job Queued
- GitHub receives workflow trigger
- Job is added to queue
- Runner selection begins

### 2. Runner Provisioned
- **GitHub-hosted**: Automatically provisioned
- **Self-hosted**: Picks up job from queue

### 3. Environment Setup
- Checks out repository code
- Sets up environment (OS, tools, dependencies)
- Configures secrets and permissions

### 4. Step Execution
- Executes each step in sequence
- Captures logs and outputs
- Handles errors and failures

### 5. Artifact Collection
- Uploads artifacts (if any)
- Saves logs
- Reports status

### 6. Cleanup
- Removes all files and data
- Runner is destroyed (ephemeral)
- Results reported to GitHub

## What a Runner Does

### Code Execution
```yaml
steps:
  - run: make test  # ← Runner executes this command
```

### Tool Installation
```yaml
steps:
  - uses: actions/setup-go@v5  # ← Runner installs Go
    with:
      go-version: '1.25'
```

### Environment Access
```yaml
steps:
  - run: echo ${{ secrets.API_KEY }}  # ← Runner has access to secrets
```

### Artifact Management
```yaml
steps:
  - uses: actions/upload-artifact@v4  # ← Runner uploads files
    with:
      name: build-artifacts
      path: bin/
```

## Runner Characteristics

### Ephemeral
- Each job runs on a **fresh, isolated runner**
- No persistence between jobs
- Use artifacts/cache for data persistence

### Stateless
- No files persist between jobs
- Each job starts with a clean environment
- Use GitHub Actions cache for dependencies

### Parallel
- Multiple jobs can run **simultaneously** on different runners
- Jobs in the same workflow can run in parallel
- Limited by your runner availability

### Scalable
- GitHub automatically provisions runners as needed
- Self-hosted runners scale based on your infrastructure
- No manual scaling required (for GitHub-hosted)

## Runner Limitations

### Timeouts
- **Default**: 6 hours per job
- **Configurable**: Set `timeout-minutes` in workflow
- **Example**:
  ```yaml
  jobs:
    build:
      timeout-minutes: 30  # Job fails after 30 minutes
      runs-on: ubuntu-latest
  ```

### Resources
- **CPU**: Limited per runner (varies by runner type)
- **Memory**: Limited per runner
- **Disk**: Limited temporary storage
- **Network**: Internet access (can be restricted for self-hosted)

### Isolation
- Each job gets a **clean environment**
- No access to previous job's files
- Secrets are scoped to the job

### Cost
- **GitHub-hosted**: Consumes minutes from quota
- **Self-hosted**: Infrastructure costs only
- **Free tier**: 2,000 minutes/month for private repos

## Example: Your CI Workflow

```yaml
jobs:
  test:
    runs-on: ubuntu-latest  # ← GitHub-hosted Ubuntu runner
    steps:
      - uses: actions/checkout@v4      # Runner checks out code
      - uses: actions/setup-go@v5      # Runner installs Go
        with:
          go-version: '1.25'
      - run: make test                 # Runner executes tests
```

**What happens:**
1. GitHub provisions an Ubuntu 22.04 runner
2. Runner checks out your repository code
3. Runner installs Go 1.25
4. Runner runs `make test`
5. Runner captures logs and reports results
6. Runner is destroyed (ephemeral)

## Self-Hosted vs GitHub-Hosted

| Aspect | GitHub-Hosted | Self-Hosted |
|--------|---------------|-------------|
| **Management** | GitHub manages | You manage |
| **Setup** | Zero setup | Requires setup |
| **Cost** | Minutes quota | Infrastructure cost |
| **Customization** | Limited | Full control |
| **Security** | GitHub's security | Your security |
| **Availability** | Always available | Depends on your infra |
| **Tools** | Pre-installed | You install |
| **Scaling** | Automatic | Manual |
| **Maintenance** | None | You maintain |

## When to Use Each

### Use GitHub-Hosted When:
- ✅ Standard tools are sufficient
- ✅ You want zero maintenance
- ✅ Low to moderate usage
- ✅ Public repositories (unlimited free)
- ✅ Standard operating systems (Linux, Windows, macOS)

### Use Self-Hosted When:
- ✅ Custom hardware requirements (ARM64, GPUs)
- ✅ High usage (cost savings)
- ✅ Security/compliance requirements
- ✅ Access to private networks/resources
- ✅ Custom software/tools not available on GitHub-hosted
- ✅ Air-gapped environments

## Runner Selection

### Automatic Selection
```yaml
jobs:
  test:
    runs-on: ubuntu-latest  # GitHub selects available Ubuntu runner
```

### Label-Based (Self-Hosted)
```yaml
jobs:
  test:
    runs-on: [self-hosted, linux, x64]  # Uses self-hosted runner with these labels
```

### Matrix Strategy
```yaml
jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
```

## Best Practices

### 1. Use Appropriate Runner Type
```yaml
# For Go projects
runs-on: ubuntu-latest  # Fast, cost-effective

# For multi-platform builds
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
```

### 2. Set Timeouts
```yaml
jobs:
  build:
    timeout-minutes: 30  # Prevent stuck builds
    runs-on: ubuntu-latest
```

### 3. Use Caching
```yaml
steps:
  - uses: actions/cache@v3  # Cache dependencies
    with:
      path: ~/go/pkg/mod
      key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

### 4. Clean Up Resources
```yaml
steps:
  - name: Cleanup
    if: always()  # Run even if job fails
    run: docker system prune -f
```

### 5. Use Concurrency Groups
```yaml
concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true  # Cancel previous runs
```

## Troubleshooting

### Runner Not Available
- **Issue**: Job stuck in "Queued" state
- **Cause**: No runners available (self-hosted) or quota exceeded
- **Fix**: Check runner status, increase quota, or add more self-hosted runners

### Job Timeout
- **Issue**: Job fails after timeout
- **Cause**: Job takes too long
- **Fix**: Increase `timeout-minutes` or optimize job steps

### Out of Disk Space
- **Issue**: Job fails with disk space errors
- **Cause**: Too many files/artifacts
- **Fix**: Clean up artifacts, use smaller cache, or optimize build

### Network Issues
- **Issue**: Downloads fail
- **Cause**: Network connectivity problems
- **Fix**: Check network settings, use retries, or use self-hosted runner with better network

## Related Concepts

- **Workflows**: Define what runs on runners
- **Jobs**: Units of work executed on runners
- **Steps**: Individual commands/actions in jobs
- **Actions**: Reusable workflow components
- **Artifacts**: Files produced by runners
- **Cache**: Persistent storage for dependencies

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub-Hosted Runners](https://docs.github.com/en/actions/using-github-hosted-runners)
- [Self-Hosted Runners](https://docs.github.com/en/actions/hosting-your-own-runners)
- [Runner Specifications](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners#supported-runners-and-hardware-resources)

## Summary

**GitHub Runners** are the compute environments that execute your CI/CD workflows. They can be:
- **GitHub-hosted**: Managed by GitHub, zero setup, pay-per-minute
- **Self-hosted**: Managed by you, full control, infrastructure costs

Understanding runners helps you:
- Choose the right runner type for your needs
- Optimize workflow performance
- Control costs
- Troubleshoot issues
- Make informed decisions about infrastructure

In your workflows, you specify runners with `runs-on:`, and GitHub handles the rest (for GitHub-hosted) or you manage them yourself (for self-hosted).
