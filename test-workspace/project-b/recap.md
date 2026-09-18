# ATD Verification Recap

## Statistics
- **Internal Status:** INITIATED (CLI: true)
- **Instances Discovered:** 2
- **Implementation Gaps (No Test Proof):** 2
- **Test Optimization:** 1 runs cached
- **Status:** PENDING LLM AUDIT

## Discovery Table
| Atom ID | File Path | Line | Coverage | Audit Status | Recommendation |
|---|---|---|---|---|---|
| atom-b | src/lib.go | 3 | ⚠️ GAP | PENDING | Add @test-link |
| atom-a | src/lib.go | 4 | ⚠️ GAP | PENDING | Add @test-link |

## Actionable Prompt
Copy the content below into your LLM to perform the granular compliance audit.

--- BUNDLE START ---
<System Objective>
You are the ATD Lead Auditor. You must perform a granular, instance-based compliance check for each provided @spec-link.
For each instance, analyze the snippet against the rule and its ancestry (requirements context). Check if the provided tests adequately verify the logic.

### IMPORTANT: OUTPUT FORMAT
You MUST respond with a JSON array of objects, one per instance, following this schema exactly:
```json
[
  {
    "tag": "file:line",
    "atom_id": "string",
    "rule_compliant": boolean,
    "test_compliant": boolean,
    "findings": "string",
    "fix_suggestion": "string"
  }
]
```
</System Objective>

## INSTANCE 1: src/lib.go:3
### ATOM: atom-b
**Ancestry Context:** project-a:atom-a

#### THE RULE
Intent: Implementation in project B that depends on A.
Logic:


#### THE CODE (SURGICAL SNIPPET)
```
package lib

// @spec-link [[atom-b]]
// @spec-link [[project-a:atom-a]]
func Lib() {}

```

#### THE VERIFICATION PROOF (TESTS)
[WARNING] No @test-link found for this atom in the project.

#### NATIVE TEST EXECUTION
```
go: cannot find main module, but found .git/config in /home/bastien/work/atd
	to create a module there, run:
	cd ../.. && go mod init

```

---

## INSTANCE 2: src/lib.go:4
### ATOM: project-a:atom-a
**Ancestry Context:** 

#### THE RULE
Intent: Requirement for project A.
Logic:


#### THE CODE (SURGICAL SNIPPET)
```
package lib

// @spec-link [[atom-b]]
// @spec-link [[project-a:atom-a]]
func Lib() {}

```

#### THE VERIFICATION PROOF (TESTS)
[WARNING] No @test-link found for this atom in the project.

#### NATIVE TEST EXECUTION
```
go: cannot find main module, but found .git/config in /home/bastien/work/atd
	to create a module there, run:
	cd ../.. && go mod init

```

---

