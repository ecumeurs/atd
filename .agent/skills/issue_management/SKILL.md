---
name: issue_management
description: File, update, and close tracked issues in /workspace/issues/. Use this skill whenever you discover a bug, design risk, data race, security concern, or technical debt during your work.
---

# Issue Management Skill

## When to Activate

Use this skill whenever you encounter **any** of the following during your work:

- A **deadlock, race condition, or concurrency hazard**
- A **silent failure path** (no panic, no log, but wrong behavior)
- A **design constraint** that must not be violated but is not enforced at compile time
- A **footgun** in a shared utility (e.g., a flag that hangs a caller if misused)
- Any **TODO/FIXME** that represents real risk, not just cosmetics
- Any time the **user explicitly asks** you to track something as an issue

---

## Directory Layout

All issues live under `/workspace/issues/`:

```
/workspace/issues/
├── README.md                         ← index table (must be kept up to date)
├── YYYYMMDD_short_slug.md            ← individual issue files
└── ...
```

If `/workspace/issues/` does not yet exist, create it along with `README.md`.

---

## Step 1 — Choose a Filename and Ref

**Filename format:** `Ref_YYYYMMDD_short_slug.md`

- `YYYYMMDD`: today's date in UTC (use the date you have been told in the conversation, do **not** try to call a system tool for the time).
- `short_slug`: lowercase, underscores only, 3–6 words max, describing **component + nature of problem**.

**Ref:** Run the script to determine the next available `ISS-NNN`:

```bash
python3 .agent/skills/issue_management/scripts/list_issues.py --next-ref
# prints: ISS-002
```

Examples:
```
20260223_actor_deadlock_risk.md   →  ISS-001
20260301_messagequeue_data_race.md  →  ISS-002
```

---

## Step 2 — Write the Issue File

Copy the template from `templates/issue.md` (path: `.agent/skills/issue_management/templates/issue.md`) and fill in every field.

**Rules:**
- `**ID:**` must match the filename (without `.md`).
- `**Ref:**` must be the `ISS-NNN` value obtained from `--next-ref`. Do not reuse or skip numbers.
- `**Severity:**` must be one of: `Critical` / `High` / `Medium` / `Low`.
- `**Status:**` starts as `Open`. Valid states: `Open` / `In Progress` / `Resolved` / `Wont Fix`.
- `**Component:**` use the Go import path or relative file path of the **primary** affected package.
- `**Affects:**` list known callers or consumers.
- The **Problem Scenario** section must include either a code block, an ASCII diagram, or a numbered step list; not plain prose alone.
- The **Recommended Fix** section must contain at least a **Short term** entry.

---

## Step 3 — Update the Index

Open `/workspace/issues/README.md` and add a row to the index table:

```markdown
| ISS-NNN | [YYYYMMDD_slug.md](YYYYMMDD_slug.md) | Severity | Status | One-line summary |
```

Keep rows in **reverse-chronological order** (newest first).

If the `README.md` does not yet have the index table, create it with this header:

```markdown
## Index

| Ref | File | Severity | Status | Summary |
|---|---|---|---|---|
```


## Update the root README.md with an active issues table
python3 .agent/skills/issue_management/scripts/list_issues.py --update-readme


---

## Step 4 — Notify the User

After filing, tell the user:
1. The **Ref** (`ISS-NNN`), ID, and file path.
2. The one-line summary.
3. Your severity assessment and why.

Example:
> Filed **ISS-001** `20260223_actor_deadlock_risk` (Medium) — actor executor goroutine can deadlock when handler blocks on a local inter-actor reply channel. See `/workspace/issues/20260223_actor_deadlock_risk.md`.

---

## Listing & Searching Issues

Use `scripts/list_issues.py` to inspect the current state of the issue tracker.  
The script reads issue files directly — it never relies on the README index — so results are always accurate.

### Script Location

```
.agent/skills/issue_management/scripts/list_issues.py
```

### Usage

```bash
# List all issues (newest first, with status breakdown)
python3 .agent/skills/issue_management/scripts/list_issues.py

# Filter by status
python3 .agent/skills/issue_management/scripts/list_issues.py --status open
python3 .agent/skills/issue_management/scripts/list_issues.py --status resolved

# Filter by severity
python3 .agent/skills/issue_management/scripts/list_issues.py --severity high

# Keyword search (title, summary, component, affects, filename)
python3 .agent/skills/issue_management/scripts/list_issues.py --search "actor"
python3 .agent/skills/issue_management/scripts/list_issues.py --search "deadlock"

# Combine filters
python3 .agent/skills/issue_management/scripts/list_issues.py --status open --severity medium

# Print full file content for matching issues
python3 .agent/skills/issue_management/scripts/list_issues.py --search "queue" --full

# Update the root README.md with an active issues table
python3 .agent/skills/issue_management/scripts/list_issues.py --update-readme

# Override issues directory (useful in non-standard setups)
python3 .agent/skills/issue_management/scripts/list_issues.py --dir /path/to/issues
```

### When the Agent Should Run This

- **Before filing a new issue**: run `--search <keyword>` to check whether a similar issue already exists. Do not file duplicates.
- **When the user asks "do we have an issue on X?"**: run `--search X` and report the results.
- **When the user asks "what's left to do?"**: run `--status open` and summarise the output.
- **At the start of a debugging session** on a known risky component: run `--search <component>` to surface any pre-existing caveats.
- **After creating or modifying an issue**: run `--update-readme` to ensure the project's root `README.md` reflects the current active issues.

### Output Fields

| Column | Source |
|---|---|
| Ref | `**Ref:**` field (`ISS-NNN`), shown in bold; `?` if missing |
| Date | `**Date:**` field in the issue file |
| Severity | `**Severity:**` field (colour coded: red=Critical/High, yellow=Medium, green=Low) |
| Status | `**Status:**` field (yellow=Open, blue=In Progress, green=Resolved, grey=Wont Fix) |
| Title | First `#` heading, with "Issue: " prefix stripped |
| Component | `**Component:**` field |
| Summary | First paragraph under `## Summary`, truncated to 110 chars |

---

## Updating an Existing Issue

When you revisit a previously filed issue (e.g., you found new information, or work was done):

1. Update **Status** and any relevant sections in the issue file.
2. Add a `## Change Log` section at the bottom if it doesn't exist, and append an entry:
   ```markdown
   ## Change Log
   - **YYYY-MM-DD**: [What changed and why]
   ```
3. Update the `Status` column in `/workspace/issues/README.md`.

---

## Closing an Issue

Set `**Status:**` to `Resolved` or `Wont Fix` in the issue file.  
Update the index row.  
Optionally add a final change log entry explaining the resolution.

---

## Severity Guide

| Severity | Meaning |
|---|---|
| **Critical** | Will cause data loss, security breach, or production outage if triggered. Fix immediately. |
| **High** | Significant, reproducible, or likely to be triggered in normal usage. Fix before next release. |
| **Medium** | Latent risk, low likelihood under current usage patterns. Track and fix proactively. |
| **Low** | Cosmetic, unlikely, or already well-mitigated. File for awareness. |

---

## Quick Reference Checklist

```
[ ] --next-ref run to determine ISS-NNN
[ ] Date + slug chosen
[ ] Issue file created from template
[ ] All mandatory fields filled (ID, Ref, Date, Severity, Status, Component, Affects)
[ ] Problem Scenario has diagram or code block
[ ] Short term fix is specified
[ ] README.md index row added (with Ref as first column)
[ ] User notified with Ref, ID, summary, severity
```
