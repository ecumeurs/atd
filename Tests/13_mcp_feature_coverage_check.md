# Task 13 — MCP Feature Coverage Check

## Objective
Audit all 14 MCP tools against the full ATD feature set. Document which operations are fully covered via MCP, which require CLI fallback, and any gaps to file as issues.

## Prerequisites
- Tasks 09–12 complete

## Tools Coverage Matrix

| MCP Tool | CLI Equivalent | LLM | Tested? | Result |
|---|---|---|---|---|
| `atd_query` | `atd query` | No | Task 10 | |
| `atd_crawl` | `atd crawl` | No | — | |
| `atd_weave` | `atd weave` | No | Task 11 | |
| `atd_update` | `atd update` | No | Task 11 | |
| `atd_roadmap` | `atd roadmap` | No | Task 12 | |
| `atd_verify` | `atd verify` | No | — | |
| `atd_assemble` | `atd assemble` | No/Yes | — | |
| `atd_test_links` | `atd test-links` | No | — | |
| `atd_dissect` | `atd dissect` | Yes | Task 10 | |
| `atd_index` | `atd index` | Yes (embed) | Task 12 | |
| `atd_search` | `atd search` | Yes (embed) | Task 12 | |
| `atd_audit` | `atd audit` | Yes | — | |
| `atd_recon` | `atd recon` | Yes | — | |
| `atd_discover` | `atd discover` | Yes | — | |

## Steps

### 1. Test atd_crawl via MCP
```json
{
  "name": "atd_crawl",
  "arguments": {
    "docs": "/home/bastien/work/skill/upsilonbattle/docs",
    "src": "/home/bastien/work/skill/upsilonbattle/battlearena",
    "gaps": false
  }
}
```

### 2. Test atd_crawl gap detection
```json
{
  "name": "atd_crawl",
  "arguments": {
    "docs": "/home/bastien/work/skill/upsilonbattle/docs",
    "gaps": true
  }
}
```

### 3. Test atd_verify via MCP
```json
{
  "name": "atd_verify",
  "arguments": {}
}
```

### 4. Test atd_assemble via MCP
```json
{
  "name": "atd_assemble",
  "arguments": {
    "starts": "ruler",
    "purpose": "Architecture overview"
  }
}
```

### 5. Test atd_recon via MCP
```json
{
  "name": "atd_recon",
  "arguments": {
    "atom": "/home/bastien/work/skill/upsilonbattle/docs/ruler.atom.md",
    "candidate": "/home/bastien/work/skill/upsilonbattle/battlearena/ruler/ruler.go"
  }
}
```

### 6. Fill in coverage matrix
For each tool tested above, mark result as `PASS`, `PARTIAL`, or `FAIL` and note any issues.

### 7. Document gaps
If any critical operation is not achievable via MCP, note it here for filing as an issue.

## Acceptance Criteria
- [ ] Coverage matrix is fully populated
- [ ] At least 12 of 14 tools return non-error results
- [ ] Any tools with issues are documented
- [ ] `atd_crawl` (gaps=true) returns orphaned STABLE atoms (if any)
- [ ] `atd_recon` returns a confidence score for ruler.go vs ruler atom
