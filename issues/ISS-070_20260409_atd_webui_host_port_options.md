# Issue: WebUI Host and Port Configuration

**ID:** `20260409_atd_webui_host_port_options`
**Ref:** `ISS-070`
**Date:** 2026-04-09
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd/cmd/webui.go`
**Affects:** `scripts/pkg/webui/server.go`

---

## Summary

The `atd webui` command currently lacks the ability to specify the listening host and port via command-line flags. While the configuration structure supports these fields, they are not exposed to the user at runtime, forcing a default of `:8080` or requiring manual `.atd` config edits.

---

## Technical Description

### Background
The ATD WebUI is started via the `atd webui` command. It initializes a Gin-based server that reads its host and port settings from the `ActiveConfig.WebUI` structure.

### The Problem Scenario
A user wanting to run the WebUI on a specific interface (e.g., `0.0.0.0` for remote access) or a different port (e.g., to avoid conflicts) cannot easily do so without modifying the `.atd` configuration file. Most CLI tools provide flags for these common overrides.

Current implementation in `scripts/cmd/atd/cmd/webui.go`:
```go
func init() {
	rootCmd.AddCommand(webuiCmd)
	webuiCmd.Flags().Bool("dev", false, "Enable development mode (serve from filesystem instead of embed)")
	webuiCmd.Flags().String("static-path", "", "Override path to static files (useful for local development)")
}
```

Current implementation in `scripts/pkg/webui/server.go`:
```go
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", config.ActiveConfig.WebUI.Host, config.ActiveConfig.WebUI.Port)
	if config.ActiveConfig.WebUI.Port == 0 {
		addr = ":8080"
	}
	fmt.Printf("WebUI server starting on http://%s (DevMode: %v)\n", addr, s.DevMode)
	return s.Engine.Run(addr)
}
```

### Where This Pattern Exists Today
- [webui.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/webui.go) (Missing flags)
- [server.go](file:///home/bastien/work/skill/scripts/pkg/webui/server.go#L66-L73) (Static address logic)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High — users find they cannot change the port |
| Current mitigant | Modifying the `.atd` JSON config file manually |

---

## Recommended Fix

**Short term:** Add `--host` and `--port` flags to the `webui` command in `scripts/cmd/atd/cmd/webui.go`.  
**Medium term:** Update `webui.go` to override `config.ActiveConfig.WebUI` values if flags are provided.  
**Long term:** Ensure all server-like commands (like `mcp serve` if it exists) follow a consistent host/port override pattern.

---

## References

- [scripts/cmd/atd/cmd/webui.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/webui.go)
- [scripts/pkg/webui/server.go](file:///home/bastien/work/skill/scripts/pkg/webui/server.go)
