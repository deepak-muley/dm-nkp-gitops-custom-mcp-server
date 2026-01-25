# Documentation Structure

Complete documentation organized by topic area. Each folder contains related guides, references, and best practices.

## Quick Navigation

### 🔌 [01-MCP/](./01-MCP/) - MCP Server & Protocol
Learn about the Model Context Protocol server and how to use it with AI assistants.

- **[MCP_PRIMER.md](./01-MCP/MCP_PRIMER.md)** - Introduction to Model Context Protocol
- **[MCP_SERVER_ARCHITECTURE.md](./01-MCP/MCP_SERVER_ARCHITECTURE.md)** - How the MCP server is designed and structured
- **[TOOLS_REFERENCE.md](./01-MCP/TOOLS_REFERENCE.md)** - Reference for all available MCP tools
- **[FLUX_MCP_SETUP.md](./01-MCP/FLUX_MCP_SETUP.md)** - Setting up MCP with Flux CD

---

### 🤖 [02-A2A/](./02-A2A/) - Agent-to-Agent Communication
Multi-agent orchestration and agent collaboration patterns.

- **[A2A_LEARNING_GUIDE.md](./02-A2A/A2A_LEARNING_GUIDE.md)** - Learning guide for A2A framework
- **[A2A_PROTOCOL.md](./02-A2A/A2A_PROTOCOL.md)** - A2A protocol specification and design

---

### 🚀 [03-CI_CD/](./03-CI_CD/) - CI/CD Pipelines & Workflows
Complete CI/CD setup, best practices, and troubleshooting.

**Start Here:**
- **[CI_CD_TEMPLATE.md](./03-CI_CD/CI_CD_TEMPLATE.md)** - Ready-to-use CI/CD template for new projects
- **[PIPELINE_SUMMARY.md](./03-CI_CD/PIPELINE_SUMMARY.md)** - Overview of the complete pipeline

**Deep Dive:**
- **[CI_CD_ARCHITECTURE.md](./03-CI_CD/CI_CD_ARCHITECTURE.md)** - Architectural design and decision rationale
- **[CI_CD_BEST_PRACTICES.md](./03-CI_CD/CI_CD_BEST_PRACTICES.md)** - Best practices and patterns
- **[CI_CD_CONSISTENCY.md](./03-CI_CD/CI_CD_CONSISTENCY.md)** - Consistency checklist across repos
- **[GITHUB_RUNNERS.md](./03-CI_CD/GITHUB_RUNNERS.md)** - Understanding GitHub Runners (concept and usage)

**Alignment & Verification:**
- **[CI_CD_DIFFERENCES_ANALYSIS.md](./03-CI_CD/CI_CD_DIFFERENCES_ANALYSIS.md)** - Compare workflows between repos
- **[FINAL_ALIGNMENT_REPORT.md](./03-CI_CD/FINAL_ALIGNMENT_REPORT.md)** - Current alignment status
- **[COMPLETE_ALIGNMENT.md](./03-CI_CD/COMPLETE_ALIGNMENT.md)** - Detailed alignment documentation
- **[GITHUB_CONSISTENCY_FINAL.md](./03-CI_CD/GITHUB_CONSISTENCY_FINAL.md)** - GitHub configuration consistency
- **[WORKFLOW_VERIFICATION.md](./03-CI_CD/WORKFLOW_VERIFICATION.md)** - Verify your CI/CD setup
- **[REPOS_ALIGNMENT.md](./03-CI_CD/REPOS_ALIGNMENT.md)** - Alignment between repositories
- **[WHY_THOSE_FILES_WERE_ADDED.md](./03-CI_CD/WHY_THOSE_FILES_WERE_ADDED.md)** - History of CI/CD file additions

---

### 📦 [04-CATALOG_UPDATE/](./04-CATALOG_UPDATE/) - Application Catalog Updates
Automatically update external GitOps catalogs with new artifacts.

**Quick Start:**
- **[CATALOG_UPDATE_QUICK_START.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md)** - Get started in 5 minutes

**Setup & Configuration:**
- **[CATALOG_UPDATE_SETUP.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_SETUP.md)** - Step-by-step setup guide
- **[CATALOG_UPDATE_DESIGN.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_DESIGN.md)** - Design and architecture

**Security & Authentication:**
- **[CATALOG_UPDATE_SECURITY.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_SECURITY.md)** - Security best practices
- **[CATALOG_UPDATE_SECRETS.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_SECRETS.md)** - Managing secrets and tokens

---

### 🚢 [05-DEPLOYMENT/](./05-DEPLOYMENT/) - Deployment & Operations
Production deployment guides and operational procedures.

