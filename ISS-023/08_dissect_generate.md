# Task 08 — Dissect & Generate

**Depends on:** Task 04 (provider), Task 05 (prompt templates), Task 06 (continue protocol)  
**Produces:** `atd dissect` and `atd generate` subcommands

## Context

`atd-dissect` (69 lines) prepends line numbers to a file and emits a prompt to stdout. `atd-ollama-generate` (110 lines) takes that output and sends it to Ollama for JSON boundary extraction. Both use the `dissect` LLM task type.

## Steps

### 8.1 Create `cmd/dissect.go`

**Flags:** `-file <path>`, `--llm` (optional: route through provider instead of stdout)

**Execution flow:**

```
Read file → prepend line numbers → build prompt (prompt.DissectBuild)
│
├── Default (no --llm): Print prompt to stdout for IDE Agent
│
├── --llm flag set:
│   ├── Resolve provider for task "dissect"
│   │   ├── Ollama available → POST with DissectFormat() schema
│   │   │   └── Print JSON result to stdout
│   │   └── IDE fallback → 
│   │       ├── pipeline.WritePromptFile("dissect_<basename>", prompt)
│   │       ├── pipeline.WriteTaskList("atd dissect", [...])
│   │       └── Print: "Task delegated. See pipeline_output/task_list.md"
│   └── Exit
```

**Source reference:** `atd-dissect/main.go` lines 13-68

**Help text:**
```
Dissect a document or source code file into atomic boundaries.

By default, outputs the analysis prompt to stdout for the IDE Agent to process.
Use --llm to route through the tiered LLM provider (Ollama).

Examples:
  atd dissect -file ruler.go              # Print prompt to stdout
  atd dissect -file ruler.go --llm        # Use Ollama for extraction
  atd dissect -file commerce.md --llm     # Dissect documentation
```

### 8.2 Create `cmd/generate.go`

**Flags:** `-dissect <path>` (path to dissect output / prompt file)

**Execution flow:**

```
Read dissect file → use it as prompt
│
├── Resolve provider for task "dissect" (same task type)
│   ├── Ollama → POST with DissectFormat(), force JSON format
│   │   └── Print JSON result
│   └── IDE fallback →
│       ├── pipeline.WritePromptFile("generate_<basename>", prompt)
│       ├── pipeline.WriteTaskList("atd generate", [...])
│       └── Print: "Task delegated"
└── Exit
```

**Source reference:** `atd-ollama-generate/main.go` lines 62-109

**Help text:**
```
Generate atom boundaries from a dissect output file.

Takes the prompt output from 'atd dissect' and sends it to the LLM
for structured JSON extraction of atom boundaries.

Example:
  atd dissect -file ruler.go > /tmp/dissect_output.txt
  atd generate -dissect /tmp/dissect_output.txt
```

### 8.3 Write tests

**`cmd/dissect_test.go`:**
- Test: `-file` with nonexistent file → error
- Test: `-file` with valid file → stdout contains line numbers `001:`, `002:`, etc.
- Test: Prompt contains `<System_Context>` and `<Document>` tags

**`cmd/generate_test.go`:**
- Test: `-dissect` with nonexistent file → error
- Test: `-dissect` with valid file → prompt passed to LLM or written to pipeline_output

## Acceptance Criteria

- [ ] `atd dissect --help` shows flags and examples
- [ ] `atd dissect -file <valid_file>` outputs prompt with numbered lines
- [ ] `atd dissect -file <valid_file> --llm` (without Ollama) creates `pipeline_output/task_list.md`
- [ ] `atd generate --help` shows usage
- [ ] Tests pass
