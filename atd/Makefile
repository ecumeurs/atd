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
REPO_ROOT  := $(abspath $(ROOT)/..)

.PHONY: verify verify-race tripwire dogfood dogfood-quick dogfood-update-baseline coverage-floors documentalist-eval

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

# WP-3 dogfood gate (test_atd_07_26.md §3.4, §6): ATD's own docs/ corpus is
# the best integration test the toolkit will ever have, and it should be
# wired in, not eyeballed. This target builds the real binary, smoke-tests
# `atd check --full` against the repo's docs/ (must not crash/error -- see
# E3/E5), and runs the two `dogfood`-tagged Go tests that ratchet `atd lint`
# and the S6 corpus-parse-audit (E4, the H3-swallow bug) against the
# checked-in baseline (testdata/dogfood_baseline.txt): pre-existing authoring
# debt does not fail the gate, any NEW finding does.
dogfood:
	@set -e; \
	echo "==> [dogfood] building atd binary"; \
	mkdir -p $(ROOT)/bin; \
	(cd $(CMD_MODULE) && go build -o $(ROOT)/bin/atd .); \
	echo "==> [dogfood] atd check --full smoke test against $(REPO_ROOT)/docs"; \
	(cd $(REPO_ROOT) && $(ROOT)/bin/atd check --full >/dev/null); \
	echo "==> [dogfood] lint ratchet (TestDogfoodLintRatchet)"; \
	(cd $(CMD_MODULE) && go test -tags dogfood ./cmd/... -run TestDogfoodLintRatchet -v); \
	echo "==> [dogfood] S6 corpus parse audit (TestDogfoodCorpusParseAudit)"; \
	(cd $(ROOT) && go test -tags dogfood ./pkg/atom/... -run TestDogfoodCorpusParseAudit -v)

# Fast subset for the pre-commit hook: the lint ratchet only -- pure file
# parsing plus link resolution against the already-loaded config, no binary
# build, no whole-project `check --full` walk. Keep this fast; it runs on
# every commit that stages an atom file.
dogfood-quick:
	@cd $(CMD_MODULE) && go test -tags dogfood ./cmd/... -run TestDogfoodLintRatchet

# WP-7 coverage floors (test_atd_07_26.md §5.3): per-package floors on the
# engine packages only -- no blanket threshold (it breeds flag-tests).
# Ratchet upward when a package durably clears a higher bar; never downward.
COVERAGE_FLOORS := pkg/exploration:60 pkg/coverage:60 pkg/atom:70

coverage-floors:
	@set -e; fail=0; \
	for spec in $(COVERAGE_FLOORS); do \
		pkg=$${spec%:*}; floor=$${spec#*:}; \
		pct=$$(cd $(ROOT) && go test -cover ./$$pkg | grep -oE 'coverage: [0-9.]+' | grep -oE '[0-9.]+'); \
		if [ -z "$$pct" ]; then echo "FAIL $$pkg: could not measure coverage"; fail=1; \
		elif awk -v p="$$pct" -v f="$$floor" 'BEGIN{exit !(p+0 >= f+0)}'; then \
			echo "ok   $$pkg coverage $$pct% >= floor $$floor%"; \
		else \
			echo "FAIL $$pkg coverage $$pct% < floor $$floor%"; fail=1; \
		fi; \
	done; exit $$fail

# Model-backed persona evaluation is deliberately opt-in: it can consume a
# provider, is slower than unit tests, and never belongs in the green gate.
documentalist-eval:
	@DOCUMENTALIST_LIVE=1 python3 $(REPO_ROOT)/tests/documentalist/runner.py --all --live

# Conscious ratchet-down (or, when reviewed, ratchet-tighten): regenerates
# testdata/dogfood_baseline.txt to mirror current reality for both
# categories. Never runs implicitly -- run `make dogfood` first, read what's
# new or fixed, and only then run this to accept it into the baseline.
dogfood-update-baseline:
	@set -e; \
	echo "==> [dogfood] regenerating baseline (lint)"; \
	(cd $(CMD_MODULE) && ATD_DOGFOOD_UPDATE=1 go test -tags dogfood ./cmd/... -run TestDogfoodLintRatchet -v); \
	echo "==> [dogfood] regenerating baseline (S6 corpus)"; \
	(cd $(ROOT) && ATD_DOGFOOD_UPDATE=1 go test -tags dogfood ./pkg/atom/... -run TestDogfoodCorpusParseAudit -v)
