# Security Scanning Workflow Guide

Complete guide to understanding the security scanning workflow (`security.yaml`) and what you get from each scanner.

## Overview

The security workflow runs **8 comprehensive security scanners** on every push, pull request, and daily via schedule. All results are uploaded to GitHub's **Security → Code scanning alerts** tab for centralized viewing and management.

**Workflow File:** `.github/workflows/security.yaml`

**Triggers:**
- ✅ Every push to `master` or `main` branches
- ✅ Every pull request
- ✅ Daily at 2 AM UTC (scheduled)

---

## Security Scanners Breakdown

### 1. CodeQL Security Analysis

**Type:** SAST (Static Application Security Testing)  
**Coverage:** Go source code analysis  
**Tool:** GitHub CodeQL (native integration)

**What It Finds:**
- SQL injection vulnerabilities
- Cross-site scripting (XSS) risks
- Insecure deserialization
- Hardcoded secrets/credentials
- Insecure random number generation
- Path traversal vulnerabilities
- Command injection risks
- Code quality issues that could lead to security problems

**Configuration:**
- Language: Go
- Query set: `security-and-quality` (comprehensive security queries)
- Category: `/language:go`

**Where to View:**
- GitHub → Security → Code scanning alerts
- Filter by: `codeql` or `/language:go`

**What to Do:**
1. Review alerts in Security tab
2. Fix high/critical severity issues immediately
3. Mark false positives as "Dismiss" with reason
4. Update code to address security issues

**Example Findings:**
- "Hardcoded API key in source code"
- "SQL query constructed from user input"
- "Insecure random number generator used"

---

### 2. Trivy Filesystem Scan

**Type:** SCA (Software Composition Analysis)  
**Coverage:** Filesystem + Go modules + OS packages  
**Tool:** Aqua Security Trivy

**What It Finds:**
- Known CVEs in Go dependencies (`go.mod`)
- Vulnerabilities in OS packages (if scanning container filesystem)
- Outdated dependencies with security issues
- CVEs from multiple vulnerability databases (NVD, GitHub, etc.)

**Configuration:**
- Scan type: Filesystem (`fs`)
- Severity levels: `CRITICAL`, `HIGH`, `MEDIUM`
- Format: SARIF (for GitHub integration)
- Skip directories: `test`, `vendor`

**Where to View:**
- GitHub → Security → Code scanning alerts
- Filter by: `trivy-fs`
- Also check: Dependabot alerts (complementary)

**What to Do:**
1. Review CVEs by severity (Critical → High → Medium)
2. Update vulnerable dependencies: `go get -u {package}@latest`
3. Check if patches are available for specific versions
4. Consider alternatives if package is unmaintained

**Example Findings:**
- "CVE-2024-12345 in github.com/example/package v1.2.3"
- "High severity vulnerability in golang.org/x/crypto"
- "Outdated dependency with known security issues"

---

### 3. OWASP Dependency-Check

**Type:** SCA (Software Composition Analysis)  
**Coverage:** Dependency vulnerability detection using OWASP database  
**Tool:** OWASP Dependency-Check

**What It Finds:**
- CVEs from OWASP's comprehensive vulnerability database
- Known vulnerabilities in all dependency types
- License compliance issues
- Outdated components

**Configuration:**
- Format: SARIF
- Experimental features: Enabled
- Log level: DEBUG (for troubleshooting)
- Suppression file: `suppression.xml` (if exists, for false positives)

**Where to View:**
- GitHub → Security → Code scanning alerts
- Filter by: `dependency-check`

**What to Do:**
1. Review OWASP-specific findings
2. Cross-reference with Trivy results (different databases)
3. Update dependencies to patched versions
4. Add suppressions for false positives in `suppression.xml`

**Example Findings:**
- "OWASP-2024-001: Vulnerability in dependency X"
- "License conflict detected"
- "Component with known security issues"

---

### 4. Gosec Security Scanner

**Type:** SAST (Static Application Security Testing)  
**Coverage:** Go-specific security issues  
**Tool:** SecureGo Gosec

**What It Finds:**
- Insecure TLS configurations
- Hardcoded credentials
- SQL injection risks
- Weak cryptographic functions
- Insecure file permissions
- Race conditions
- Error handling issues
- Insecure random number generation

