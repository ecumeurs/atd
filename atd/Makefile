SHELL := /bin/bash

# ATD lives as two separate Go modules:
#   .          -> module "atd-tools" (pkg/, config/, ...)
#   ./cmd/atd  -> module "atd"       (the CLI, nested go.mod)
#
# `go build ./...` / `go test ./...` run from this directory ONLY cover the
# first module -- the nested module is invisible to them even though a
# go.work file exists at this level (verified: `go list ./...` from here
# lists 17 atd-tools packages and zero atd/cmd/atd packages). That gap is
# exactly how a five-error compile break in cmd/atd shipped behind a
# "verified green" build. `verify` is therefore the ONLY definition of
# "green" for this repo: it always runs both modules, plus the WP-0
# tripwire that fails the build if a test run left the tree dirty.

ROOT       := $(CURDIR)
CMD_MODULE := $(CURDIR)/cmd/atd
MODULES    := $(ROOT) $(CMD_MODULE)

.PHONY: verify verify-race tripwire

verify:
	@set -e; \
	for m in $(MODULES); do \
		echo "==> [verify] $$m"; \
		(cd "$$m" && go build ./... && go vet ./... && go test ./...); \
	done
	$(MAKE) tripwire

verify-race:
	@set -e; \
	for m in $(MODULES); do \
		echo "==> [verify-race] $$m"; \
		(cd "$$m" && go build ./... && go vet ./... && go test -race ./...); \
	done
	$(MAKE) tripwire

# W-1 tripwire (test_atd_07_26.md §3.6): a test run must be a pure function
# of the checked-out tree. If anything changed on disk -- source rewritten,
# stray artifacts written, fixtures mutated -- fail loudly with the diff
# instead of letting a "green" run silently commit damage later (see
# incident I-1, §2.1).
tripwire:
	@test -z "$$(git status --porcelain)" || (echo "!! tripwire: working tree changed during verify !!"; git diff; git status --porcelain; exit 1)
