package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ServeStdio runs the MCP server over stdin/stdout.
// This is the primary transport — compatible with VS Code, Claude Desktop,
// and any MCP host that launches the server as a subprocess.
// Blocking — returns only on EOF or read error.
func (r *Registry) ServeStdio() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := r.Handle(line)
		if resp == nil {
			// Notification — no reply.
			continue
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "[mcp/stdio] encode error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
