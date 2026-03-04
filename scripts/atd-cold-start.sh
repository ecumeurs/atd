#!/bin/bash
set -e

# atd-cold-start.sh
# Unifies Iteration 14 (Top-Down Primer) and Iteration 15 (Hybrid Extraction Pipeline)
# Usage: ./atd-cold-start.sh <project_dir> [target_package_filter]

PROJECT_ROOT=$1
TARGET_PACKAGE=${2:-""}
SCRIPTS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -f "$SCRIPTS_ROOT/lib/logging.sh" ]]; then
    source "$SCRIPTS_ROOT/lib/logging.sh"
else
    atd_log() { :; }
fi

if [ -z "$PROJECT_ROOT" ]; then
    echo "Usage: ./atd-cold-start.sh <project_dir> [target_package_filter]"
    exit 1
fi

# Convert PROJECT_ROOT to absolute path
PROJECT_ROOT=$(realpath "$PROJECT_ROOT")

echo "=========================================="
echo "      ATD Streamlined Pipeline (v14+15)   "
echo "=========================================="
echo "Project: $PROJECT_ROOT"
if [ -n "$TARGET_PACKAGE" ]; then
    echo "Target Package Filter: $TARGET_PACKAGE"
fi
echo "------------------------------------------"
atd_log "atd-cold-start.sh" "Started cold start pipeline for $PROJECT_ROOT"

# Phase 1: Top-Down Primer (Iteration 14)
echo -e "\n[Phase 1] Discovery: Extracting Top-Down Primer (READMEs)..."
MD_FILES=$(find "$PROJECT_ROOT" -name "*.md" | grep -v "docs/")
for md in $MD_FILES; do
    if [ -n "$TARGET_PACKAGE" ] && [[ "$md" != *"$TARGET_PACKAGE"* ]]; then
        continue # Skip documentation not inside the target package
    fi
    echo "-> Found Documentation: $md"
    # Create a unique name from the path: e.g. battlearena_ruler_rules_README
    output_name=$(echo "$md" | sed "s|$PROJECT_ROOT/||g" | sed 's|/|_|g' | sed 's|.md$||g')
    # We prefix with "domain_" so the LLM Assistant knows to extract Domain atoms.
    mkdir -p "$PROJECT_ROOT/pipeline_output"
    # We just copy the MD to the pipeline output for the LLM to read.
    cp "$md" "$PROJECT_ROOT/pipeline_output/domain_${output_name}.md.txt"
done

# Phase 2: Building Mechanical Map (Iteration 15)
echo -e "\n[Phase 2] Building Mechanical Map (Roadmap)..."
cd "$SCRIPTS_ROOT/atd-roadmap-builder"
go run main.go -dir "$PROJECT_ROOT" -out "$PROJECT_ROOT/roadmap.json"

# Phase 3: Vectorizing Codebase (Iteration 15)
echo -e "\n[Phase 3] Vectorizing Codebase (Ollama Indexer)..."
cd "$SCRIPTS_ROOT/atd-ollama-indexer"
go run main.go -dir "$PROJECT_ROOT" -db "$PROJECT_ROOT/.atd_index.db"

# Phase 4: Priority Queuing
echo -e "\n[Phase 4] Priority Queuing..."
cd "$PROJECT_ROOT"

if [ -n "$TARGET_PACKAGE" ]; then
    # Filter by target package before sorting density
    jq -r '.items[].file_path' roadmap.json | grep "$TARGET_PACKAGE" | sort | uniq -c | sort -nr | head -n 5 > top_targets.txt
else
    jq -r '.items[].file_path' roadmap.json | sort | uniq -c | sort -nr | head -n 5 > top_targets.txt
fi

echo "Top dense files prioritized:"
cat top_targets.txt

# Phase 5: Executing Cloud Dissection (Iteration 15)
echo -e "\n[Phase 5] Executing Mechanical Dissection (atd-dissect)..."
while read line; do
    file_path=$(echo "$line" | awk '{print $2}')
    if [ -z "$file_path" ]; then continue; fi

    echo "-> Dissecting target: $file_path"
    cd "$SCRIPTS_ROOT/atd-dissect"
    output_name=$(basename "$file_path" .go)
    go run main.go -file "$file_path" > "$PROJECT_ROOT/pipeline_output/dissect_${output_name}.json"
done < "$PROJECT_ROOT/top_targets.txt"

echo -e "\n[Pipeline Paused] Mechanical Extraction Complete."
echo "Agent (Cloud LLM) must now:"
echo " 1. Read 'pipeline_output/domain_*.md.txt' and generate DOMAIN atoms."
echo " 2. Read 'pipeline_output/dissect_*.json' and generate MECHANIC atoms (reconciling with the DOMAIN atoms)."

# Phase 6: Automatic Dependent Weaving (Iteration 15)
echo -e "\n[Phase 6] Weaving Atom Dependencies..."
echo "(Note: This step requires the Agent to have already written the .atom.md files in docs/)"
if [ "$(ls -A $PROJECT_ROOT/docs/*.atom.md 2>/dev/null)" ]; then
    cd "$SCRIPTS_ROOT/atd-link-weaver"
    go run main.go -docs "$PROJECT_ROOT/docs"
else
    echo "No .atom.md files found in docs/. Skipping link weaving."
fi

# Phase 7: Automated Tagging (Iteration 16 Search-Then-Recon)
echo -e "\n[Phase 7] Tagging Codebase via Search-Then-Recon..."
echo "Executing Recon in the background. Check 'recon_audit.log' for progress."
if [ "$(ls -A $PROJECT_ROOT/docs/*.atom.md 2>/dev/null)" ]; then
    (
        for atom in $PROJECT_ROOT/docs/*.atom.md; do
            # 1. Extract the INTENT from the Atom
            intent=$(awk '/^## INTENT/{flag=1; next} /^##/{flag=0} flag' "$atom" | tr '\n' ' ' | tr -d '"'\''' )
            if [ -z "$intent" ]; then
                echo "-> Skipping $(basename $atom): No INTENT found to query." >> "$PROJECT_ROOT/recon_audit.log"
                continue
            fi
            
            # 2. Search for the best semantic match
            echo "-> Searching index for semantic match for $(basename $atom)..." >> "$PROJECT_ROOT/recon_audit.log"
            cd "$SCRIPTS_ROOT/atd-ollama-search"
            match_output=$(go run main.go -db "$PROJECT_ROOT/.atd_index.db" -query "$intent" -limit 1 2>> "$PROJECT_ROOT/recon_audit.log")
            
            # Extract the raw file path from the output string using sed
            matched_file=$(echo "$match_output" | grep "\[Match 1\]" | sed -n 's/.*File: \([^ ]*\).*/\1/p')
            
            if [ -n "$matched_file" ] && [ -f "$matched_file" ]; then
                # 3. Recon against the single match
                echo "-> Semantic match found: $(basename $matched_file). Running Recon..." >> "$PROJECT_ROOT/recon_audit.log"
                cd "$SCRIPTS_ROOT/atd-recon"
                go run main.go -atom "$atom" -candidate "$matched_file" >> "$PROJECT_ROOT/recon_audit.log" 2>&1
            else
                echo "-> No valid semantic match found for $(basename $atom)." >> "$PROJECT_ROOT/recon_audit.log"
            fi
        done
        echo "Recon Background Job Finished." >> "$PROJECT_ROOT/recon_audit.log"
    ) &
else
    echo "No .atom.md files found in docs/. Skipping recon tagging."
fi

echo -e "\n[Pipeline Finished] Streamlined Protocol Complete."
atd_log "atd-cold-start.sh" "Completed cold start pipeline"
