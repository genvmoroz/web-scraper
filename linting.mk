OSNAME=$(shell go env GOOS)

TMP_DIR = /tmp
GOLANGCI_LINT_VERSION=1.41.1
GOLANGCI_DIR = $(TMP_DIR)/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_TMP_BIN = $(GOLANGCI_DIR)/golangci-lint

GOLANGCI_CMD = golangci-lint run --allow-parallel-runners -c .golangci.yml
GOLANGCI_LINT_ARCHIVE = golangci-lint-$(GOLANGCI_LINT_VERSION)-$(OSNAME)-amd64.tar.gz

# Run this on local machine.
# It downloads a version of golangci-lint and execute it locally.
.PHONY: lint
lint: $(GOLANGCI_TMP_BIN)
	$(GOLANGCI_DIR)/$(GOLANGCI_CMD)

# install a local golangci-lint if not found.
$(GOLANGCI_TMP_BIN):
	curl -OL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(GOLANGCI_LINT_ARCHIVE)
	mkdir -p $(GOLANGCI_DIR)/
	tar -xf $(GOLANGCI_LINT_ARCHIVE) --strip-components=1 -C $(GOLANGCI_DIR)/
	chmod +x $(GOLANGCI_TMP_BIN)
	rm -f $(GOLANGCI_LINT_ARCHIVE)
