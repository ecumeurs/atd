# Step 1: Backend Gemini API Proxy

## Objective
Add a `/api/gemini/chat` POST endpoint to the Go backend that proxies chat messages to the Gemini API using the official `google.golang.org/genai` SDK and returns structured JSON responses. Support user-selectable models via a `/api/gemini/models` listing endpoint.

## Prerequisites
- `google.golang.org/genai` is already in `go.mod`
- `.env` file exists at `webui/.env` with `GEMINI_API_KEY=<key>`
- Default model: `gemini-2.5-flash` (user-selectable from UI)

## Files to Modify

### `webui/main.go`

**Current state**: 270-line Go/Gin server with routes under `/api/` group. Uses `gin-gonic/gin`. Config loaded from `config.json`. Atoms parsed from disk.

### Changes Required

#### 1. Add imports

Add these imports to the existing import block:

```go
import (
    // ... existing imports ...
    "context"
    "github.com/joho/godotenv"
    "google.golang.org/genai"
)
```

> **Note**: You will also need to `go get github.com/joho/godotenv` for .env loading.

#### 2. Add .env loading in `main()`

At the very start of `main()`, before `loadConfig()`:

```go
func main() {
    // Load .env file for GEMINI_API_KEY
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not found, relying on environment variables")
    }
    
    loadConfig()
    refreshAtoms()
    // ... rest unchanged
}
```

#### 3. Define request/response structs

Add these structs after the existing `Config` struct:

```go
// --- Gemini Chat Types ---

type ChatMessage struct {
    Role    string `json:"role"`    // "user" or "model"
    Content string `json:"content"`
}

type ChatRequest struct {
    Messages   []ChatMessage            `json:"messages"`    // Full conversation history
    Model      string                   `json:"model"`       // Selected Gemini model (default: gemini-2.5-flash)
    AtdContext []map[string]interface{} `json:"atd_context"` // ATD atoms to inject as context
    Actions    []ActionRecord           `json:"actions"`     // Accept/reject history
}

type ActionRecord struct {
    ProposalID string `json:"proposal_id"`
    AtomID     string `json:"atom_id"`
    Action     string `json:"action"` // "ACCEPTED" or "REJECTED"
    Summary    string `json:"summary"`
}

type GeminiProposal struct {
    Action        string                 `json:"action"`         // CREATE, UPDATE, DELETE
    AtomID        string                 `json:"atom_id"`
    Content       map[string]interface{} `json:"content"`
    ImpactSummary string                 `json:"impact_summary"`
}

type GeminiResponse struct {
    Message   string           `json:"message"`
    Proposals []GeminiProposal `json:"proposals"`
}
```

#### 4. Define the ATD Manifesto system prompt

Add this as a package-level constant:

```go
const atdManifesto = `You are an ATD (Atomic Traceable Documentation) Specification Architect.

RULES YOU MUST FOLLOW:
0. You are a sounding board for the user. You are not here to replace the user's judgement, but to help them make better decisions. You may challenge the user's assumptions and propose alternative solutions. You are not expected to provide new/update ATD at every message. You may ask for clarifications.
1. Every atom has EXACTLY ONE state-changing rule. If an intent needs "and" or "also", split into multiple atoms.
2. Atoms have strict YAML frontmatter: id, human_name, type, layer, version, status, priority, tags, parents, dependents.
3. The hierarchy is: CUSTOMER (requirements, usecases) -> ARCHITECTURE (modules, APIs) -> IMPLEMENTATION (mechanics, builds).
4. Valid types: MODULE, SERVICE, ENTITY, RULE, MECHANIC, DOMAIN, API, UI, DATA, USAGE, BUILD, REQUIREMENT, SPECIFICATION, USECASE, USER_STORY.
5. Valid statuses: DRAFT, REVIEW, STABLE.
6. Valid layers: CUSTOMER, ARCHITECTURE, IMPLEMENTATION.
7. Each atom has 4 mandatory H2 sections: INTENT, THE RULE / LOGIC, TECHNICAL INTERFACE, EXPECTATION.
8. The INTENT must be ONE sentence, no "and" or "also".
9. Parents link upward (impl -> arch -> customer). Dependents link downward.

