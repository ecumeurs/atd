#!/usr/bin/env python3
"""
list_issues.py — Issue listing utility for /workspace/issues/

Reads issue files directly (not the README index) so results always reflect
the actual file state, even if the index is stale.

Usage:
  python3 list_issues.py                          # list all issues
  python3 list_issues.py --status open            # filter by status
  python3 list_issues.py --severity high          # filter by severity
  python3 list_issues.py --search "actor"         # keyword search in title/summary/component/ref
  python3 list_issues.py --status open --search "queue"
  python3 list_issues.py --full                   # print full file content for matches
  python3 list_issues.py --next-ref               # print the next available ISS-NNN
  python3 list_issues.py --update-readme          # update the root README.md with a table of active issues

Status values:  open | in progress | resolved | wont fix
Severity values: critical | high | medium | low
"""

import argparse
import os
import re
import sys
from dataclasses import dataclass, field
from typing import Optional


def _find_issues_dir() -> str:
    """Walk up from cwd until we find an 'issues/' directory, then return it."""
    # First, try relative to this script (4 levels up: scripts -> issue_management -> skills -> .agent -> root)
    candidate = os.path.normpath(os.path.join(os.path.dirname(__file__), "..", "..", "..", "..", "issues"))
    if os.path.isdir(candidate):
        return candidate
    # Fallback: walk up from cwd
    cwd = os.getcwd()
    while True:
        candidate = os.path.join(cwd, "issues")
        if os.path.isdir(candidate):
            return candidate
        parent = os.path.dirname(cwd)
        if parent == cwd:
            break
        cwd = parent
    # Last resort: original path (will fail gracefully in load_issues)
    return os.path.normpath(os.path.join(os.path.dirname(__file__), "..", "..", "..", "issues"))

ISSUES_DIR = _find_issues_dir()


@dataclass
class Issue:
    filename: str
    filepath: str
    id: str = ""
    ref: str = ""   # Short sequential ID, e.g. ISS-001
    date: str = ""
    severity: str = ""
    status: str = ""
    component: str = ""
    affects: str = ""
    title: str = ""
    summary: str = ""
    raw: str = ""

    def matches_status(self, status_filter: Optional[str]) -> bool:
        if not status_filter:
            return True
        return self.status.lower() == status_filter.lower()

    def matches_severity(self, severity_filter: Optional[str]) -> bool:
        if not severity_filter:
            return True
        return self.severity.lower() == severity_filter.lower()

    def matches_search(self, keyword: Optional[str]) -> bool:
        if not keyword:
            return True
        kw = keyword.lower()
        haystack = " ".join([
            self.title, self.summary, self.component,
            self.affects, self.id, self.ref, self.filename
        ]).lower()
        return kw in haystack

    def ref_number(self) -> int:
        """Return the integer part of ISS-NNN, or 0 if unparseable."""
        m = re.search(r"ISS-(\d+)", self.ref, re.IGNORECASE)
        return int(m.group(1)) if m else 0

    def status_color(self) -> str:
        colors = {
            "open": "\033[93m",          # yellow
            "in progress": "\033[94m",   # blue
            "resolved": "\033[92m",      # green
            "wont fix": "\033[90m",      # grey
        }
        return colors.get(self.status.lower(), "\033[0m")

    def severity_color(self) -> str:
        colors = {
            "critical": "\033[91m",      # red
            "high": "\033[91m",          # red
            "medium": "\033[93m",        # yellow
            "low": "\033[92m",           # green
        }
        return colors.get(self.severity.lower(), "\033[0m")


RESET = "\033[0m"
BOLD  = "\033[1m"
DIM   = "\033[2m"


