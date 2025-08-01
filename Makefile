.PHONY: test test-fixtures test-integration clean-fixtures

# Run all tests
test:
	go test -v ./...

# Run only integration tests
test-integration:
	go test -v ./cmd -run TestIntegration

# Setup test fixtures
test-fixtures:
	@echo "Setting up test fixtures..."
	cd test/fixtures/simple-cli && go mod tidy
	cd test/fixtures/simple-cli && cliguard generate --project-path . --entrypoint github.com/test/simple-cli/cmd.NewRootCmd > cliguard.yaml
	@echo "Test fixtures ready"

# Clean and regenerate test fixtures
clean-fixtures:
	@echo "Cleaning test fixtures..."
	rm -f test/fixtures/simple-cli/cliguard.yaml
	$(MAKE) test-fixtures