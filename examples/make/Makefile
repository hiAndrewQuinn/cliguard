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