def parse_issue(filepath: str) -> Issue:
    with open(filepath, "r", encoding="utf-8") as f:
        raw = f.read()

    filename = os.path.basename(filepath)
    issue = Issue(filename=filename, filepath=filepath, raw=raw)

    # Title: first markdown heading
    title_match = re.search(r"^#\s+(.+)", raw, re.MULTILINE)
    if title_match:
        issue.title = title_match.group(1).strip()
        # Strip "Issue: " prefix if present
        issue.title = re.sub(r"^Issue:\s*", "", issue.title)

    # Metadata fields: **Key:** value
    def extract_field(label: str) -> str:
        m = re.search(rf"\*\*{label}:\*\*\s*`?([^`\n]+)`?", raw)
        return m.group(1).strip() if m else ""

    issue.id        = extract_field("ID")
    issue.ref       = extract_field("Ref")
    issue.date      = extract_field("Date")
    issue.severity  = extract_field("Severity")
    issue.status    = extract_field("Status")
    issue.component = extract_field("Component")
    issue.affects   = extract_field("Affects")

    # Summary: first paragraph after "## Summary"
    summary_match = re.search(r"## Summary\s*\n+(.+?)(?:\n\n|\n##)", raw, re.DOTALL)
    if summary_match:
        # Collapse whitespace into a single line
        issue.summary = " ".join(summary_match.group(1).split())

    # Fallback: use filename slug if fields are missing
    if not issue.id:
        issue.id = filename.replace(".md", "")

    return issue


def load_issues(issues_dir: str) -> list[Issue]:
    if not os.path.isdir(issues_dir):
        print(f"[error] Issues directory not found: {issues_dir}", file=sys.stderr)
        sys.exit(1)

    issues = []
    for fname in sorted(os.listdir(issues_dir), reverse=True):  # newest first
        if fname.endswith(".md") and fname != "README.md":
            fpath = os.path.join(issues_dir, fname)
            try:
                issues.append(parse_issue(fpath))
            except Exception as e:
                print(f"[warn] Could not parse {fname}: {e}", file=sys.stderr)
    return issues


def print_summary_row(issue: Issue) -> None:
    sc  = issue.severity_color()
    tc  = issue.status_color()
    ref   = f"{BOLD}{issue.ref:<8}{RESET}" if issue.ref else f"{DIM}{'?':<8}{RESET}"
    sev   = f"{sc}{issue.severity:<8}{RESET}"
    stat  = f"{tc}{issue.status:<12}{RESET}"
    title = f"{BOLD}{issue.title}{RESET}" if issue.title else issue.filename
    date  = f"{DIM}{issue.date}{RESET}" if issue.date else ""
    print(f"  {ref}  {date}  {sev}  {stat}  {title}")
    if issue.component:
        print(f"                 {DIM}component: {issue.component}{RESET}")
    if issue.summary:
        # Truncate long summaries
        summ = issue.summary if len(issue.summary) <= 110 else issue.summary[:107] + "..."
        print(f"                 {summ}")
    print()


def generate_markdown_table(issues: list[Issue], issues_dir_name: str = "issues") -> str:
    """Generate a GFM table for the given issues."""
    if not issues:
        return "No open issues.\n"

    lines = []
    lines.append("| Name | Date | Status | Severity | Oneliner |")
    lines.append("|---|---|---|---|---|")
    
    for issue in issues:
        title = issue.title if issue.title else issue.filename
        date = issue.date if issue.date else "N/A"
        status = issue.status if issue.status else "Open"
        severity = issue.severity if issue.severity else "Unknown"
        summary = issue.summary if issue.summary else ""
        
        # Truncate summary for the table if it's too long
        if len(summary) > 80:
            summary = summary[:77] + "..."

        # Format the name as a link to the issue file
        name_link = f"[{title}]({issues_dir_name}/{issue.filename})"

        lines.append(f"| {name_link} | {date} | {status} | {severity} | {summary} |")

    return "\n".join(lines) + "\n"

