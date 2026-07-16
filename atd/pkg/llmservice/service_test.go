package llmservice

// WP-6 (test_atd_07_26.md §3.5, §6): pkg/llmservice previously had zero
// test files. These tests cover (a) the chat.Provider injection seam
// ((*Service).AddProvider — the seam pkg/testutil/fakeprovider.ChatProvider
// is designed for), (b) provider selection by model-name substring, and
// (c) the chat manifesto prompt golden (llmservice's half of the WP-6
// prompt-snapshot deliverable, alongside pkg/prompt's goldens).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/chat"
	"atd-tools/pkg/testutil"
	"atd-tools/pkg/testutil/fakeprovider"
)

// TestServiceChat_InjectedFakeProvider proves the injection seam: a fake
// provider added via AddProvider is used for Chat with no env vars, no
// config providers, and no network. Also asserts the service stamps the
// chat manifesto into req.System before dispatch — the prompt-assembly
// behavior the goldens below pin the content of.
func TestServiceChat_InjectedFakeProvider(t *testing.T) {
	testutil.SnapshotConfig(t)

	fake := &fakeprovider.ChatProvider{
		NameStr: "fake",
		ChatFunc: func(ctx context.Context, req chat.ChatRequest) (*chat.ChatResponse, error) {
			return &chat.ChatResponse{Message: "canned zzfix reply"}, nil
		},
	}

	svc := NewService(&config.Config{})
	svc.AddProvider(fake)

	resp, err := svc.Chat(context.Background(), chat.ChatRequest{
		Model:    "fake-model",
		Messages: []chat.Message{{Role: "user", Content: "hello"}},
	}, map[string]string{})
	if err != nil {
		t.Fatalf("Chat via injected fake: %v", err)
	}
	if resp.Message != "canned zzfix reply" {
		t.Errorf("expected the fake's canned reply, got: %q", resp.Message)
	}

	if fake.CallCount() != 1 {
		t.Fatalf("expected exactly 1 provider call, got %d", fake.CallCount())
	}
	seen := fake.Calls[0]
	if seen.System != GetChatManifesto() {
		t.Errorf("Chat must stamp the chat manifesto into req.System before dispatching to the provider")
	}
}

// TestServiceChat_SelectsProviderByModelSubstring pins the routing rule in
// (*Service).Chat: the provider whose Name() is a substring of the
// requested model wins.
func TestServiceChat_SelectsProviderByModelSubstring(t *testing.T) {
	testutil.SnapshotConfig(t)

	alpha := &fakeprovider.ChatProvider{NameStr: "alpha"}
	beta := &fakeprovider.ChatProvider{NameStr: "beta"}

	svc := NewService(&config.Config{})
	svc.AddProvider(alpha)
	svc.AddProvider(beta)

	_, err := svc.Chat(context.Background(), chat.ChatRequest{Model: "beta-large-v2"}, map[string]string{})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if beta.CallCount() != 1 || alpha.CallCount() != 0 {
		t.Errorf("expected model 'beta-large-v2' to route to provider 'beta' (alpha=%d beta=%d calls)", alpha.CallCount(), beta.CallCount())
	}
}

// TestServiceChat_ProviderErrorIsWrappedLoudly: a provider failure must
// surface as an error naming the provider — never a nil-response success.
func TestServiceChat_ProviderErrorIsWrappedLoudly(t *testing.T) {
	testutil.SnapshotConfig(t)

	boom := errors.New("zzfix provider exploded")
	fake := &fakeprovider.ChatProvider{
		NameStr: "fake",
		ChatFunc: func(ctx context.Context, req chat.ChatRequest) (*chat.ChatResponse, error) {
			return nil, boom
		},
	}

	svc := NewService(&config.Config{})
	svc.AddProvider(fake)

	_, err := svc.Chat(context.Background(), chat.ChatRequest{Model: "fake"}, map[string]string{})
	if err == nil {
		t.Fatal("expected the provider error to propagate, got nil")
	}
	if !errors.Is(err, boom) {
		t.Errorf("expected the wrapped provider error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "fake") {
		t.Errorf("error must name the failing provider, got: %v", err)
	}
}

// TestGoldenPrompt_ChatManifesto goldens the webui chat system prompt
// (llmservice/prompts.go) under testdata/golden/prompts/, completing the
// WP-6 prompt-snapshot set (pkg/prompt's builders are golden'd in
// pkg/prompt/golden_test.go).
func TestGoldenPrompt_ChatManifesto(t *testing.T) {
	sb := &testutil.SB{}
	sb.Golden(t, "prompts/chat_manifesto.txt", GetChatManifesto())
}
