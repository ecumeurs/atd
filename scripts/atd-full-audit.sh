#!/usr/bin/env bash
# atd-full-audit.sh
# Full ATD audit pipeline: audit -> fixer -> compare -> unified report
#
# Usage:
#   ./atd-full-audit.sh -docs /path/to/docs [-bin /path/to/bin] [-tmp /tmp/atd_run]
#
# Output:
#   /tmp/atd_run/full_report.md
#   /tmp/atd_run/audit_report.txt
#   /tmp/atd_run/fixer_log.txt
#   /tmp/atd_run/compare_<a>__<b>.md  (one per collision pair)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [[ -f "$SCRIPT_DIR/lib/logging.sh" ]]; then
    source "$SCRIPT_DIR/lib/logging.sh"
else
    atd_log() { :; }
fi
# --- Defaults -----------------------------------------------------------------
DOCS=""
if [[ -d "$(dirname "$0")/bin" ]]; then
    BIN_DIR="$(cd "$(dirname "$0")/bin" && pwd)"
else
    BIN_DIR="$(cd "$(dirname "$0")" && pwd)"
fi
TMP_DIR="/tmp/atd_run"

# --- Argument parsing ---------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    -docs) DOCS="$2";    shift 2 ;;
    -bin)  BIN_DIR="$2"; shift 2 ;;
    -tmp)  TMP_DIR="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$DOCS" ]]; then
  echo "Error: -docs <path> is required." >&2
  exit 1
fi

ATD_AUDIT="${BIN_DIR}/atd-audit"
ATD_FIXER="${BIN_DIR}/atd-audit-fixer"
ATD_COMPARE="${BIN_DIR}/atd-compare"

for bin in "$ATD_AUDIT" "$ATD_FIXER" "$ATD_COMPARE"; do
  if [[ ! -x "$bin" ]]; then
    echo "Error: binary not found or not executable: $bin" >&2
    echo "Run: cd scripts && go build -o bin/$(basename "$bin") ./$(basename "$bin")" >&2
    exit 1
  fi
done

mkdir -p "$TMP_DIR"
AUDIT_REPORT="$TMP_DIR/audit_report.txt"
FIXER_LOG="$TMP_DIR/fixer_log.txt"
FULL_REPORT="$TMP_DIR/full_report.md"

echo "============================================"
echo "       ATD Full Audit Pipeline              "
echo "============================================"
echo "Docs:  $DOCS"
echo "Bin:   $BIN_DIR"
echo "Tmp:   $TMP_DIR"
echo ""

atd_log "atd-full-audit.sh" "Started full audit pipeline for docs at DOCS"

# --- Phase 1+2: Audit ---------------------------------------------------------
echo "[1/5] Running atd-audit..."
"$ATD_AUDIT" -docs "$DOCS" | tee "$AUDIT_REPORT"
echo ""
echo "Audit report saved: $AUDIT_REPORT"
echo ""

# --- Phase 3: Auto-fix bloated atoms ------------------------------------------
BLOAT_COUNT=$(grep -c "\[BLOATED\]" "$AUDIT_REPORT" || true)
if [[ "$BLOAT_COUNT" -gt 0 ]]; then
  echo "[3/5] Fixing $BLOAT_COUNT bloated atom(s) via atd-audit-fixer..."
  "$ATD_FIXER" -audit "$AUDIT_REPORT" -docs "$DOCS" | tee "$FIXER_LOG"
  echo ""
  echo "Fixer log saved: $FIXER_LOG"
else
  echo "[3/5] No bloated atoms -- skipping fixer."
  echo "" > "$FIXER_LOG"
fi
echo ""

# --- Phase 4: Per-collision comparison reports --------------------------------
COLLISION_COUNT=$(grep -c "\[MISSING ABSTRACTION\]" "$AUDIT_REPORT" || true)
if [[ "$COLLISION_COUNT" -gt 0 ]]; then
  echo "[4/5] Generating $COLLISION_COUNT collision report(s) via atd-compare..."

  # Parse collision pairs: lines like "[COLLISION] file_a.atom.md <--> file_b.atom.md"
  while IFS= read -r line; do
    if [[ "$line" =~ \[COLLISION\][[:space:]]([^[:space:]]+)[[:space:]]'<-->'[[:space:]]([^[:space:]]+) ]]; then
      FILE_A="${BASH_REMATCH[1]}"
      FILE_B="${BASH_REMATCH[2]}"

      # Only produce a report if the next line is MISSING ABSTRACTION
      NEXT_LINE=$(grep -A1 -F "$line" "$AUDIT_REPORT" | tail -1 || true)
      if [[ "$NEXT_LINE" != *"MISSING ABSTRACTION"* ]]; then
        continue
      fi

      PATH_A="$DOCS/$FILE_A"
      PATH_B="$DOCS/$FILE_B"

      if [[ ! -f "$PATH_A" || ! -f "$PATH_B" ]]; then
        echo "  [SKIP] File not found: $FILE_A or $FILE_B"
        continue
      fi

      SAFE_A="${FILE_A//.atom.md/}"
      SAFE_B="${FILE_B//.atom.md/}"
      OUT_MD="$TMP_DIR/compare_${SAFE_A}__${SAFE_B}.md"

      echo "  Comparing $FILE_A  <-->  $FILE_B"
      "$ATD_COMPARE" -a "$PATH_A" -b "$PATH_B" -out "$OUT_MD" 2>/dev/null || \
        echo "  [WARN] atd-compare failed for this pair"
    fi
  done < "$AUDIT_REPORT"

  echo "Collision reports written to: $TMP_DIR/compare_*.md"
else
  echo "[4/5] No missing-abstraction collisions -- skipping compare."
fi
echo ""

# --- Phase 5: Assemble unified report -----------------------------------------
echo "[5/5] Assembling unified report..."
{
  echo "# ATD Full Audit Report"
  echo ""
  echo "> Generated: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "> Docs: \`$DOCS\`"
  echo ""
  echo "---"
  echo ""
  echo "## Audit Summary"
  echo ""
  echo '```'
  cat "$AUDIT_REPORT"
  echo '```'
  echo ""

  if [[ -s "$FIXER_LOG" ]]; then
    echo "---"
    echo ""
    echo "## Fixer Log"
    echo ""
    echo '```'
    cat "$FIXER_LOG"
    echo '```'
    echo ""
  fi

  for md in "$TMP_DIR"/compare_*.md; do
    [[ -f "$md" ]] || continue
    echo "---"
    echo ""
    cat "$md"
    echo ""
  done

} > "$FULL_REPORT"

echo "Full report saved: $FULL_REPORT"
echo ""
echo "============================================"
echo "  Pipeline complete."
echo "============================================"
atd_log "atd-full-audit.sh" "Completed full audit pipeline"
