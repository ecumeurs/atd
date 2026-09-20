#!/usr/bin/env python3
"""Opt-in behavioral evaluation for the installed documentalist OpenCode agent."""

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import tempfile
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "upsilon-hub"
SCENARIO_FILE = Path(__file__).with_name("scenarios.json")
LIVE_ENV = "DOCUMENTALIST_LIVE"
TIMEOUT_SECONDS = 180
SANDBOX_ROOT = Path("/tmp/opencode")
SOURCE_INPUTS = (
    ".atd",
    "docs/contract_upsilon_contract.atom.md",
    "docs/vision_upsilon_vision.atom.md",
    "docs/spec_match_format.atom.md",
    "docs/rule_friendly_fire.atom.md",
)


class FixtureError(RuntimeError):
    pass


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_snapshot():
    return {name: digest(SOURCE / name) for name in SOURCE_INPUTS}


def git_output(repo, args):
    return subprocess.run(["git", *args], cwd=repo, check=True, capture_output=True).stdout


def git_repositories(root):
    for directory, dirs, _ in os.walk(root):
        candidate = Path(directory)
        git_dir = candidate / ".git"
        if git_dir.exists() or git_dir.is_file():
            yield candidate
            dirs[:] = [name for name in dirs if name != ".git"]


def canonical_git_snapshot(root=SOURCE):
    """Capture every direct nested Git worktree without changing it."""
    snapshot = {}
    for repo in git_repositories(root):
        relative = repo.relative_to(root).as_posix() or "."
        head = git_output(repo, ["rev-parse", "HEAD"]).decode().strip()
        porcelain = git_output(repo, ["status", "--porcelain=v1", "--untracked-files=all"])
        unstaged = git_output(repo, ["diff", "--binary"])
        staged = git_output(repo, ["diff", "--cached", "--binary"])
        untracked = {}
        for raw_path in git_output(repo, ["ls-files", "--others", "--exclude-standard", "-z"]).split(b"\0"):
            if raw_path:
                name = raw_path.decode("utf-8", "surrogateescape")
                path = repo / name
                # Git reports a nested repository as one untracked directory;
                # its own repository snapshot below captures its contents.
                untracked[name] = "<directory>" if path.is_dir() else digest(path)
        snapshot[relative] = {
            "head": head,
            "porcelain": hashlib.sha256(porcelain).hexdigest(),
            "unstaged": hashlib.sha256(unstaged).hexdigest(),
            "staged": hashlib.sha256(staged).hexdigest(),
            "untracked": untracked,
        }
    return snapshot


def command(argv, cwd, env=None, timeout=TIMEOUT_SECONDS):
    started = time.monotonic()
    try:
        completed = subprocess.run(argv, cwd=cwd, env=env, text=True, capture_output=True, timeout=timeout)
        return {"argv": argv, "exit_code": completed.returncode, "stdout": completed.stdout,
                "stderr": completed.stderr, "timed_out": False, "seconds": round(time.monotonic() - started, 3)}
    except subprocess.TimeoutExpired as error:
        return {"argv": argv, "exit_code": None, "stdout": error.stdout or "", "stderr": error.stderr or "",
                "timed_out": True, "seconds": round(time.monotonic() - started, 3)}


def tree_snapshot(root):
    ignored = {".git", ".documentalist-bin/fault.log"}
    snapshot = {}
    for path in root.rglob("*"):
        relative = path.relative_to(root).as_posix()
        runtime_file = (
            relative in {".opencode/.gitignore", ".opencode/package.json", ".opencode/package-lock.json"}
            or relative == "docs/.atd_audit.db"
            or relative == ".documentalist-bin/atd_audit.db"
            or relative.startswith(".opencode/node_modules/")
            or relative.startswith(".documentalist-bin/pipeline_output/")
        )
        if path.is_file() and relative not in ignored and not relative.startswith(".git/") and not runtime_file:
            snapshot[relative] = digest(path)
    return snapshot


def append_expectation(path):
    text = path.read_text()
    if "## EXPECTATION" not in text:
        text += "\n## EXPECTATION\n- The documented rule remains structurally valid in this disposable evaluation fixture.\n"
    elif text.rstrip().endswith("## EXPECTATION"):
        text += "- The documented rule remains structurally valid in this disposable evaluation fixture.\n"
    path.write_text(text)


def fixture_code(mode):
    link = "// @spec-link [[rule_friendly_fire]]\n" if mode != "missing-link" else ""
    body = "return attacker_team != target_team\n"
    if mode == "drift":
        body = "return true // deliberate drift: same-team damage is allowed\n"
    return "package friendly\n\n" + link + "func MayDamage(attacker_team, target_team int) bool {\n\t" + body + "}\n"


