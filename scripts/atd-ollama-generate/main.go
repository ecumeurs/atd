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
	Format string `json:"format,omitempty"`
}

type GenerateResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func queryLlama(prompt string, useJsonFormat bool, model string) (string, error) {
	reqBody := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	if useJsonFormat {
		reqBody.Format = "json"
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		return "", err
	}
	return genResp.Response, nil
}

func main() {
	var dissectFile string
	flag.StringVar(&dissectFile, "dissect", "", "Path to the atd-dissect output file (prompt)")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if dissectFile == "" {
		fmt.Println("Error: -dissect parameter is required.")
		os.Exit(1)
	}

	config.Load()
	config.Log("atd-ollama-generate", "Starting atom generation bounds extraction")

	promptData, err := os.ReadFile(dissectFile)
	if err != nil {
		fmt.Printf("Failed to read %s\n", dissectFile)
		os.Exit(1)
	}

	fmt.Println("--- Asking llama3.2 to extract atom boundaries ---")

	modelToUse := config.ActiveConfig.Model
	if modelToUse == "" {
		modelToUse = "llama3.2" // fallback
	}

	response, err := queryLlama(string(promptData), true, modelToUse) // Force JSON format
	if err != nil {
		fmt.Printf("Error querying local LLM: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Local Llama Boundary Result ---")
	fmt.Println(response)
	config.Log("atd-ollama-generate", "Completed generation successfully")
}
