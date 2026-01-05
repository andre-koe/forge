# Security Policy

**Forge is in early development.** Security issues will be addressed on a best-effort basis.

## Reporting a Vulnerability

Found a security issue? Please **do not** create a public GitHub issue.

**Report privately via:**
- **Email:** andre.koeniger1997@gmail.com
- **GitHub:** Use "Security" → "Advisories" → "Report a vulnerability"

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Your contact information

I'll respond as soon as possible and coordinate disclosure after a fix is available.

## Important: Workflow Execution Risks

Forge executes shell commands from YAML workflow files:

- **Arbitrary Command Execution:** Workflows can run any command with your user permissions
- **No Sandboxing:** Commands are not isolated or restricted
- **File System Access:** Workflows can read/write any files you have access to

**Before running a workflow:**
1. Review the YAML file carefully
2. Check with `forge dry-run` first
3. Only run workflows from trusted sources

## Current Limitations

- No built-in secrets management (use environment variables)
- No execution sandboxing
- Commands inherit Forge's permissions

---

**Security is your responsibility when running workflows. Review before execution.**