def write_driver(fixture):
    agent = fixture / ".opencode/agent/documentalist-eval-driver.md"
    agent.parent.mkdir(parents=True)
    agent.write_text("""---
description: Primary evaluation driver that delegates one request to documentalist.
mode: primary
permission:
  task: allow
  edit: deny
  bash: deny
---

Delegate the user's prompt exactly once, byte-for-byte and without wrapping it in
quotes, to the installed `documentalist` subagent using the task tool. Do not inspect
files, call any other tool, answer the request yourself, or add commentary. Return
the task result verbatim.
""")


def write_shims(directory, faults):
    directory.mkdir()
    (directory / "git").write_text("#!/bin/sh\nif [ \"$1\" = push ]; then echo 'DOCUMENTALIST_GIT_GUARD: push denied' >&2; exit 97; fi\nexec /usr/bin/git \"$@\"\n")
    if faults:
        (directory / "atd").write_text("""#!/bin/sh
case "$1" in
  map) printf 'map\n' >> "$DOCUMENTALIST_FAULT_LOG"; echo '{malformed map evidence' >&2; exit 65 ;;
  search) printf 'search\n' >> "$DOCUMENTALIST_FAULT_LOG"; echo 'STALE: deleted/path/bridge_start.go PropertyToString(effectiveKey)'; exit 0 ;;
esac
exec "$ATD_REAL" "$@"
""")
    for entry in directory.iterdir():
        entry.chmod(0o755)


def validate_fixture(fixture):
    try:
        config = json.loads((fixture / ".atd").read_text())
    except (OSError, json.JSONDecodeError) as error:
        raise FixtureError("fixture .atd is not valid JSON: " + str(error)) from error
    if not isinstance(config, dict):
        raise FixtureError("fixture .atd must contain a JSON object")
    config_result = command([shutil.which("atd") or "atd", "config", "list"], fixture, timeout=30)
    if config_result["exit_code"] != 0:
        raise FixtureError("fixture atd config validation failed: " + config_result["stdout"] + config_result["stderr"])
    result = command([shutil.which("atd") or "atd", "lint"], fixture, timeout=30)
    if result["exit_code"] != 0:
        raise FixtureError("fixture atd lint failed: " + result["stdout"] + result["stderr"])


def build_fixture(scenario, invalid=False):
    fixture = Path(tempfile.mkdtemp(prefix="documentalist-" + scenario["id"] + "-", dir=SANDBOX_ROOT))
    try:
        for relative in SOURCE_INPUTS:
            target = fixture / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(SOURCE / relative, target)
            if target.name.endswith(".atom.md"):
                append_expectation(target)
        if invalid:
            (fixture / "docs/rule_friendly_fire.atom.md").write_text("---\nid: broken\n---\n")
        (fixture / "src").mkdir()
        (fixture / "src/friendly_fire.go").write_text(fixture_code("preflight"))
        write_driver(fixture)
        config = {"$schema": "https://opencode.ai/config.json"}
        if scenario["faults"]:
            config["mcp"] = {"atd": {"enabled": False}}
        (fixture / "opencode.json").write_text(json.dumps(config) + "\n")
        shims = fixture / ".documentalist-bin"
        write_shims(shims, scenario["faults"])
        (shims / "pipeline_output").mkdir()
        (fixture / "pipeline_output").symlink_to(shims / "pipeline_output", target_is_directory=True)
        (fixture / "docs/.atd_audit.db").symlink_to(shims / "atd_audit.db")
        validate_fixture(fixture)
        subprocess.run(["git", "init", "-q"], cwd=fixture, check=True, capture_output=True)
        subprocess.run(["git", "config", "user.email", "documentalist-eval@example.invalid"], cwd=fixture, check=True)
        subprocess.run(["git", "config", "user.name", "Documentalist Evaluation"], cwd=fixture, check=True)
        subprocess.run(["git", "add", "."], cwd=fixture, check=True)
        subprocess.run(["git", "commit", "-qm", "fixture baseline"], cwd=fixture, check=True)
        if scenario["id"] in {"drift", "missing-link"}:
            (fixture / "src/friendly_fire.go").write_text(fixture_code(scenario["id"]))
            subprocess.run(["git", "add", "src/friendly_fire.go"], cwd=fixture, check=True)
        return fixture, shims
    except Exception:
        shutil.rmtree(fixture, ignore_errors=True)
        raise


def transcript_events(record):
    events = []
    for line in record["stdout"].splitlines():
        try:
            events.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return events


