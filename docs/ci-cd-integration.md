# CI/CD Integration Guide for CLIGuard

This guide provides comprehensive examples and best practices for integrating CLIGuard into your continuous integration and deployment pipelines.

## Table of Contents

- [Overview](#overview)
- [When to Validate](#when-to-validate)
- [Platform-Specific Examples](#platform-specific-examples)
  - [GitHub Actions](#github-actions)
  - [GitLab CI](#gitlab-ci)
  - [CircleCI](#circleci)
  - [Jenkins](#jenkins)
- [Docker-Based Validation](#docker-based-validation)
- [Makefile Integration](#makefile-integration)
- [Breaking Change Workflows](#breaking-change-workflows)
- [Multi-Module Repositories](#multi-module-repositories)
- [Best Practices](#best-practices)

## Overview

CLIGuard validation in CI/CD pipelines helps you:
- Prevent unintentional breaking changes to your CLI
- Enforce CLI structure consistency across teams
- Automatically document CLI changes
- Provide early feedback on pull requests

## When to Validate

### Pull Request Validation (Recommended)

Run validation on every pull request that modifies:
- Go source files (`*.go`)
- Module dependencies (`go.mod`, `go.sum`)
- The CLI contract file (`cliguard.yaml`)

This provides immediate feedback to developers about potential breaking changes.

### Pre-Release Validation

Include validation in your release pipeline to ensure:
- The contract matches the code being released
- No breaking changes slip through PR reviews
- The contract file is up-to-date

### Continuous Contract Updates

Automatically generate and update contracts when:
- Code is merged to the main branch
- A new release is tagged
- Manual workflow dispatch is triggered

## Platform-Specific Examples

### GitHub Actions

#### Validation Workflow

The validation workflow runs on pull requests and blocks merging if the CLI contract is violated:

```yaml
name: Validate CLI Contract

on:
  pull_request:
    paths:
      - '**/*.go'
      - 'go.mod'
      - 'go.sum'
      - 'cliguard.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Install cliguard
        run: go install github.com/hiAndrewQuinn/cliguard@latest
        
      - name: Validate CLI structure
        run: |
          cliguard validate \
            --project-path . \
            --entrypoint "github.com/${{ github.repository }}/cmd.NewRootCmd"
            
      - name: Comment on PR if validation fails
        if: failure()
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '❌ CLI contract validation failed! This PR introduces breaking changes to the CLI structure. Please review the changes or update the contract file.'
            })
```

#### Contract Generation Workflow

This workflow automatically creates a PR when the CLI structure changes:

```yaml
name: Update CLI Contract

on:
  workflow_dispatch:
  push:
    branches:
      - main
    paths:
      - 'cmd/**'
      - 'internal/cmd/**'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Install cliguard
        run: go install github.com/hiAndrewQuinn/cliguard@latest
        
      - name: Generate contract
        run: |
          cliguard generate \
            --project-path . \
            --entrypoint "github.com/${{ github.repository }}/cmd.NewRootCmd" \
            > cliguard.yaml.new
            
      - name: Check for changes
        id: diff
        run: |
          if ! diff -q cliguard.yaml cliguard.yaml.new; then
            echo "changed=true" >> $GITHUB_OUTPUT
          fi
          
      - name: Create PR
        if: steps.diff.outputs.changed == 'true'
        uses: peter-evans/create-pull-request@v5
        with:
          title: 'Update CLI contract'
          body: 'This PR updates the CLI contract to match the current implementation.'
          commit-message: 'chore: update CLI contract'
          branch: update-cli-contract
```

### GitLab CI

Add to your `.gitlab-ci.yml`:

```yaml
stages:
  - validate

validate-cli:
  stage: validate
  image: golang:1.21
  before_script:
    - go install github.com/hiAndrewQuinn/cliguard@latest
  script:
    - cliguard validate --project-path . --entrypoint "gitlab.com/$CI_PROJECT_PATH/cmd.NewRootCmd"
  only:
    changes:
      - "**/*.go"
      - go.mod
      - go.sum
      - cliguard.yaml
```

For automatic contract updates, add:

```yaml
generate-contract:
  stage: deploy
  image: golang:1.21
  before_script:
    - go install github.com/hiAndrewQuinn/cliguard@latest
  script:
    - |
      cliguard generate \
        --project-path . \
        --entrypoint "gitlab.com/$CI_PROJECT_PATH/cmd.NewRootCmd" \
        > cliguard.yaml.new
    - |
      if ! diff -q cliguard.yaml cliguard.yaml.new; then
        mv cliguard.yaml.new cliguard.yaml
        git config user.email "ci@example.com"
        git config user.name "GitLab CI"
        git add cliguard.yaml
        git commit -m "chore: update CLI contract"
        git push "https://${CI_JOB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git" HEAD:${CI_COMMIT_BRANCH}
      fi
  only:
    - main
```

### CircleCI

Add to `.circleci/config.yml`:

```yaml
version: 2.1

jobs:
  validate-cli:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run:
          name: Install cliguard
          command: go install github.com/hiAndrewQuinn/cliguard@latest
      - run:
          name: Validate CLI contract
          command: |
            cliguard validate \
              --project-path . \
              --entrypoint "github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}/cmd.NewRootCmd"

workflows:
  version: 2
  validate:
    jobs:
      - validate-cli
```

For advanced workflows with contract generation:

```yaml
version: 2.1

orbs:
  github-cli: circleci/github-cli@2.0

jobs:
  validate-cli:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run:
          name: Install cliguard
          command: go install github.com/hiAndrewQuinn/cliguard@latest
      - run:
          name: Validate CLI contract
          command: |
            cliguard validate \
              --project-path . \
              --entrypoint "github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}/cmd.NewRootCmd"

  generate-contract:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run:
          name: Install cliguard
          command: go install github.com/hiAndrewQuinn/cliguard@latest
      - run:
          name: Generate new contract
          command: |
            cliguard generate \
              --project-path . \
              --entrypoint "github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}/cmd.NewRootCmd" \
              > cliguard.yaml.new
      - run:
          name: Check for changes
          command: |
            if ! diff -q cliguard.yaml cliguard.yaml.new; then
              echo "Contract needs updating"
              mv cliguard.yaml.new cliguard.yaml
            fi
      - github-cli/install
      - run:
          name: Create PR if needed
          command: |
            if git diff --exit-code cliguard.yaml; then
              echo "No changes needed"
            else
              git config user.email "ci@example.com"
              git config user.name "CircleCI"
              git checkout -b update-cli-contract
              git add cliguard.yaml
              git commit -m "chore: update CLI contract"
              git push origin update-cli-contract
              gh pr create --title "Update CLI contract" --body "Automated contract update"
            fi

workflows:
  version: 2
  validate-and-generate:
    jobs:
      - validate-cli
      - generate-contract:
          requires:
            - validate-cli
          filters:
            branches:
              only: main
```

### Jenkins

For Jenkins Pipeline (Jenkinsfile):

```groovy
pipeline {
    agent {
        docker {
            image 'golang:1.21'
        }
    }
    
    stages {
        stage('Validate CLI') {
            steps {
                sh 'go install github.com/hiAndrewQuinn/cliguard@latest'
                sh 'cliguard validate --project-path . --entrypoint "github.com/org/repo/cmd.NewRootCmd"'
            }
        }
    }
    
    post {
        failure {
            emailext (
                subject: "CLI Contract Validation Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}",
                body: "The CLI contract validation failed. This indicates breaking changes to the CLI structure.",
                to: "${env.CHANGE_AUTHOR_EMAIL}"
            )
        }
    }
}
```

For multibranch pipelines with contract generation:

```groovy
pipeline {
    agent {
        docker {
            image 'golang:1.21'
        }
    }
    
    environment {
        GITHUB_TOKEN = credentials('github-token')
    }
    
    stages {
        stage('Validate CLI') {
            when {
                changeRequest()
            }
            steps {
                sh 'go install github.com/hiAndrewQuinn/cliguard@latest'
                sh 'cliguard validate --project-path . --entrypoint "github.com/org/repo/cmd.NewRootCmd"'
            }
        }
        
        stage('Generate Contract') {
            when {
                branch 'main'
            }
            steps {
                sh 'go install github.com/hiAndrewQuinn/cliguard@latest'
                sh '''
                    cliguard generate \
                        --project-path . \
                        --entrypoint "github.com/org/repo/cmd.NewRootCmd" \
                        > cliguard.yaml.new
                '''
                script {
                    def changed = sh(
                        script: 'diff -q cliguard.yaml cliguard.yaml.new',
                        returnStatus: true
                    ) != 0
                    
                    if (changed) {
                        sh '''
                            mv cliguard.yaml.new cliguard.yaml
                            git config user.email "jenkins@example.com"
                            git config user.name "Jenkins CI"
                            git checkout -b update-cli-contract
                            git add cliguard.yaml
                            git commit -m "chore: update CLI contract"
                            git push https://${GITHUB_TOKEN}@github.com/org/repo.git update-cli-contract
                        '''
                        
                        sh '''
                            curl -X POST \
                                -H "Authorization: token ${GITHUB_TOKEN}" \
                                -H "Accept: application/vnd.github.v3+json" \
                                https://api.github.com/repos/org/repo/pulls \
                                -d '{"title":"Update CLI contract","head":"update-cli-contract","base":"main"}'
                        '''
                    }
                }
            }
        }
    }
}
```

## Docker-Based Validation

For environments where Go installation is not desired, use Docker:

### Dockerfile.validate

```dockerfile
FROM golang:1.21-alpine AS validator

RUN go install github.com/hiAndrewQuinn/cliguard@latest

WORKDIR /app
COPY . .

ENTRYPOINT ["cliguard", "validate", "--project-path", ".", "--entrypoint", "cmd.NewRootCmd"]
```

### Usage in CI

```bash
# Build the validation image
docker build -f Dockerfile.validate -t cli-validator .

# Run validation
docker run --rm cli-validator

# Or with custom entrypoint
docker run --rm cli-validator cliguard validate --project-path . --entrypoint "github.com/org/repo/cmd.NewRootCmd"
```

### Docker Compose Integration

```yaml
version: '3.8'

services:
  validate:
    build:
      context: .
      dockerfile: Dockerfile.validate
    volumes:
      - .:/app
    command: validate --project-path /app --entrypoint "cmd.NewRootCmd"
```

## Makefile Integration

Add CLIGuard targets to your project's Makefile:

```makefile
CLIGUARD_VERSION := latest
ENTRYPOINT := github.com/org/repo/cmd.NewRootCmd

.PHONY: install-cliguard
install-cliguard:
	@go install github.com/hiAndrewQuinn/cliguard@$(CLIGUARD_VERSION)

.PHONY: validate-cli
validate-cli: install-cliguard
	@echo "Validating CLI contract..."
	@cliguard validate --project-path . --entrypoint $(ENTRYPOINT)

.PHONY: generate-contract
generate-contract: install-cliguard
	@echo "Generating CLI contract..."
	@cliguard generate --project-path . --entrypoint $(ENTRYPOINT) > cliguard.yaml

.PHONY: ci
ci: test validate-cli
```

Use in CI pipelines:

```yaml
# GitHub Actions
- name: Validate CLI
  run: make validate-cli

# GitLab CI
script:
  - make validate-cli

# Jenkins
sh 'make validate-cli'
```

## Breaking Change Workflows

### Strategy 1: Strict Validation (Default)

Block all breaking changes:

1. Validation fails on PR
2. Developer must either:
   - Revert the breaking change
   - Get approval and update the contract

### Strategy 2: Approved Breaking Changes

Allow breaking changes with approval:

```yaml
name: Validate CLI with Override

on:
  pull_request:
    types: [opened, synchronize, labeled]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Check for breaking change approval
        id: check-approval
        run: |
          if [[ "${{ contains(github.event.pull_request.labels.*.name, 'breaking-change-approved') }}" == "true" ]]; then
            echo "skip_validation=true" >> $GITHUB_OUTPUT
          fi
      
      - name: Validate CLI structure
        if: steps.check-approval.outputs.skip_validation != 'true'
        run: |
          cliguard validate \
            --project-path . \
            --entrypoint "github.com/${{ github.repository }}/cmd.NewRootCmd"
```

### Strategy 3: Version-Based Contracts

Maintain contracts for multiple versions:

```bash
# Directory structure
contracts/
├── v1.0.0/
│   └── cliguard.yaml
├── v2.0.0/
│   └── cliguard.yaml
└── latest/
    └── cliguard.yaml
```

Validation script:

```bash
#!/bin/bash
VERSION=${1:-latest}
cliguard validate \
  --project-path . \
  --contract "contracts/${VERSION}/cliguard.yaml" \
  --entrypoint "cmd.NewRootCmd"
```

## Multi-Module Repositories

For repositories with multiple CLI tools:

### Directory Structure

```
repo/
├── cmd/
│   ├── tool1/
│   │   └── main.go
│   └── tool2/
│       └── main.go
├── contracts/
│   ├── tool1.yaml
│   └── tool2.yaml
```

### Validation Workflow

```yaml
name: Validate Multiple CLIs

on:
  pull_request:
    paths:
      - 'cmd/**'
      - 'contracts/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        cli:
          - name: tool1
            entrypoint: "github.com/${{ github.repository }}/cmd/tool1.NewRootCmd"
            contract: contracts/tool1.yaml
          - name: tool2
            entrypoint: "github.com/${{ github.repository }}/cmd/tool2.NewRootCmd"
            contract: contracts/tool2.yaml
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Install cliguard
        run: go install github.com/hiAndrewQuinn/cliguard@latest
        
      - name: Validate ${{ matrix.cli.name }}
        run: |
          cliguard validate \
            --project-path . \
            --contract ${{ matrix.cli.contract }} \
            --entrypoint "${{ matrix.cli.entrypoint }}"
```

## Best Practices

### 1. Start Early

- Add CLIGuard validation when creating a new CLI project
- Generate the initial contract as part of project setup
- Include contract files in version control

### 2. Use Semantic Versioning

- Major version bumps for breaking changes
- Update contracts when releasing new major versions
- Maintain backward compatibility within minor/patch releases

### 3. Document Changes

When updating contracts, include:
- What changed in the CLI structure
- Why the change was necessary
- Migration instructions for users

Example PR description:

```markdown
## CLI Contract Update

### Changes
- Added `--format` flag to `list` command
- Renamed `--output` to `--out` for consistency
- Removed deprecated `legacy` subcommand

### Migration Guide
- Replace `--output` with `--out` in scripts
- Remove any usage of the `legacy` command
- Use `--format json` instead of `--json` flag
```

### 4. Monitor and Alert

- Set up notifications for validation failures
- Track contract changes in release notes
- Use dashboards to monitor CLI stability

### 5. Test Contract Generation

Regularly test that contract generation works:

```yaml
name: Test Contract Generation

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday

jobs:
  test-generation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Generate and validate contract
        run: |
          cliguard generate --project-path . --entrypoint "cmd.NewRootCmd" > test.yaml
          cliguard validate --project-path . --contract test.yaml --entrypoint "cmd.NewRootCmd"
```

### 6. Cache Dependencies

Speed up CI runs by caching Go modules and CLIGuard:

```yaml
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-

- name: Cache CLIGuard
  uses: actions/cache@v3
  with:
    path: ~/go/bin/cliguard
    key: ${{ runner.os }}-cliguard-${{ env.CLIGUARD_VERSION }}
```

## Troubleshooting

### Common Issues

1. **Validation fails with "entrypoint not found"**
   - Verify the entrypoint path matches your module name
   - Check that the function is exported (capitalized)
   - Ensure go.mod is present in the project root

2. **Contract generation produces empty file**
   - Verify the CLI can be built successfully
   - Check for compilation errors
   - Ensure all dependencies are available

3. **Validation passes locally but fails in CI**
   - Check for environment-specific differences
   - Ensure the same CLIGuard version is used
   - Verify file paths are correct for the CI environment

### Debug Mode

Enable verbose output for troubleshooting:

```bash
cliguard validate \
  --project-path . \
  --entrypoint "cmd.NewRootCmd" \
  --verbose
```

## Conclusion

Integrating CLIGuard into your CI/CD pipeline ensures CLI stability and provides early feedback on structural changes. Choose the integration approach that best fits your workflow and team practices. Start with basic validation and gradually adopt more advanced patterns as your CLI evolves.