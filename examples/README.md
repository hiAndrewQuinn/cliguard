# CLIGuard CI/CD Integration Examples

This directory contains example configurations for integrating CLIGuard into various CI/CD platforms.

## Available Examples

### 📁 [`circleci/`](./circleci/)
CircleCI configuration example with validation job and workflow setup.

### 📁 [`docker/`](./docker/)
Docker-based validation approach for containerized environments.

### 📁 [`gitlab/`](./gitlab/)
GitLab CI pipeline configuration for CLIGuard validation.

### 📁 [`jenkins/`](./jenkins/)
Jenkins Pipeline (Jenkinsfile) example with email notifications.

### 📁 [`make/`](./make/)
Makefile integration for local development and CI pipelines.

## Quick Start

1. **Choose your CI/CD platform** from the directories above
2. **Copy the example configuration** to your project
3. **Customize the configuration**:
   - Update the `entrypoint` to match your CLI's entry function
   - Adjust paths if your project structure differs
   - Configure any platform-specific settings

## GitHub Actions Examples

For GitHub Actions, see the `.github/workflows/` directory in the root of this repository:
- [`cliguard-validate.yml`](../.github/workflows/cliguard-validate.yml) - PR validation workflow
- [`cliguard-generate.yml`](../.github/workflows/cliguard-generate.yml) - Automatic contract generation

## Common Configuration Points

All examples need these key configurations:

| Setting | Description | Example |
|---------|-------------|---------|
| `entrypoint` | Your CLI's main entry function | `github.com/org/repo/cmd.NewRootCmd` |
| `project-path` | Path to your Go project | `.` (current directory) |
| `contract` | Path to contract file | `cliguard.yaml` |

## Documentation

For comprehensive documentation on CI/CD integration strategies, best practices, and troubleshooting, see the [CI/CD Integration Guide](../docs/ci-cd-integration.md).

## Platform Support

| Platform | Validation | Generation | PR Comments | Auto-PR |
|----------|------------|------------|-------------|---------|
| GitHub Actions | ✅ | ✅ | ✅ | ✅ |
| GitLab CI | ✅ | ✅ | ✅ | ✅ |
| CircleCI | ✅ | ✅ | ⚠️ | ⚠️ |
| Jenkins | ✅ | ✅ | ✅ | ⚠️ |
| Docker | ✅ | ✅ | N/A | N/A |
| Makefile | ✅ | ✅ | N/A | N/A |

✅ = Fully supported with examples
⚠️ = Requires additional configuration
N/A = Not applicable

## Contributing

If you have examples for other CI/CD platforms or improvements to existing examples, please submit a pull request!