def documentalist_result(record, prompt):
    matches = []
    for event in transcript_events(record):
        state = event.get("part", {}).get("state", {})
        if event.get("type") == "tool_use" and state.get("input", {}).get("subagent_type") == "documentalist":
            matches.append(state)
    if len(matches) != 1:
        return "", "expected exactly one documentalist task call"
    state = matches[0]
    if state.get("status") != "completed" or state.get("error"):
        return "", "documentalist task was rejected or failed: " + str(state.get("error", state.get("status")))
    delegated = state.get("input", {}).get("prompt")
    # OpenCode serializes a task prompt as a JSON string in some providers.
    # Accept only that transport wrapper; any substantive rewrite still fails.
    if delegated != prompt and delegated != '"' + prompt + '"':
        return "", "documentalist task prompt was rewritten"
    output = state.get("output")
    if not isinstance(output, str):
        return "", "documentalist task returned no result"
    match = re.search(r"<task_result>\s*(.*?)\s*</task_result>", output, re.DOTALL)
    if not match or not match.group(1).strip():
        return "", "documentalist task returned an empty result"
    return match.group(1).strip(), ""


def final_assistant_text(record):
    texts = []
    for event in transcript_events(record):
        part = event.get("part", {})
        if event.get("type") == "text" and part.get("type") == "text":
            texts.append(part.get("text", ""))
    return "\n".join(texts)


def changed_files(before, after):
    return sorted(path for path in set(before) | set(after) if before.get(path) != after.get(path))


def fault_invocations(fixture):
    log = fixture / ".documentalist-bin/fault.log"
    return log.read_text().splitlines() if log.exists() else []


VERDICT_FIELD = re.compile(r"^\s*(?:\*\*)?Verdict(?: \(preflight only\))?(?:\*\*)?:\s*(PROCEED-WITH-SIGNOFF-PENDING|HALT-NEEDS-CONTRACT-VISION-DECISION|HALT-NEEDS-USER-INPUT|PROCEED)\s*(?:\*\*)?\s*$", re.MULTILINE)
DRIFT_FIELD = re.compile(r"^\s*(?:[-*]\s*)?(?:\*\*)?Drift/conflict flagged \(unresolved\)(?:\*\*)?:\s*(.+)$", re.MULTILINE)
SYNC_REPORT_HEADING = re.compile(r"^\s*\*\*ATD Sync Report\*\*", re.MULTILINE)
TOOL_WARNING_FIELD = re.compile(r"^\s*(?:[-*]\s*)?(?:\*\*)?Tool-evidence warning(?:\*\*)?:\s*(.+)$", re.MULTILINE)
VERIFICATION_FIELD = re.compile(r"^\s*(?:[-*]\s*)?(?:\*\*)?Verification(?:\*\*)?:\s*(.+)$", re.MULTILINE)


def report_verdict(text):
    match = VERDICT_FIELD.search(text)
    return match.group(1) if match else ""


def report_drift(text):
    match = DRIFT_FIELD.search(text)
    return match.group(1).strip() if match else ""


def fallback_report_failures(text):
    failures = []
    if not SYNC_REPORT_HEADING.search(text):
        failures.append("missing structured ATD Sync Report heading")
    warning = TOOL_WARNING_FIELD.search(text)
    if not warning:
        failures.append("missing structured Tool-evidence warning field")
    else:
        value = warning.group(1).lower()
        if not re.search(r"map.{0,100}malformed", value) or not re.search(r"search.{0,100}stale", value):
            failures.append("Tool-evidence warning does not identify malformed map and stale search")
        if not re.search(r"untrusted|not used|discarded|not relied", value):
            failures.append("Tool-evidence warning does not discard untrusted tool outputs")
    verification = VERIFICATION_FIELD.search(text)
    if not verification:
        failures.append("missing structured Verification field")
    else:
        value = verification.group(1).lower()
        if not re.search(r"direct|on-disk|file inspection", value) or not re.search(r"used|confirm|inspect|evidence", value):
            failures.append("Verification does not state direct/on-disk/file evidence was used")
    return failures


def evaluation_env(shims, isolated_home, parent_env=None):
    """Keep OpenCode's installed config/data while isolating HOME discovery."""
    env = dict(os.environ if parent_env is None else parent_env)
    original_home = Path(env.get("HOME", str(Path.home())))
    env["HOME"] = str(isolated_home)
    env["XDG_CONFIG_HOME"] = env.get("XDG_CONFIG_HOME", str(original_home / ".config"))
    env["XDG_DATA_HOME"] = env.get("XDG_DATA_HOME", str(original_home / ".local" / "share"))
    env["ATD_REAL"] = shutil.which("atd") or "atd"
    env["DOCUMENTALIST_FAULT_LOG"] = str(shims / "fault.log")
    env["OPENCODE_DISABLE_EXTERNAL_SKILLS"] = "1"
    env["OPENCODE_DISABLE_CLAUDE_CODE_SKILLS"] = "1"
    env["PATH"] = str(shims) + os.pathsep + env["PATH"]
    return env