RESPONSE FORMAT:
You MUST respond with valid JSON matching this schema:
{
  "message": "Your conversational response explaining your reasoning",
  "proposals": [
    {
      "action": "CREATE" | "UPDATE" | "DELETE",
      "atom_id": "type_snake_case_name",
      "content": {
        "human_name": "Human Readable Name",
        "type": "MECHANIC",
        "layer": "IMPLEMENTATION",
        "status": "DRAFT",
        "priority": "3",
        "tags": ["tag1"],
        "parents": ["parent_atom_id"],
        "intent": "Single sentence why this exists.",
        "logic": "The core specification.",
        "technical_interface": "API endpoints, code tags, test names.",
        "expectation": "Verifiable acceptance criteria."
      },
      "impact_summary": "Brief description of what this change means."
    }
  ]
}

When the conversation is exploratory or you need clarification, return "proposals": []. Only propose atoms when you have sufficient information and the user's intent is clear.
When proposing updates, only include fields that change in "content".
When no proposals are needed (e.g. answering a question, challenging an assumption, or asking for more details), return an empty proposals array.
Always explain your reasoning in "message" before listing proposals.`

const geminiResponseSchema = `{
  "type": "object",
  "properties": {
    "message": { "type": "string" },
    "proposals": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "action": { "type": "string", "enum": ["CREATE", "UPDATE", "DELETE"] },
          "atom_id": { "type": "string" },
          "content": {
            "type": "object",
            "properties": {
              "human_name": { "type": "string" },
              "type": { "type": "string" },
              "layer": { "type": "string" },
              "status": { "type": "string" },
              "priority": { "type": "string" },
              "tags": { "type": "array", "items": { "type": "string" } },
              "parents": { "type": "array", "items": { "type": "string" } },
              "intent": { "type": "string" },
              "logic": { "type": "string" },
              "technical_interface": { "type": "string" },
              "expectation": { "type": "string" }
            }
          },
          "impact_summary": { "type": "string" }
        },
        "required": ["action", "atom_id"]
      }
    }
  },
  "required": ["message"]
}`
```

#### 5. Add the `/api/gemini/chat` endpoint

Inside the `api := r.Group("/api")` block, add:

```go
api.POST("/gemini/chat", handleGeminiChat)
```

#### 6. Implement `handleGeminiChat`

Add this function:

```go
func handleGeminiChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }

    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not configured"})
        return
    }

    // Set the env var the SDK expects
    os.Setenv("GOOGLE_API_KEY", apiKey)

    ctx := context.Background()
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gemini client: " + err.Error()})
        return
    }

    // Build the system instruction with manifesto
    systemInstruction := atdManifesto

    // Append ATD context if provided
    if len(req.AtdContext) > 0 {
        systemInstruction += "\n\n--- CURRENT ATD CONTEXT ---\nThe following atoms are currently relevant to this conversation:\n"
        for _, atd := range req.AtdContext {
            atdJSON, _ := json.Marshal(atd)
            systemInstruction += string(atdJSON) + "\n"
        }
    }

    // Append action history if provided
    if len(req.Actions) > 0 {
        systemInstruction += "\n\n--- USER ACTION HISTORY ---\nThe user has taken the following actions on previous proposals:\n"
        for _, action := range req.Actions {
            systemInstruction += fmt.Sprintf("- Proposal for atom '%s': %s. %s\n", action.AtomID, action.Action, action.Summary)
        }
    }

    // Build conversation parts from history
    var contents []*genai.Content
    for _, msg := range req.Messages {
        role := msg.Role
        if role == "assistant" {
            role = "model"
        }
        contents = append(contents, &genai.Content{
            Role: role,
            Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
        })
    }

    // Call Gemini API
    config := &genai.GenerateContentConfig{
        SystemInstruction: &genai.Content{
            Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
        },
        ResponseMIMEType: "application/json",
        ResponseSchema: &genai.Schema{
            Type: genai.TypeObject,
            Properties: map[string]*genai.Schema{
                "message": {Type: genai.TypeString},
                "proposals": {
                    Type: genai.TypeArray,
                    Items: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "action":         {Type: genai.TypeString, Enum: []string{"CREATE", "UPDATE", "DELETE"}},
                            "atom_id":        {Type: genai.TypeString},
                            "content":        {Type: genai.TypeObject},
                            "impact_summary": {Type: genai.TypeString},
                        },
                        Required: []string{"action", "atom_id"},
                    },
                },
            },
            Required: []string{"message"},
        },
        Temperature: genai.Ptr(float32(0.7)),
    }

    // Use selected model or default
    model := req.Model
    if model == "" {
        model = "gemini-2.5-flash"
    }

    result, err := client.Models.GenerateContent(ctx, model, contents, config)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini API error: " + err.Error()})
        return
    }

    // Extract the text response
    responseText := result.Text()

    // Parse the structured JSON response
    var geminiResp GeminiResponse
    if err := json.Unmarshal([]byte(responseText), &geminiResp); err != nil {
        // If parsing fails, return raw text as message
        geminiResp = GeminiResponse{
            Message:   responseText,
            Proposals: []GeminiProposal{},
        }
    }

    c.JSON(http.StatusOK, geminiResp)
}
```

