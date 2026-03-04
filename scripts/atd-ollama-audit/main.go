package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"atd-tools/config"
)

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func queryLlama(prompt string) (string, int, int, error) {
	reqBody := GenerateRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Stream: false,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, 0, err
	}

	fmt.Println("--- Sending request to Ollama (llama3.2) ---")
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, 0, err
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		return "", 0, 0, err
	}
	return genResp.Response, genResp.PromptEvalCount, genResp.EvalCount, nil
}

func main() {
	var atomPath string
	var codeContent string

	flag.StringVar(&atomPath, "atom", "", "Path to the Atom Markdown file containing the Rule")
	flag.StringVar(&codeContent, "code", "", "Snippet of code returned from the semantic search")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-ollama-audit", "Started process")


	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if atomPath == "" || codeContent == "" {
		fmt.Println("Error: -atom and -code parameters are required.")
		os.Exit(1)
	}

	atomContent, err := os.ReadFile(atomPath)
	if err != nil {
		fmt.Printf("Failed to read %s\n", atomPath)
		os.Exit(1)
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Auditor strictly verifying deterministic implementation against stated rules. You must output JSON format only: {"passed": boolean, "resolutionMessage": string}
</System Objective>

<Rule to Validate>
%s
</Rule>

<Target Source Code>
%s
</Target Source Code>

Analyze the <Target Source Code> strictly through the constraints defined in <Rule>. Does the code perfectly satisfy the rule?
`, string(atomContent), codeContent)

	response, promptTokens, evalTokens, err := queryLlama(prompt)
	if err != nil {
		fmt.Printf("Error querying local LLM: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Local Auditor Result (`llama3.2`) ---")
	fmt.Println(response)
	fmt.Printf("\n[TOKEN USAGE] Prompt: %d | Eval: %d | Total: %d\n", promptTokens, evalTokens, promptTokens+evalTokens)
}