**Configuration:**
- Format: SARIF
- No-fail mode: Enabled (doesn't fail workflow)
- Scan path: `./...` (all Go packages)

**Where to View:**
- GitHub → Security → Code scanning alerts
- Filter by: `gosec`

**What to Do:**
1. Fix high-priority Go security issues
2. Review code patterns flagged by Gosec
3. Add `#nosec` comments for false positives (with justification)
4. Update code to use secure alternatives

**Example Findings:**
- "G101: Potential hardcoded credentials"
- "G402: TLS InsecureSkipVerify set to true"
- "G104: Errors unhandled"
- "G501: Import blocklist: crypto/md5"

---

### 5. Grype Vulnerability Scan

**Type:** SCA (Software Composition Analysis)  
**Coverage:** Comprehensive vulnerability detection  
**Tool:** Anchore Grype

**What It Finds:**
- CVEs from multiple vulnerability databases
- Vulnerabilities in all dependency types
- OS package vulnerabilities
- Application-level vulnerabilities
- Comprehensive vulnerability matching

**Configuration:**
- Fail on: `high` severity vulnerabilities
- Output: JSON (for analysis)
- Scan target: Directory (`.`)

**Where to View:**
- Workflow logs (JSON output saved to `grype-results.json`)
- Note: Results are logged but not uploaded to GitHub Security tab (informational)

**What to Do:**
1. Review workflow logs for vulnerability summary
2. Check JSON output for detailed findings
3. Update dependencies based on findings
4. Cross-reference with other scanners

**Example Findings:**
- "CVE-2024-XXXXX: High severity in package Y"
- "Multiple vulnerabilities found in dependency Z"
- "Outdated package with known security issues"

---

### 6. Container Image Vulnerability Scan

**Type:** Image Scanning  
**Coverage:** Runtime container vulnerabilities  
**Tool:** Aqua Security Trivy (container mode)

**What It Finds:**
- Vulnerabilities in container base images
- OS package vulnerabilities in container
- Application dependencies in container
- Runtime security issues
- Outdated packages in container layers

**Configuration:**
- Trigger: Only on push to `master` (not PRs)
- Severity: `CRITICAL`, `HIGH`
- Format: SARIF
- Image: Built from Dockerfile during scan

**Where to View:**
- GitHub → Security → Code scanning alerts
- Filter by: `trivy-container`

**What to Do:**
1. Review container-specific vulnerabilities
2. Update base image to patched version
3. Update application dependencies in container
4. Consider using distroless or minimal base images
5. Keep base images updated regularly

**Example Findings:**
- "CVE-2024-XXXXX in base image alpine:3.18"
- "High severity vulnerability in container OS packages"
- "Outdated package in container layer"

---

### 7. License Compliance Check

**Type:** Compliance  
**Coverage:** Dependency license analysis  
**Tool:** Google License Classifier

**What It Finds:**
- License types of all dependencies
- License compatibility issues
- Potentially problematic licenses (GPL, AGPL, etc.)
- License compliance status

**Configuration:**
- Mode: Informational (non-blocking)
- Continue on error: Yes
- Output: Console (lists dependencies)

**Where to View:**
- Workflow logs (informational output)
- Not uploaded to Security tab (compliance check only)

**What to Do:**
1. Review dependency licenses in workflow logs
2. Ensure license compatibility with your project
3. Document license decisions
4. Consider alternatives for incompatible licenses

**Example Findings:**
- "Dependency X uses GPL license (may require disclosure)"
- "All dependencies use permissive licenses (MIT, Apache)"
- "License compatibility check passed"

---

## Where to View All Results

### GitHub Security Tab

**Location:** `https://github.com/{owner}/{repo}/security`

**Navigation:**
1. Go to your repository
2. Click **Security** tab
3. Click **Code scanning alerts**

**Features:**
- ✅ All SARIF results in one place
- ✅ Filter by scanner (CodeQL, Trivy, Gosec, etc.)
- ✅ Filter by severity (Critical, High, Medium, Low)
- ✅ Filter by status (Open, Dismissed, Fixed)
- ✅ View alert details and remediation steps
- ✅ Dismiss false positives with reason
- ✅ Track alert history

### Workflow Logs

**Location:** `https://github.com/{owner}/{repo}/actions/runs/{run-id}`

**What You Get:**
- Real-time scan execution logs
- Detailed error messages
- Scan summaries
- Grype JSON output (downloadable)
- License check output

---

## Understanding Scan Results

### Severity Levels

| Severity | Meaning | Action Required |
|----------|---------|-----------------|
| **Critical** | Immediate security risk | Fix immediately, block merge |
| **High** | Significant security risk | Fix before production |
| **Medium** | Moderate security risk | Fix in next sprint |
| **Low** | Minor security risk | Fix when convenient |
| **Info** | Informational only | Review and document |

### Result Types

**SAST (Static Analysis):**
- Finds issues in source code
- No runtime execution needed
- Fast scanning
- May have false positives

**SCA (Software Composition Analysis):**
- Finds vulnerabilities in dependencies
- Based on known CVE databases
- Requires dependency updates
- Generally accurate

**Image Scanning:**
- Finds vulnerabilities in container images
- Includes base image and application layers
- Critical for production deployments
- Requires image rebuilds

---

## What You Get: Summary

### ✅ Comprehensive Coverage

| Aspect | Coverage |
|--------|----------|
| **Source Code** | ✅ CodeQL + Gosec |
| **Dependencies** | ✅ Trivy + Dependency-Check + Grype |
| **Container Images** | ✅ Trivy Container Scan |
| **Licenses** | ✅ License Compliance Check |
| **Code Quality** | ✅ CodeQL quality queries |

### ✅ Centralized Viewing

- All results in GitHub Security tab
- Unified alert management
- Historical tracking
- False positive management

### ✅ Automated Scanning

- Runs on every push/PR
- Daily scheduled scans
- No manual intervention needed
- Integrated into CI/CD pipeline

### ✅ Actionable Results

- Clear severity levels
- Remediation guidance
- Direct links to CVE databases
- Code location references

---

## Interpreting Results

### High Priority (Fix Immediately)

- **Critical CVEs** in dependencies
- **SQL injection** vulnerabilities
- **Hardcoded secrets** in code
- **Insecure TLS** configurations
- **Critical container** vulnerabilities

### Medium Priority (Fix Soon)

- **High CVEs** in dependencies
- **Code quality** issues
- **Outdated dependencies**
- **Medium container** vulnerabilities

### Low Priority (Fix When Convenient)

- **Low severity** CVEs
- **Code style** issues
- **Informational** findings
- **License** compatibility notes

---

## Common Actions

### 1. Update Vulnerable Dependencies

```bash
# Update specific package
go get -u github.com/vulnerable/package@latest

# Update all dependencies
go get -u ./...

# Update to specific version (if patch available)
go get github.com/vulnerable/package@v1.2.4
```

### 2. Fix Code Issues

- Review CodeQL/Gosec alerts
- Fix security patterns
- Add input validation
- Use secure alternatives

### 3. Update Container Base Images

```dockerfile
# Update base image in Dockerfile
FROM golang:1.22-alpine3.19  # Use latest patched version
```

### 4. Dismiss False Positives

1. Go to Security → Code scanning alerts
2. Click on alert
3. Click "Dismiss"
4. Select reason (false positive, won't fix, etc.)
5. Add comment explaining why

---

## Workflow Status Indicators

| Status | Meaning |
|--------|---------|
| ✅ **Success** | All scans completed, results uploaded |
| ⚠️ **Warning** | Some scans found issues (non-blocking) |
| ❌ **Failure** | Scan failed (check logs for errors) |
| ⊘ **Skipped** | Condition not met (e.g., container scan on PR) |

---

## Troubleshooting

### Scan Not Appearing in Security Tab

**Possible Causes:**
- SARIF file not generated
- Upload step failed
- Missing `security-events: write` permission

**Solution:**
- Check workflow logs for upload errors
- Verify permissions in workflow file
- Ensure scanner generated SARIF output

### Too Many False Positives

**Solution:**
- Dismiss false positives in Security tab
- Add suppressions (e.g., `suppression.xml` for Dependency-Check)
- Use `#nosec` comments for Gosec (with justification)
- Configure scanner-specific ignore patterns

### Scan Taking Too Long

**Solution:**
- Scans run in parallel (should be fast)
- CodeQL can take 2-5 minutes (normal)
- Container scan only runs on master push
- Consider excluding large directories

---

## Best Practices

### 1. Review Alerts Regularly

- Check Security tab weekly
- Fix critical/high issues immediately
- Plan medium/low fixes in sprints

### 2. Keep Dependencies Updated

- Run `go get -u ./...` regularly
- Review Dependabot PRs
- Update base images monthly

### 3. Fix Root Causes

- Don't just suppress alerts
- Fix insecure code patterns
- Use secure coding practices

### 4. Document Decisions

- Document why false positives are dismissed
- Explain why certain vulnerabilities are accepted
- Track remediation progress

---

## Summary

**You Get:**
- ✅ 8 comprehensive security scanners
- ✅ Automated scanning on every change
- ✅ Centralized results in GitHub Security tab
- ✅ Actionable alerts with severity levels
- ✅ Coverage of code, dependencies, and containers
- ✅ License compliance checking

**What to Do:**
1. Review Security → Code scanning alerts regularly
2. Fix critical/high issues immediately
3. Update dependencies regularly
4. Dismiss false positives with reason
5. Keep base images updated

**Result:**
- Strong security posture
- Early vulnerability detection
- Compliance with security best practices
- Production-ready, secure codebase

---

## Related Documentation

- [Security Policy](./SECURITY.md) - Overall security model
- [K8s Security Runbook](./K8S_SECURITY_RUNBOOK.md) - Kubernetes security
- [CI/CD Template](../03-CI_CD/CI_CD_TEMPLATE.md) - Workflow setup
- [Pipeline Summary](../03-CI_CD/PIPELINE_SUMMARY.md) - Complete pipeline overview