def update_readme(issues_dir: str, root_dir: str, all_issues: list[Issue]) -> None:
    """Update the root README.md with a table of Open and In Progress issues."""
    readme_path = os.path.join(root_dir, "README.md")
    if not os.path.exists(readme_path):
        print(f"[error] Root README.md not found at {readme_path}", file=sys.stderr)
        sys.exit(1)

    # Filter for Open and In Progress issues
    active_issues = [
        i for i in all_issues
        if i.status.lower() in ("open", "in progress")
    ]
    
    issues_dir_name = os.path.basename(os.path.normpath(issues_dir))
    table_content = generate_markdown_table(active_issues, issues_dir_name)

    with open(readme_path, "r", encoding="utf-8") as f:
        content = f.read()

    header = "## Open Issues"
    header_idx = content.find(header)

    if header_idx != -1:
        # Header exists, replace everything from the header onwards, or until the next header
        
        # Find the end of the Open Issues section (either next header '## ' or EOF)
        start_search = header_idx + len(header)
        next_header_idx = content.find("\n## ", start_search)
        
        if next_header_idx != -1:
            # We have a next section
            before_section = content[:header_idx]
            after_section = content[next_header_idx:]
            new_content = before_section + header + "\n\n" + table_content + "\n" + after_section
        else:
            # Open Issues is the last section
            before_section = content[:header_idx]
            new_content = before_section + header + "\n\n" + table_content + "\n"
    else:
        # Header doesn't exist, append it
        if not content.endswith("\n\n"):
            if content.endswith("\n"):
                content += "\n"
            else:
                content += "\n\n"
        new_content = content + header + "\n\n" + table_content + "\n"

    with open(readme_path, "w", encoding="utf-8") as f:
        f.write(new_content)
        
    print(f"Successfully updated root README.md with {len(active_issues)} active issues.")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="List and search tracked issues in /workspace/issues/",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("--status",   metavar="STATUS",   help="Filter by status (open, 'in progress', resolved, 'wont fix')")
    parser.add_argument("--severity", metavar="SEVERITY", help="Filter by severity (critical, high, medium, low)")
    parser.add_argument("--search",   metavar="KEYWORD",  help="Search in title, summary, component, affects, ref")
    parser.add_argument("--full",     action="store_true", help="Print full file content for matching issues")
    parser.add_argument("--next-ref", action="store_true", help="Print the next available ISS-NNN and exit")
    parser.add_argument("--update-readme", action="store_true", help="Update the root README.md with a table of active issues")
    parser.add_argument("--dir",      metavar="PATH",     help="Override path to issues directory", default=None)
    args = parser.parse_args()

    issues_dir = os.path.realpath(args.dir if args.dir else ISSUES_DIR)
    all_issues = load_issues(issues_dir)

    # --next-ref: compute and print next available ISS-NNN then exit
    if args.next_ref:
        highest = max((i.ref_number() for i in all_issues), default=0)
        print(f"ISS-{highest + 1:03d}")
        return

    # --update-readme: inject active issues table into README.md and exit
    if args.update_readme:
        root_dir = os.path.realpath(os.path.join(issues_dir, ".."))
        update_readme(issues_dir, root_dir, all_issues)
        return

    filtered = [
        i for i in all_issues
        if i.matches_status(args.status)
        and i.matches_severity(args.severity)
        and i.matches_search(args.search)
    ]

    # Header
    total = len(all_issues)
    shown = len(filtered)
    filters = []
    if args.status:   filters.append(f"status={args.status}")
    if args.severity: filters.append(f"severity={args.severity}")
    if args.search:   filters.append(f"search='{args.search}'")
    filter_str = f"  [{', '.join(filters)}]" if filters else ""

    print()
    print(f"{BOLD}Issues{RESET}{filter_str}  —  {shown}/{total} shown")
    print("─" * 72)

    if not filtered:
        print(f"  {DIM}No issues match the given filters.{RESET}")
        print()
        return

    if args.full:
        for issue in filtered:
            print(f"\n{'═' * 72}")
            print(f"  {BOLD}{issue.filename}{RESET}")
            print(f"{'═' * 72}")
            print(issue.raw)
    else:
        for issue in filtered:
            print_summary_row(issue)

    # Status breakdown at the bottom (only when listing all)
    if not filters:
        status_counts: dict[str, int] = {}
        for issue in all_issues:
            key = issue.status.lower() or "unknown"
            status_counts[key] = status_counts.get(key, 0) + 1
        print("─" * 72)
        breakdown = "  ".join(f"{k}: {v}" for k, v in sorted(status_counts.items()))
        print(f"  {DIM}{breakdown}{RESET}\n")


if __name__ == "__main__":
    main()