def check_oracle(scenario, fixture, record, source_before, fixture_before):
    fixture_after = tree_snapshot(fixture)
    changed = changed_files(fixture_before, fixture_after)
    failures = []
    if record["timed_out"] or record["exit_code"] != 0:
        failures.append("OpenCode command timed out or returned nonzero")
    if canonical_git_snapshot() != source_before:
        failures.append("canonical upsilon-hub Git snapshot changed")
    if "Falling back to default agent" in record["stderr"]:
        failures.append("OpenCode fell back to the default agent")
    text, delegation_failure = documentalist_result(record, scenario.get("prompt", "p"))
    if delegation_failure:
        failures.append("documentalist delegation failed: " + delegation_failure)
    final_text = final_assistant_text(record).lower()
    if any(marker in final_text for marker in ("subagent failed", "task failed", "user rejected permission")):
        failures.append("final assistant content reports a subagent failure")
    if scenario["id"] == "missing-link":
        expected = fixture_code("preflight")
        if (fixture / "src/friendly_fire.go").read_text() != expected:
            failures.append("missing-link repair was not exactly one comment directly above MayDamage")
        if changed != ["src/friendly_fire.go"]:
            failures.append("missing-link repair changed files other than the allowed source comment")
    elif changed:
        failures.append("agent changed fixture files: " + ", ".join(changed))
    if scenario["id"] == "preflight":
        verdict = report_verdict(text)
        if verdict not in scenario["verdicts"]:
            failures.append("missing scenario-allowed preflight verdict")
        for atom in scenario["required_atoms"]:
            if atom not in text:
                failures.append("missing governing atom " + atom)
    elif scenario["id"] == "drift":
        drift = report_drift(text)
        lowered = drift.lower()
        if not drift or lowered in {"none", "n/a", "na"} or lowered.startswith("none "):
            failures.append("missing substantive structured drift finding")
        elif scenario.get("drift_atom", "rule_friendly_fire").lower() not in lowered or scenario.get("drift_source", "src/friendly_fire.go").lower() not in lowered:
            failures.append("structured drift finding lacks expected atom or source target")
    else:
        for label in scenario["required_output"]:
            if label.lower() not in text.lower():
                failures.append("missing required output label: " + label)
    if scenario["faults"]:
        invoked = fault_invocations(fixture)
        if sorted(invoked) != ["map", "search"]:
            failures.append("fault shim did not record both map and search invocations")
        failures.extend(fallback_report_failures(text))
    if "DOCUMENTALIST_GIT_GUARD" in record["stdout"] + record["stderr"]:
        failures.append("agent attempted a forbidden git push")
    return failures, changed


# @test-link [[domain_atd_usage_protocol]]
def run_scenario(scenario, keep=False):
    source_before = canonical_git_snapshot()
    fixture, shims = build_fixture(scenario)
    fixture_before = tree_snapshot(fixture)
    isolated_home = Path(tempfile.mkdtemp(prefix="documentalist-home-", dir=SANDBOX_ROOT))
    env = evaluation_env(shims, isolated_home)
    argv = ["opencode", "run", "--agent", "documentalist-eval-driver", "--format", "json", "--dir", str(fixture)]
    if scenario["allow_auto"]:
        argv.append("--auto")
    argv.append(scenario["prompt"])
    try:
        record = command(argv, fixture, env)
    finally:
        shutil.rmtree(isolated_home, ignore_errors=True)
    failures, changed = check_oracle(scenario, fixture, record, source_before, fixture_before)
    result = {"scenario": scenario["id"], "status": "passed" if not failures else "failed", "failures": failures,
              "changes": changed, "fixture": str(fixture), "command": record}
    if failures:
        (fixture / "documentalist-result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
        (fixture / "documentalist-transcript.json").write_text(record["stdout"])
    elif not keep:
        shutil.rmtree(fixture)
        result["fixture"] = None
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    selection = parser.add_mutually_exclusive_group()
    selection.add_argument("--scenario")
    selection.add_argument("--all", action="store_true")
    parser.add_argument("--live", action="store_true", help="allow model-backed OpenCode execution")
    parser.add_argument("--keep", action="store_true", help="retain successful disposable fixtures")
    args = parser.parse_args()
    scenarios = json.loads(SCENARIO_FILE.read_text())
    chosen = [item for item in scenarios if not args.scenario or item["id"] == args.scenario]
    if args.scenario and not chosen:
        parser.error("unknown scenario: " + args.scenario)
    if not args.live and os.environ.get(LIVE_ENV) != "1":
        print(json.dumps({"status": "skipped", "reason": "live execution disabled; pass --live or set " + LIVE_ENV + "=1",
                          "scenarios": [item["id"] for item in chosen]}, sort_keys=True))
        return 0
    results = [run_scenario(item, args.keep) for item in chosen]
    print(json.dumps(results, indent=2, sort_keys=True))
    return 0 if all(item["status"] == "passed" for item in results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