- **[DEPLOYMENT_AND_USAGE.md](./05-DEPLOYMENT/DEPLOYMENT_AND_USAGE.md)** - How to deploy and use the server
- **[ENTERPRISE_DEPLOYMENT.md](./05-DEPLOYMENT/ENTERPRISE_DEPLOYMENT.md)** - Enterprise-grade deployment
- **[NKP_PRODUCTION_DEPLOYMENT.md](./05-DEPLOYMENT/NKP_PRODUCTION_DEPLOYMENT.md)** - NKP-specific production setup

---

### 🔐 [06-SECURITY/](./06-SECURITY/) - Security & Compliance
Security policies, threat models, and compliance.

- **[SECURITY.md](./06-SECURITY/SECURITY.md)** - Security policies and vulnerability reporting
- **[SECURITY_SCANNING_WORKFLOW.md](./06-SECURITY/SECURITY_SCANNING_WORKFLOW.md)** - Complete guide to security.yaml workflow and scanners
- **[K8S_SECURITY_RUNBOOK.md](./06-SECURITY/K8S_SECURITY_RUNBOOK.md)** - Kubernetes security best practices

---

### 🔧 [07-TROUBLESHOOTING/](./07-TROUBLESHOOTING/) - Debugging & Troubleshooting
Troubleshooting guides and testing procedures.

- **[K8S_TROUBLESHOOTING_RUNBOOK.md](./07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md)** - Debug Kubernetes issues
- **[TESTING_GUIDE.md](./07-TROUBLESHOOTING/TESTING_GUIDE.md)** - Testing strategies and procedures

---

### 📚 [08-LEARNING/](./08-LEARNING/) - Learning Resources
Learning guides and comparative analysis.

- **[KMCP_LEARNING_GUIDE.md](./08-LEARNING/KMCP_LEARNING_GUIDE.md)** - Learning guide for Kubernetes MCP
- **[FRAMEWORK_COMPARISON.md](./08-LEARNING/FRAMEWORK_COMPARISON.md)** - Compare different frameworks and patterns
- **[RUNBOOK_BEST_PRACTICES.md](./08-LEARNING/RUNBOOK_BEST_PRACTICES.md)** - Best practices for runbooks

---

### 📋 [09-GOVERNANCE/](./09-GOVERNANCE/) - Repository Governance
Governance automation and best practices.

- **[GOVERNANCE.md](./09-GOVERNANCE/GOVERNANCE.md)** - Repository governance setup (CODEOWNERS, Dependabot, PR templates)

---

## Documentation by Use Case

### I'm New Here
1. Start with [01-MCP/MCP_PRIMER.md](./01-MCP/MCP_PRIMER.md) - Understand what this is
2. Read [03-CI_CD/PIPELINE_SUMMARY.md](./03-CI_CD/PIPELINE_SUMMARY.md) - See the big picture
3. Explore [AGENTS.md](../AGENTS.md) - Project overview from root

### I'm Setting Up CI/CD
1. [03-CI_CD/CI_CD_TEMPLATE.md](./03-CI_CD/CI_CD_TEMPLATE.md) - Copy the template
2. [03-CI_CD/CI_CD_BEST_PRACTICES.md](./03-CI_CD/CI_CD_BEST_PRACTICES.md) - Learn best practices
3. [03-CI_CD/WORKFLOW_VERIFICATION.md](./03-CI_CD/WORKFLOW_VERIFICATION.md) - Verify your setup

### I'm Deploying to Production
1. [05-DEPLOYMENT/ENTERPRISE_DEPLOYMENT.md](./05-DEPLOYMENT/ENTERPRISE_DEPLOYMENT.md) - Enterprise setup
2. [06-SECURITY/SECURITY.md](./06-SECURITY/SECURITY.md) - Security checklist
3. [05-DEPLOYMENT/NKP_PRODUCTION_DEPLOYMENT.md](./05-DEPLOYMENT/NKP_PRODUCTION_DEPLOYMENT.md) - NKP specifics

### I'm Debugging Issues
1. [07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md](./07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md) - Debug Kubernetes
2. [03-CI_CD/WORKFLOW_VERIFICATION.md](./03-CI_CD/WORKFLOW_VERIFICATION.md) - Verify workflows
3. Check [01-MCP/TOOLS_REFERENCE.md](./01-MCP/TOOLS_REFERENCE.md) - Find the right tool

### I'm Managing the Catalog
1. [04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md) - Get started
2. [04-CATALOG_UPDATE/CATALOG_UPDATE_SECURITY.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_SECURITY.md) - Secure it
3. [04-CATALOG_UPDATE/CATALOG_UPDATE_SETUP.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_SETUP.md) - Full setup

