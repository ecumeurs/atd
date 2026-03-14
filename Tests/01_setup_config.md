# Task 01 — Setup: atd init + Smoke Test

## Objective
Initialize the ATD framework in `upsilonbattle/` and verify the binary works correctly.

## Prerequisites
- `atd` binary is compiled and on PATH: `cd /home/bastien/work/skill/scripts/cmd/atd && go build -o /usr/local/bin/atd .`

## Steps

### 1. Initialize ATD config in upsilonbattle
```bash
atd init --dir /home/bastien/work/skill/upsilonbattle
```

### 2. Verify the config was created
```bash
cat /home/bastien/work/skill/upsilonbattle/.atd
```

### 3. Verify the docs directory was created
```bash
ls /home/bastien/work/skill/upsilonbattle/docs/
```

### 4. Smoke test: binary help
```bash
atd --help
atd init --help
atd serve --help
```

### 5. Smoke test: query in the skill project (has existing ATDs)
```bash
cd /home/bastien/work/skill && atd query --search atd_cli
```

## Expected Output
- `.atd` file created at `upsilonbattle/.atd` with valid JSON (provider chain, model routing)
- `upsilonbattle/docs/` directory exists (empty)
- `atd --help` lists all subcommands including `init` and `serve`
- `atd query --search atd_cli` returns the `atd_cli` atom JSON

## Acceptance Criteria
- [ ] `.atd` exists and is valid JSON
- [ ] `docs/` directory created
- [ ] `atd --help` shows `init` and `serve` subcommands
- [ ] `atd query` returns results from `skill/docs/`
