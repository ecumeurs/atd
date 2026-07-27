# ATD Trace Tool Enhancements - Summary

## Changes Made

### 1. Added `file_path` to Trace Output

**Problem:** The `atd_trace` command did not output the absolute file path to the atom definition, making it impossible for agents to read the full atom file.

**Solution:**
- Added `FilePath` field to `exploration.AtomBrief` struct (`atd/pkg/exploration/types.go`)
- Added `FilePath` field to `prompt.AtomBrief` struct (`atd/pkg/prompt/trace_summary.go`)
- Updated `addContext` function in `trace.go` to populate `FilePath` from `node.FilePath`

**Result:** Each atom in the `context` map now includes `"file_path": "/absolute/path/to/atom.atom.md"`

### 2. Fixed `--summary` Option Behavior

**Problem:** When using `--summary`, the command returned ONLY the narrative summary instead of adding it as a field to the JSON output. This meant users had to choose between structured data (blast radius, health metrics) OR narrative context, but not both.

**Solution:**
- Added `Summary` field to `TraceSnapshot` struct (`atd/pkg/exploration/types.go`)
- Modified CLI handler in `cmd/atd/cmd/trace.go` to append the summary to the snapshot instead of replacing the entire output
- Updated both LLM and IDE fallback paths to use the same approach

**Result:** When `--summary` is used, the JSON output now includes both:
- All structured data (target_id, layer, health_summary, metrics, graph_slice, context, warnings)
- A new `"summary"` field with the narrative summary

### 3. Updated Documentation

**ATD.md:**
- Updated `atd_trace` tool documentation to reflect new behavior
- Added detailed field descriptions for the output
- Clarified that `summary` adds a field rather than replacing output

### 4. Updated Skills

**atd-alter-atom:**
- Added Step 3: "Read the full atom file to understand complete context"
- Instructs agents to use `file_path` from trace output to read the atom file

**atd-pre-code-change:**
- Added instructions to read full atom files using `file_path` from trace output
- Updated pre-flight checklist to mention file reading capability
- Updated quick reference checklist

**atd-create-atom:**
- Updated CONTRACT/VISION reading instructions to use `file_path` from query results

## Testing

All tests pass:
```bash
cd /home/bastien/work/atd/atd
make verify
```

## Example Output

### Without `--summary`:
```json
{
  "target_id": "contract_atd",
  "layer": "BUSINESS",
  "health_summary": {
    "ancestry_complete": true,
    "has_business_origin": true,
    "implementation_rate": 0,
    "test_coverage_rate": 0
  },
  "context": {
    "contract_atd": {
      "id": "contract_atd",
      "human_name": "ATD Project Contract",
      "type": "CONTRACT",
      "layer": "BUSINESS",
      "intent": "...",
      "logic": "...",
      "file_path": "/home/bastien/work/atd/docs/contract_atd.atom.md"
    }
  },
  "warnings": ["Business atom has no Architecture dependents"]
}
```

### With `--summary`:
Same as above, plus:
```json
  "summary": "The contract_atd atom plays a crucial role in pinning down hard, non-negotiable invariants..."
```

## Files Modified

- `atd/pkg/exploration/types.go` - Added `FilePath` and `Summary` fields
- `atd/pkg/exploration/trace.go` - Populate `FilePath` in `addContext`
- `atd/pkg/prompt/trace_summary.go` - Added `FilePath` to `AtomBrief`
- `atd/cmd/atd/cmd/trace.go` - Changed `--summary` to add field instead of replace
- `atd/testdata/golden/prompts/trace_summary.txt` - Updated golden test
- `ATD.md` - Updated documentation
- `.agent/skills/atd-alter-atom/SKILL.md` - Added file reading instructions
- `.agent/skills/atd-pre-code-change/SKILL.md` - Added file reading instructions
- `.agent/skills/atd-create-atom/SKILL.md` - Added file reading instructions