### I'm Creating a New Project
1. [03-CI_CD/CI_CD_TEMPLATE.md](./03-CI_CD/CI_CD_TEMPLATE.md) - Copy CI/CD template
2. [09-GOVERNANCE/GOVERNANCE.md](./09-GOVERNANCE/GOVERNANCE.md) - Copy governance files
3. [08-LEARNING/RUNBOOK_BEST_PRACTICES.md](./08-LEARNING/RUNBOOK_BEST_PRACTICES.md) - Follow best practices

---

## Folder Organization

```
docs/
├── 01-MCP/              ← Model Context Protocol (core technology)
├── 02-A2A/              ← Agent-to-Agent communication
├── 03-CI_CD/            ← Continuous Integration/Deployment (largest section)
├── 04-CATALOG_UPDATE/   ← GitOps catalog automation
├── 05-DEPLOYMENT/       ← Production deployment
├── 06-SECURITY/         ← Security & compliance
├── 07-TROUBLESHOOTING/  ← Debugging & testing
├── 08-LEARNING/         ← Educational resources
└── 09-GOVERNANCE/       ← Repository governance
```

---

## Key Documents by Category

### Architecture & Design
- MCP Server: [01-MCP/MCP_SERVER_ARCHITECTURE.md](./01-MCP/MCP_SERVER_ARCHITECTURE.md)
- CI/CD: [03-CI_CD/CI_CD_ARCHITECTURE.md](./03-CI_CD/CI_CD_ARCHITECTURE.md)
- Catalog: [04-CATALOG_UPDATE/CATALOG_UPDATE_DESIGN.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_DESIGN.md)

### Quick Starts
- CI/CD: [03-CI_CD/CI_CD_TEMPLATE.md](./03-CI_CD/CI_CD_TEMPLATE.md)
- Catalog: [04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md)

### Best Practices
- CI/CD: [03-CI_CD/CI_CD_BEST_PRACTICES.md](./03-CI_CD/CI_CD_BEST_PRACTICES.md)
- Runbooks: [08-LEARNING/RUNBOOK_BEST_PRACTICES.md](./08-LEARNING/RUNBOOK_BEST_PRACTICES.md)
- Security: [06-SECURITY/K8S_SECURITY_RUNBOOK.md](./06-SECURITY/K8S_SECURITY_RUNBOOK.md)

### Verification & Troubleshooting
- Workflows: [03-CI_CD/WORKFLOW_VERIFICATION.md](./03-CI_CD/WORKFLOW_VERIFICATION.md)
- Kubernetes: [07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md](./07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md)

### References
- MCP Tools: [01-MCP/TOOLS_REFERENCE.md](./01-MCP/TOOLS_REFERENCE.md)
- A2A Protocol: [02-A2A/A2A_PROTOCOL.md](./02-A2A/A2A_PROTOCOL.md)

---

## Contributing

When adding new documentation:
1. Determine the appropriate folder based on topic
2. Use clear, descriptive names (e.g., `FEATURE_TOPIC.md`)
3. Add cross-references in this README
4. Keep files focused and well-indexed

---

## Documentation Stats

- **Total Docs:** 34 files
- **Organized into:** 9 categories
- **Largest Category:** CI/CD (12 files)
- **Last Updated:** 2026-01-24
- **Status:** Complete and organized ✅

---

## Quick Search

**Looking for X? Try these:**

| Need | File |
|------|------|
| Protocol details | [01-MCP/MCP_PRIMER.md](./01-MCP/MCP_PRIMER.md) |
| Available tools | [01-MCP/TOOLS_REFERENCE.md](./01-MCP/TOOLS_REFERENCE.md) |
| Agent communication | [02-A2A/A2A_PROTOCOL.md](./02-A2A/A2A_PROTOCOL.md) |
| CI/CD setup | [03-CI_CD/CI_CD_TEMPLATE.md](./03-CI_CD/CI_CD_TEMPLATE.md) |
| Catalog updates | [04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md](./04-CATALOG_UPDATE/CATALOG_UPDATE_QUICK_START.md) |
| Production setup | [05-DEPLOYMENT/ENTERPRISE_DEPLOYMENT.md](./05-DEPLOYMENT/ENTERPRISE_DEPLOYMENT.md) |
| Security policies | [06-SECURITY/SECURITY.md](./06-SECURITY/SECURITY.md) |
| Security scanning | [06-SECURITY/SECURITY_SCANNING_WORKFLOW.md](./06-SECURITY/SECURITY_SCANNING_WORKFLOW.md) |
| Debugging help | [07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md](./07-TROUBLESHOOTING/K8S_TROUBLESHOOTING_RUNBOOK.md) |
| Best practices | [08-LEARNING/RUNBOOK_BEST_PRACTICES.md](./08-LEARNING/RUNBOOK_BEST_PRACTICES.md) |
| Governance | [09-GOVERNANCE/GOVERNANCE.md](./09-GOVERNANCE/GOVERNANCE.md) |
