package ai

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBuildAdvisorContextJSONIncludesCredits(t *testing.T) {
	credits := &CreditsView{Available: 10, IncludedUnlocked: 15, IncludedUsed: 5}
	raw := buildAdvisorContextJSON(context.Background(), 1, credits, nil)
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["credits"] == nil {
		t.Fatal("expected credits in context")
	}
	if _, ok := payload["liabilities"]; ok {
		t.Fatal("nil deps should omit liabilities")
	}
}
