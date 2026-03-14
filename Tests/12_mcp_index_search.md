# Task 12 — MCP: Index + Search via MCP Tools

## Objective
Test `atd_index` and `atd_search` through the MCP interface. Requires Ollama with `nomic-embed-text` to be reachable.

## Prerequisites
- Task 09 complete (MCP connected)
- Task 05 complete (initial index exists, or willing to rebuild)
- Ollama `nomic-embed-text` reachable

## Steps

### 1. Rebuild the source code index via MCP
> *"Use atd_index to re-index the upsilonbattle battlearena directory in 'all' mode."*

**MCP call:**
```json
{
  "name": "atd_index",
  "arguments": {
    "dir": "/home/bastien/work/skill/upsilonbattle/battlearena",
    "mode": "all"
  }
}
```
**Expected:** String like `"Indexed N chunks across M files (K skipped unchanged)"`

### 2. Semantic search for code via MCP
```json
{
  "name": "atd_search",
  "arguments": {
    "query": "damage computation attack defense",
    "scope": "code",
    "limit": 3
  }
}
```
**Expected:** Top 3 code chunks with similarity scores, pointing to `ruler.go` or similar.

### 3. Semantic search for docs via MCP
```json
{
  "name": "atd_search",
  "arguments": {
    "query": "turn order entity selection",
    "scope": "docs",
    "limit": 3
  }
}
```
**Expected:** ATD atom chunks with similarity scores.

### 4. Grep search via MCP
```json
{
  "name": "atd_search",
  "arguments": {
    "grep": "spec-link"
  }
}
```
**Expected:** List of files containing `spec-link`.

### 5. Roadmap scan via MCP
```json
{
  "name": "atd_roadmap",
  "arguments": {
    "dir": "/home/bastien/work/skill/upsilonbattle/battlearena"
  }
}
```
**Expected:** JSON complexity map identifying high-density files.

## Acceptance Criteria
- [ ] `atd_index` returns a summary with chunk counts > 0
- [ ] Semantic search (code scope) returns results from `.go` files
- [ ] Semantic search (docs scope) returns results from `.atom.md` files
- [ ] Grep search correctly lists files with the keyword
- [ ] `atd_roadmap` returns a valid JSON complexity report
