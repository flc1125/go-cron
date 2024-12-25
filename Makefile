
TOOLS_MOD_DIR := ./internal/tools

ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)
ROOT_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR), $(ALL_GO_MOD_DIRS))

GO = go

# Tools

TOOLS = $(CURDIR)/.tools

$(TOOLS):
	@mkdir -p $@
$(TOOLS)/%: $(TOOLS_MOD_DIR)/go.mod | $(TOOLS)
	cd $(TOOLS_MOD_DIR) && \
	$(GO) build -o $@ $(PACKAGE)


GOLANGCI_LINT = $(TOOLS)/golangci-lint
$(TOOLS)/golangci-lint: PACKAGE=github.com/golangci/golangci-lint/cmd/golangci-lint

GORELEASE = $(TOOLS)/gorelease
$(GORELEASE): PACKAGE=golang.org/x/exp/cmd/gorelease

.PHONY: tools
tools: $(GOLANGCI_LINT) $(GORELEASE)

# Tests



.PHONY: golangci-lint golangci-lint-fix
golangci-lint-fix: ARGS=--fix
golangci-lint-fix: golangci-lint
golangci-lint: $(ROOT_GO_MOD_DIRS:%=golangci-lint/%)
golangci-lint/%: DIR=$*
golangci-lint/%: $(GOLANGCI_LINT)
	@echo 'golangci-lint $(if $(ARGS),$(ARGS) ,)$(DIR)' \
		&& cd $(DIR) \
		&& $(GOLANGCI_LINT) run --allow-serial-runners $(ARGS)

.PHONY: go-mod-tidy
go-mod-tidy: $(ALL_GO_MOD_DIRS:%=go-mod-tidy/%)
go-mod-tidy/%: DIR=$*
go-mod-tidy/%: 
	@echo "$(GO) mod tidy in $(DIR)" \
		&& cd $(DIR) \
		&& $(GO) mod tidy -compat=1.22.0

.PHONY: lint-modules
lint-modules: go-mod-tidy

.PHONY: lint
lint: lint-modules golangci-lint

.PHONY: gorelease
gorelease: $(ROOT_GO_MOD_DIRS:%=gorelease/%)
gorelease/%: DIR=$*
gorelease/%:| $(GORELEASE)
	@echo "gorelease in $(DIR):" \
		&& cd $(DIR) \
		&& $(GORELEASE) \
		|| echo ""