#### 7. Add a model listing endpoint

Add inside the API group:

```go
api.GET("/gemini/models", func(c *gin.Context) {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not configured"})
        return
    }
    os.Setenv("GOOGLE_API_KEY", apiKey)

    ctx := context.Background()
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client: " + err.Error()})
        return
    }

    var models []map[string]interface{}
    iter := client.Models.List(ctx, nil)
    for {
        model, err := iter.Next()
        if err != nil {
            break
        }
        models = append(models, map[string]interface{}{
            "id":           model.Name,
            "display_name": model.DisplayName,
            "description":  model.Description,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "models":  models,
        "default": "gemini-2.5-flash",
    })
})
```

#### 8. Add an endpoint to list atoms for context selection

Add inside the API group:

```go
api.GET("/gemini/atoms", func(c *gin.Context) {
    query := strings.ToLower(c.Query("q"))
    var results []map[string]interface{}
    for _, atom := range Atoms {
        if query == "" || strings.Contains(strings.ToLower(atom.ID), query) || strings.Contains(strings.ToLower(atom.HumanName), query) {
            results = append(results, map[string]interface{}{
                "id":         atom.ID,
                "human_name": atom.HumanName,
                "type":       atom.Type,
                "layer":      atom.Layer,
                "status":     atom.Status,
                "intent":     extractIntent(atom.Content),
            })
        }
    }
    c.JSON(http.StatusOK, results)
})
```

#### 8. Add helper to extract intent from atom content

```go
func extractIntent(content string) string {
    lines := strings.Split(content, "\n")
    inIntent := false
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "## INTENT") {
            inIntent = true
            continue
        }
        if inIntent && strings.HasPrefix(trimmed, "## ") {
            break
        }
        if inIntent && trimmed != "" {
            return trimmed
        }
    }
    return ""
}
```

## Dependencies to install

```bash
cd webui
go get github.com/joho/godotenv
```

## Expected Outcome

After this step:
1. `POST /api/gemini/chat` accepts `{ messages, model, atd_context, actions }` and returns `{ message, proposals[] }`
2. `GET /api/gemini/models` returns available Gemini models (via `client.Models.List()`)
3. `GET /api/gemini/atoms?q=search` returns filtered atom list for context selection
4. The Gemini API key is loaded from `.env` securely
5. The ATD Manifesto is sent as system instruction on every request
6. The response is constrained to valid JSON via `ResponseMIMEType` and `ResponseSchema`
7. Default model is `gemini-2.5-flash`, overridable by `model` field in request

## Verification

```bash
# Build and run
cd webui && go build -o webui_bin . && ./webui_bin

# Test the endpoint (requires valid GEMINI_API_KEY in .env)
curl -X POST http://localhost:8081/api/gemini/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"I want to create a spec for a login system"}]}'

# Expected: JSON with "message" and "proposals" fields
```
