package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDshAnthropicUsesApiKeyAdapter(t *testing.T) {
	home := t.TempDir()
	argv := []string{"dsh", "--profile", "headless", "--json", "hi"}
	_, env, err := ApplyAgentCLICredential(AgentCLIDeepSeek,
		ProviderCredential{APIKey: "sk-mem-test", BaseURL: "http://172.32.153.183:8096/dsh/default/v1", Protocol: AgentCLIProtocolMessages},
		AgentCLIOptions{Model: "GLM-5.2"}, home, home, argv)
	if err != nil {
		t.Fatal(err)
	}
	if env[agentCLIKeyEnv] != "sk-mem-test" {
		t.Fatalf("key env missing: %v", env)
	}
	data, err := os.ReadFile(filepath.Join(home, ".dsh", "agentflow-model.json"))
	if err != nil {
		t.Fatal(err)
	}
	var patch []map[string]any
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatal(err)
	}
	if patch[0]["name"] != "@deepseek-ai/dsh-llm-deepseek-api-key" {
		t.Fatalf("expected api-key adapter, got %v", patch[0])
	}
	cfg := patch[0]["config"].(map[string]any)
	if cfg["baseURL"] != "http://172.32.153.183:8096/dsh/default" {
		t.Fatalf("baseURL /v1 not stripped: %v", cfg["baseURL"])
	}
	t.Logf("patch: %s", data)
}

func TestDshTDBMIdentityEnv(t *testing.T) {
	home := t.TempDir()
	argv := []string{"dsh", "--profile", "headless", "--json", "hi"}
	_, env, err := ApplyAgentCLICredential(AgentCLIDeepSeek,
		ProviderCredential{APIKey: "sk-mem-test", BaseURL: "http://p:8096/dsh/default", Protocol: AgentCLIProtocolMessages,
			TDBMTeam: "LoopForge", TDBMTask: "release-check"},
		AgentCLIOptions{Model: "m", TDBMAgent: "reviewer"}, home, home, argv)
	if err != nil {
		t.Fatal(err)
	}
	if env["TDBM_AGENT_ID"] != "reviewer" || env["TDBM_TEAM_ID"] != "LoopForge" || env["TDBM_TASK_ID"] != "release-check" {
		t.Fatalf("tdbm identity env wrong: %v", env)
	}

	// Credential values win over platform env; without credential values the
	// platform env applies.
	t.Setenv("AGENTFLOW_TDBM_TEAM", "PlatformTeam")
	_, env, _ = ApplyAgentCLICredential(AgentCLIDeepSeek,
		ProviderCredential{APIKey: "sk", BaseURL: "http://p:8096/dsh/default", Protocol: AgentCLIProtocolMessages, TDBMTeam: "UserTeam"},
		AgentCLIOptions{Model: "m"}, home, home, argv)
	if env["TDBM_TEAM_ID"] != "UserTeam" {
		t.Fatalf("user team should win, got %q", env["TDBM_TEAM_ID"])
	}
	_, env, _ = ApplyAgentCLICredential(AgentCLIDeepSeek,
		ProviderCredential{APIKey: "sk", BaseURL: "http://p:8096/dsh/default", Protocol: AgentCLIProtocolMessages},
		AgentCLIOptions{Model: "m"}, home, home, argv)
	if env["TDBM_TEAM_ID"] != "PlatformTeam" {
		t.Fatalf("platform fallback should apply, got %q", env["TDBM_TEAM_ID"])
	}
}
