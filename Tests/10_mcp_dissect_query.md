# Task 10 — MCP: Dissect + Query via MCP Tools

## Objective
Exercise `atd_dissect` and `atd_query` through the MCP interface. Verify the IDE Agent can use these tools correctly and returns coherent results.

## Prerequisites
- Task 09 complete (MCP connection verified, 14 tools listed)
- `atd_dissect` requires a readable file path that the `atd` server process can access

## Steps

### 1. Query atoms via MCP
Ask the IDE Agent (MCP enabled):
> *"Use the atd_query MCP tool to list all atoms with type MECHANIC."*

**MCP call:**
```json
{
  "name": "atd_query",
  "arguments": {
    "field": "type",
    "search": "MECHANIC"
  }
}
```
**Expected:** JSON array of MECHANIC atoms from `upsilonbattle/docs/`.

### 2. Query by ID via MCP
> *"Use atd_query to find the atom with id 'ruler'."*

**MCP call:**
```json
{
  "name": "atd_query",
  "arguments": {
    "field": "id",
    "search": "ruler"
  }
}
```

### 3. Dissect a file via MCP (passthrough mode)
> *"Use atd_dissect to analyze battlearena/ruler/ruler.go — return the IDE prompt without calling Ollama."*

**MCP call:**
```json
{
  "name": "atd_dissect",
  "arguments": {
    "file": "/home/bastien/work/skill/upsilonbattle/battlearena/ruler/ruler.go",
    "llm": false
  }
}
```
**Expected:** The dissect prompt text (long string with line-numbered code content).

### 4. Dissect with Ollama via MCP (if available)
```json
{
  "name": "atd_dissect",
  "arguments": {
    "file": "/home/bastien/work/skill/upsilonbattle/battlearena/entity/entity.go",
    "llm": true
  }
}
```
**Expected:** JSON atom boundary proposals or IDE Agent delegation message.

## Acceptance Criteria
- [ ] `atd_query` with `type=MECHANIC` returns at least 3 atoms as JSON
- [ ] `atd_query` with `id=ruler` returns the ruler atom
- [ ] `atd_dissect` with `llm=false` returns a non-empty prompt string
- [ ] `atd_dissect` with `llm=true` returns either JSON proposals or delegation message
