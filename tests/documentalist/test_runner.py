"""Deterministic containment and oracle tests; these never invoke a model."""

import importlib.util
import json
import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

RUNNER = Path(__file__).with_name("runner.py")
SPEC = importlib.util.spec_from_file_location("documentalist_runner", RUNNER)
runner = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(runner)


def record(prompt, text=""):
    event = {"type": "tool_use", "part": {"state": {"status": "completed", "input": {"subagent_type": "documentalist", "prompt": prompt}, "output": "<task_result>" + text + "</task_result>"}}}
    return {"exit_code": 0, "timed_out": False, "stdout": json.dumps(event), "stderr": ""}


class DocumentalistHarnessTests(unittest.TestCase):
    # @test-link [[specification_atd_test_links]]
    def test_default_cli_skips_live_execution(self):
        completed = subprocess.run(["python3", str(RUNNER), "--all"], text=True, capture_output=True, check=True)
        payload = json.loads(completed.stdout)
        self.assertEqual("skipped", payload["status"])
        self.assertIn("DOCUMENTALIST_LIVE", payload["reason"])

    # @test-link [[service_atd_lint]]
    def test_fixture_is_valid_independent_and_has_no_remote(self):
        scenario = {"id": "missing-link", "faults": False}
        fixture, _ = runner.build_fixture(scenario)
        try:
            self.assertNotEqual(runner.SOURCE.resolve(), fixture.resolve())
            self.assertIsInstance(json.loads((fixture / ".atd").read_text()), dict)
            self.assertEqual(0, runner.command([shutil.which("atd") or "atd", "lint"], fixture)["exit_code"])
            remote = subprocess.run(["git", "remote"], cwd=fixture, text=True, capture_output=True, check=True)
            self.assertEqual("", remote.stdout)
        finally:
            shutil.rmtree(fixture)

    def test_invalid_fixture_setup_fails_loudly(self):
        with self.assertRaises(runner.FixtureError):
            runner.build_fixture({"id": "preflight", "faults": False}, invalid=True)

    def test_primary_driver_is_discovered_without_replacing_documentalist(self):
        fixture, _ = runner.build_fixture({"id": "preflight", "faults": False})
        home = Path(tempfile.mkdtemp(prefix="documentalist-test-home-"))
        try:
            env = runner.evaluation_env(fixture / ".documentalist-bin", home)
            listed = subprocess.run(["opencode", "agent", "list"], cwd=fixture, env=env, text=True, capture_output=True, check=True)
            self.assertIn("documentalist-eval-driver (primary)", listed.stdout)
            self.assertIn("documentalist (subagent)", listed.stdout)
        finally:
            shutil.rmtree(fixture)
            shutil.rmtree(home)

    def test_isolated_home_preserves_open_code_config_and_data_roots(self):
        isolated = Path("/tmp/documentalist-isolated-home")
        env = runner.evaluation_env(Path("/tmp/documentalist-shims"), isolated, {
            "HOME": "/original-home", "XDG_CONFIG_HOME": "/kept-config", "XDG_DATA_HOME": "/kept-data", "PATH": "/bin",
        })
        self.assertEqual(str(isolated), env["HOME"])
        self.assertEqual("/kept-config", env["XDG_CONFIG_HOME"])
        self.assertEqual("/kept-data", env["XDG_DATA_HOME"])
        self.assertEqual("1", env["OPENCODE_DISABLE_EXTERNAL_SKILLS"])
        self.assertEqual("1", env["OPENCODE_DISABLE_CLAUDE_CODE_SKILLS"])

    # @test-link [[api_atd_serve_map]]
    # @test-link [[service_atd_search]]
    def test_fault_shim_logs_map_search_and_forwards_other_commands(self):
        fixture, shims = runner.build_fixture({"id": "tool-fallback", "faults": True})
        try:
            env = os.environ.copy()
            env["ATD_REAL"] = shutil.which("atd") or "atd"
            env["DOCUMENTALIST_FAULT_LOG"] = str(shims / "fault.log")
            env["PATH"] = str(shims) + os.pathsep + env["PATH"]
            self.assertEqual(65, subprocess.run(["atd", "map", "--file", "src/friendly_fire.go"], cwd=fixture, env=env).returncode)
            self.assertEqual(0, subprocess.run(["atd", "search", "--grep", "friendly"], cwd=fixture, env=env).returncode)
            self.assertEqual(0, subprocess.run(["atd", "lint"], cwd=fixture, env=env).returncode)
            self.assertEqual(["map", "search"], runner.fault_invocations(fixture))
        finally:
            shutil.rmtree(fixture)

    def test_audit_runtime_artifacts_are_ignored_but_normal_files_are_not(self):
        fixture, shims = runner.build_fixture({"id": "drift", "faults": False})
        try:
            before = runner.tree_snapshot(fixture)
            (shims / "atd_audit.db").write_text("runtime cache")
            self.assertEqual([], runner.changed_files(before, runner.tree_snapshot(fixture)))
            (fixture / "agent-edit.txt").write_text("observable")
            self.assertEqual(["agent-edit.txt"], runner.changed_files(before, runner.tree_snapshot(fixture)))
        finally:
            shutil.rmtree(fixture)

    def test_canonical_snapshot_detects_root_and_nested_git_mutations(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            nested = root / "nested"
            nested.mkdir()
            for repo, name in ((root, "root.txt"), (nested, "nested.txt")):
                subprocess.run(["git", "init", "-q"], cwd=repo, check=True)
                subprocess.run(["git", "config", "user.email", "eval@example.invalid"], cwd=repo, check=True)
                subprocess.run(["git", "config", "user.name", "Evaluation"], cwd=repo, check=True)
                (repo / name).write_text("baseline")
                subprocess.run(["git", "add", name], cwd=repo, check=True)
                subprocess.run(["git", "commit", "-qm", "baseline"], cwd=repo, check=True)
            baseline = runner.canonical_git_snapshot(root)
            (root / "root.txt").write_text("unstaged change")
            self.assertNotEqual(baseline, runner.canonical_git_snapshot(root))
            (root / "root.txt").write_text("staged change")
            subprocess.run(["git", "add", "root.txt"], cwd=root, check=True)
            self.assertNotEqual(baseline, runner.canonical_git_snapshot(root))
            untracked = root / "untracked.txt"
            untracked.write_text("one")
            with_untracked = runner.canonical_git_snapshot(root)
            untracked.write_text("two")
            self.assertNotEqual(with_untracked, runner.canonical_git_snapshot(root))
            untracked.unlink()
            self.assertNotEqual(with_untracked, runner.canonical_git_snapshot(root))
            (nested / "nested.txt").write_text("nested change")
            self.assertNotEqual(baseline, runner.canonical_git_snapshot(root))

    # @test-link [[service_atd_check]]
    def test_preexisting_drift_is_not_blamed_on_agent(self):
        scenario = {"id": "drift", "faults": False, "required_output": ["Drift/conflict flagged", "rule_friendly_fire"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            report = "Drift/conflict flagged (unresolved): rule_friendly_fire differs at src/friendly_fire.go"
            failures, changed = runner.check_oracle(scenario, fixture, record("p", report), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertEqual([], changed)
            self.assertFalse(any("agent changed fixture" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_agent_mutation_is_caught_against_pre_execution_snapshot(self):
        scenario = {"id": "drift", "faults": False, "required_output": ["Drift/conflict flagged", "rule_friendly_fire"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            before = runner.tree_snapshot(fixture)
            (fixture / "docs/rule_friendly_fire.atom.md").write_text("unsafe")
            report = "Drift/conflict flagged (unresolved): rule_friendly_fire differs at src/friendly_fire.go"
            failures, _ = runner.check_oracle(scenario, fixture, record("p", report), runner.canonical_git_snapshot(), before)
            self.assertTrue(any("agent changed fixture" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_disallowed_preflight_verdict_is_caught(self):
        scenario = {"id": "preflight", "faults": False, "verdicts": ["PROCEED"], "required_atoms": ["rule_friendly_fire", "spec_match_format"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            failures, _ = runner.check_oracle(scenario, fixture, record("p", "Verdict: HALT-NEEDS-USER-INPUT\nrule_friendly_fire spec_match_format"), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("scenario-allowed" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_prose_verdict_mention_without_anchored_field_is_caught(self):
        scenario = {"id": "preflight", "faults": False, "verdicts": ["PROCEED"], "required_atoms": ["rule_friendly_fire", "spec_match_format"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            failures, _ = runner.check_oracle(scenario, fixture, record("p", "PROCEED is mentioned in prose; rule_friendly_fire spec_match_format"), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("scenario-allowed" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_structured_none_drift_is_caught(self):
        scenario = {"id": "drift", "faults": False, "required_output": []}
        fixture, _ = runner.build_fixture(scenario)
        try:
            failures, _ = runner.check_oracle(scenario, fixture, record("p", "Drift/conflict flagged (unresolved): None"), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("substantive structured drift" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    # @test-link [[service_atd_update]]
    def test_wrong_missing_link_placement_is_caught(self):
        scenario = {"id": "missing-link", "faults": False, "required_output": ["Links added", "rule_friendly_fire"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            before = runner.tree_snapshot(fixture)
            (fixture / "src/friendly_fire.go").write_text("// @spec-link [[rule_friendly_fire]]\n" + runner.fixture_code("missing-link"))
            failures, _ = runner.check_oracle(scenario, fixture, record("p", "Links added rule_friendly_fire"), runner.canonical_git_snapshot(), before)
            self.assertTrue(any("directly above MayDamage" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_missing_fault_shim_invocation_is_caught(self):
        scenario = {"id": "tool-fallback", "faults": True, "required_output": ["ATD Sync Report"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            failures, _ = runner.check_oracle(scenario, fixture, record("p", "ATD Sync Report untrusted malformed stale direct"), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("fault shim" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_fault_prompt_echo_fails_even_when_both_shims_ran(self):
        scenario = {"id": "tool-fallback", "faults": True, "required_output": ["ATD Sync Report"]}
        fixture, shims = runner.build_fixture(scenario)
        try:
            (shims / "fault.log").write_text("map\nsearch\n")
            echoed = "Run map malformed and search stale; use direct evidence."
            failures, _ = runner.check_oracle(scenario, fixture, record("p", echoed), runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertFalse(any("fault shim did not" in item for item in failures))
            self.assertTrue(any("Tool-evidence warning" in item for item in failures))
            self.assertTrue(any("Verification" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_missing_documentalist_delegation_evidence_is_caught(self):
        scenario = {"id": "drift", "faults": False, "required_output": ["Drift/conflict flagged", "rule_friendly_fire"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            plain = {"exit_code": 0, "timed_out": False, "stdout": "Drift/conflict flagged rule_friendly_fire", "stderr": ""}
            failures, _ = runner.check_oracle(scenario, fixture, plain, runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("delegation" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_errored_task_prompt_echo_cannot_satisfy_fallback_oracle(self):
        scenario = {"id": "tool-fallback", "faults": True, "required_output": ["ATD Sync Report"]}
        fixture, _ = runner.build_fixture(scenario)
        try:
            # The prompt contains every label the old raw-stdout oracle accepted.
            error = {"type": "tool_use", "part": {"state": {"status": "error", "input": {"subagent_type": "documentalist", "prompt": "ATD Sync Report not trusted malformed stale direct"}, "error": "Subagent failed: user rejected permission"}}}
            final = {"type": "text", "part": {"type": "text", "text": "Subagent failed: user rejected permission"}}
            failed = {"exit_code": 0, "timed_out": False, "stdout": json.dumps(error) + "\n" + json.dumps(final), "stderr": ""}
            failures, _ = runner.check_oracle(scenario, fixture, failed, runner.canonical_git_snapshot(), runner.tree_snapshot(fixture))
            self.assertTrue(any("delegation failed" in item for item in failures))
            self.assertTrue(any("final assistant" in item for item in failures))
            self.assertTrue(any("required output" in item for item in failures))
            self.assertTrue(any("Tool-evidence warning" in item for item in failures))
        finally:
            shutil.rmtree(fixture)

    def test_git_guard_denies_push(self):
        fixture, shims = runner.build_fixture({"id": "preflight", "faults": False})
        try:
            env = os.environ.copy()
            env["PATH"] = str(shims) + os.pathsep + env["PATH"]
            denied = subprocess.run(["git", "push"], cwd=fixture, env=env, text=True, capture_output=True)
            self.assertEqual(97, denied.returncode)
            self.assertIn("push denied", denied.stderr)
        finally:
            shutil.rmtree(fixture)


if __name__ == "__main__":
    unittest.main()
