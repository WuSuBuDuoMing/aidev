// NeoCode — AI coding agent optimized for Chinese and international models.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/agent"
	"github.com/WuSuBuDuoMing/aidev/internal/cli"
	"github.com/WuSuBuDuoMing/aidev/internal/config"
	"github.com/WuSuBuDuoMing/aidev/internal/config/model"
	"github.com/WuSuBuDuoMing/aidev/internal/permission"
	"github.com/WuSuBuDuoMing/aidev/internal/provider"
	"github.com/WuSuBuDuoMing/aidev/internal/session"
	"github.com/WuSuBuDuoMing/aidev/internal/skill"
	"github.com/WuSuBuDuoMing/aidev/internal/storage"
	"github.com/WuSuBuDuoMing/aidev/internal/tool/builtin"
	"github.com/WuSuBuDuoMing/aidev/internal/tool/mcp"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

const Version = "2.6.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("neocode %s\n", Version)
			return
		case "status":
			showStatus()
			return
		case "providers":
			showProviders()
			return
		case "upgrade":
			if err := cli.Upgrade(Version); err != nil {
				fmt.Fprintf(os.Stderr, "Upgrade error: %v\n", err)
				os.Exit(1)
			}
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	// Default: interactive chat
	runChat()
}

func printUsage() {
	fmt.Println(`NeoCode — AI coding agent for Chinese and international models

Usage:
  neocode                    Start interactive chat
  neocode "question"         Ask a single question
  neocode status             Show configuration and status
  neocode providers          List available providers
  neocode version            Show version

Options:
  -p, --provider <name>      Set provider (deepseek, claude, openai, gemini, mimo, ...)
  -m, --model <id>           Set model ID
  -k, --api-key <key>        Set API key
  --temperature <float>      Set temperature (0-2)
  --effort <level>           Set effort (low, medium, high, xhigh, ultracode)
  --mode <mode>              Set permission mode (ask, auto, plan, edit)
  --no-stream                Disable streaming output
  --theme <name>             Set theme (dark, light, monokai, dracula)
  --thinking                 Enable thinking/reasoning mode

Environment Variables:
  NEOCODE_PROVIDER, NEOCODE_MODEL, NEOCODE_API_KEY, NEOCODE_BASE_URL
  ANTHROPIC_API_KEY, OPENAI_API_KEY, DEEPSEEK_API_KEY, MI_API_KEY
  QWEN_API_KEY, GLM_API_KEY, MOONSHOT_API_KEY, GOOGLE_API_KEY

ccSwitch Compatible:
  NeoCode reads all standard ccSwitch environment variables.
  Run 'neocode providers import --from-ccswitch' to import ccSwitch config.`)
}

func showStatus() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	m := model.GetSpec(cfg.DefaultModel)
	fmt.Printf(`
  NeoCode Status
  ─────────────────────────────────────
  Provider:     %s
  Model:        %s (context: %dK)
  API Key:      %s
  Base URL:     %s
  Temperature:  %.2f
  Effort:       %s
  Permission:   %s
  Stream:       %v
  Theme:        %s
  Language:     %s
  Config Dir:   %s
  Version:      %s
`,
		cfg.DefaultProvider,
		cfg.DefaultModel, m.ContextWindow/1000,
		maskKey(cfg.APIKey),
		cfg.BaseURL,
		cfg.Temperature,
		cfg.Effort,
		cfg.PermissionMode,
		cfg.Stream,
		cfg.Theme,
		cfg.Language,
		config.ConfigDir(),
		Version,
	)
}

func showProviders() {
	providers := []struct {
		Name    string
		URL     string
		Default string
	}{
		{"deepseek", "https://api.deepseek.com", "deepseek-v4"},
		{"claude", "https://api.anthropic.com", "claude-sonnet-4-6"},
		{"openai", "https://api.openai.com", "gpt-4o"},
		{"gemini", "https://generativelanguage.googleapis.com", "gemini-2.5-flash"},
		{"mimo", "https://api.xiaomimimo.com/v1", "mimo-v2.5-pro"},
		{"glm", "https://open.bigmodel.cn/api/paas/v4", "glm-5.1"},
		{"moonshot", "https://api.moonshot.cn/v1", "kimi-k2.7-code"},
		{"stepfun", "https://api.stepfun.com/v1", "step-3.5-flash-2603"},
		{"minimax", "https://api.minimax.chat/v1", "minimax-m2.7"},
		{"baidu", "https://qianfan.baidubce.com/v2", "qianfan-code-latest"},
		{"volcengine", "https://ark.cn-beijing.volces.com/api/v3", "ark-code-latest"},
		{"bailing", "https://api.tbox.cn/v1", "ling-2.5-1t"},
		{"ollama", "http://localhost:11434", "llama3.1"},
	}

	fmt.Println("\n  Available Providers")
	fmt.Println("  ──────────────────────────────────────────")
	for _, p := range providers {
		fmt.Printf("  %-14s %-50s %s\n", p.Name, p.URL, p.Default)
	}
	fmt.Println()
}

func runChat() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI overrides
	for i := 1; i < len(os.Args)-1; i++ {
		switch os.Args[i] {
		case "-p", "--provider":
			cfg.DefaultProvider = os.Args[i+1]
			i++
		case "-m", "--model":
			cfg.DefaultModel = os.Args[i+1]
			i++
		case "-k", "--api-key":
			cfg.APIKey = os.Args[i+1]
			i++
		case "--effort":
			cfg.Effort = os.Args[i+1]
			i++
		case "--mode":
			cfg.PermissionMode = os.Args[i+1]
			i++
		}
	}

	// Check API key
	if cfg.APIKey == "" && cfg.DefaultProvider != "ollama" {
		fmt.Fprintf(os.Stderr, "\n  No API key configured.\n")
		fmt.Fprintf(os.Stderr, "  Set NEOCODE_API_KEY or configure via neocode.toml\n\n")
		os.Exit(1)
	}

	// Create provider and tool registry
	prov, modelName, err := provider.CreateProvider(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Provider error: %v\n", err)
		os.Exit(1)
	}
	tools := builtin.NewDefaultRegistry()
	mcpManager := mcp.NewManager()
	defer mcpManager.CloseAll()

	// Initialize session manager
	db, err := storage.Open(config.DBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	sessMgr := session.NewManager(db)

	// Load skills
	skills := skill.Discover(config.SkillsDir(), filepath.Join(".", ".neocode", "skills"))

	// Create conversation
	conv := &types.Conversation{
		ID:        fmt.Sprintf("conv-%d", time.Now().UnixMilli()),
		Title:     "New Conversation",
		Messages:  []types.Message{},
		CreatedAt: time.Now(),
		Model:     modelName,
		Provider:  cfg.DefaultProvider,
	}

	// Context cancellation on SIGINT
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Auto-discover and connect MCP servers from .mcp.json
	cwd, _ := os.Getwd()
	mcpConfigPath := filepath.Join(cwd, ".mcp.json")
	if _, err := os.Stat(mcpConfigPath); err == nil {
		if err := mcpManager.ConnectFromConfig(ctx, mcpConfigPath, tools); err != nil {
			fmt.Printf("  ⚠ MCP: %v\n", err)
		}
	}
	mcpConfigPath2 := filepath.Join(cwd, ".neocode", "mcp.json")
	if _, err := os.Stat(mcpConfigPath2); err == nil {
		if err := mcpManager.ConnectFromConfig(ctx, mcpConfigPath2, tools); err != nil {
			fmt.Printf("  ⚠ MCP: %v\n", err)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nGoodbye! Happy coding.")
		cancel()
		os.Exit(0)
	}()

	// Single question mode
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		question := os.Args[1]
		conv.Messages = []types.Message{
			{Role: types.RoleUser, Content: question, Timestamp: time.Now()},
		}

		result, _, err := agent.Run(ctx, agent.Options{
			Provider:       prov,
			Tools:          tools,
			SystemPrompt:   buildSystemPrompt(),
			ModelID:        modelName,
			PermissionMode: types.PermissionModeName(cfg.PermissionMode),
			Stream:         cfg.Stream,
			OnToken: func(chunk types.StreamChunk) {
				if chunk.Content != "" {
					fmt.Print(chunk.Content)
				}
			},
			OnToolStart: func(name string, args map[string]interface{}) {
				fmt.Printf("\n  🔧 %s\n", permission.FormatToolCall(name, args))
			},
			OnToolEnd: func(name string, tr *types.ToolResult) {
				if tr.Success {
					preview := tr.Output
					if len(preview) > 200 {
						preview = preview[:200] + "..."
					}
					fmt.Printf("  ✓ %s: %s\n", name, strings.ReplaceAll(preview, "\n", " "))
				} else {
					fmt.Printf("  ✗ %s: %s\n", name, tr.Error)
				}
			},
		}, conv.Messages)

		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
		printUsageStats(result, modelName, cfg.ShowTokenCounts)
		os.Exit(0)
	}

	// Interactive REPL
	printBanner()
	fmt.Printf("  Model: %s (%s)\n", modelName, prov.Name())
	fmt.Printf("  Effort: %s | Mode: %s\n", cfg.Effort, cfg.PermissionMode)
	fmt.Printf("  Tools: %d registered\n", len(tools.GetAll()))
	fmt.Printf("  Skills: %d loaded\n", len(skills))
	fmt.Println("  Type /help for commands, or just start chatting.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("neocode> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Slash commands
		if strings.HasPrefix(input, "/") {
			if handleCommand(input, cfg, sessMgr, skills, conv) {
				continue
			}
			break
		}

		conv.Messages = append(conv.Messages, types.Message{
			Role:      types.RoleUser,
			Content:   input,
			Timestamp: time.Now(),
		})

		// Auto-title from first user message
		if conv.Title == "New Conversation" {
			conv.Title = session.GenerateTitle(input)
		}

		// Prune old tool results before sending
		conv.Messages = agent.PruneToolResults(conv.Messages)

		// Auto-compact if needed
		spec := model.GetSpec(modelName)
		conv.Messages, _ = agent.CompactIfNeeded(conv.Messages, model.CompactThreshold(spec.ContextWindow))

		result, finalMsgs, err := agent.Run(ctx, agent.Options{
			Provider:       prov,
			Tools:          tools,
			SystemPrompt:   buildSystemPrompt(),
			ModelID:        modelName,
			PermissionMode: types.PermissionModeName(cfg.PermissionMode),
			Stream:         cfg.Stream,
			OnToken: func(chunk types.StreamChunk) {
				if chunk.Content != "" {
					fmt.Print(chunk.Content)
				}
			},
			OnToolStart: func(name string, args map[string]interface{}) {
				fmt.Printf("\n  🔧 %s\n", permission.FormatToolCall(name, args))
			},
			OnToolEnd: func(name string, tr *types.ToolResult) {
				if tr.Success {
					preview := tr.Output
					if len(preview) > 200 {
						preview = preview[:200] + "..."
					}
					fmt.Printf("  ✓ %s: %s\n", name, strings.ReplaceAll(preview, "\n", " "))
				} else {
					fmt.Printf("  ✗ %s: %s\n", name, tr.Error)
				}
			},
		}, conv.Messages)

		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			continue
		}

		conv.Messages = finalMsgs

		// Auto-save session
		if err := sessMgr.Save(conv); err != nil {
			fmt.Fprintf(os.Stderr, "  [save warning: %v]\n", err)
		}

		fmt.Println()
		printUsageStats(result, modelName, cfg.ShowTokenCounts)
		fmt.Println()
	}
}

func buildSystemPrompt() string {
	cwd, _ := os.Getwd()
	return fmt.Sprintf(`You are NeoCode — an expert AI coding agent running in the terminal.

## Core Principles
- Be concise, accurate, and direct.
- Prefer showing code over describing it.
- When modifying files, use the provided tools rather than printing code blocks.
- Always explain what you are about to do before using a tool.
- Respect the user's existing code style and architecture.
- If unsure about intent, ask for clarification.

## Working Directory
%s

## Output Style
- Use markdown for structured responses.
- Use fenced code blocks with language tags.
- Keep explanations brief — the user is a developer.`, cwd)
}

func handleCommand(input string, cfg *config.Config, sessMgr *session.Manager, skills []skill.Skill, conv *types.Conversation) bool {
	cmd := strings.Fields(input)[0]
	switch cmd {
	case "/help":
		fmt.Println(`Commands:
  /help              Show this help
  /status            Show current configuration
  /providers         List available providers
  /model <id>        Switch model
  /provider <id>     Switch provider
  /effort <lvl>      Set effort (low/medium/high/xhigh/ultracode)
  /mode <mode>       Set mode (ask/auto/plan/edit)
  /plan <task>       Analyze task in read-only plan mode
  /agents            List available sub-agents
  /history [n]       List recent sessions
  /history search X  Search sessions
  /resume <id>       Resume a past session
  /skills            List available skills
  /new               Start a new conversation
  /upgrade           Check for updates
  /clear             Clear screen
  /exit              Exit`)
	case "/status":
		showStatus()
	case "/providers":
		showProviders()
	case "/model":
		parts := strings.Fields(input)
		if len(parts) > 1 {
			cfg.DefaultModel = parts[1]
			fmt.Printf("  Model switched to: %s\n", parts[1])
		}
	case "/provider":
		parts := strings.Fields(input)
		if len(parts) > 1 {
			cfg.DefaultProvider = parts[1]
			fmt.Printf("  Provider switched to: %s\n", parts[1])
		}
	case "/effort":
		parts := strings.Fields(input)
		if len(parts) > 1 {
			cfg.Effort = parts[1]
			fmt.Printf("  Effort set to: %s\n", parts[1])
		}
	case "/mode":
		parts := strings.Fields(input)
		if len(parts) > 1 {
			cfg.PermissionMode = parts[1]
			fmt.Printf("  Mode set to: %s\n", parts[1])
		}
	case "/history":
		parts := strings.Fields(input)
		if len(parts) > 1 && parts[1] == "search" && len(parts) > 2 {
			query := strings.Join(parts[2:], " ")
			sessions, err := sessMgr.Search(query, 10)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				return true
			}
			if len(sessions) == 0 {
				fmt.Println("  No matching sessions found.")
				return true
			}
			fmt.Println("  Search Results:")
			for _, s := range sessions {
				fmt.Printf("  %s  %-40s  %d msgs  %s\n", s.ID[:8], s.Title, s.MessageCount, s.UpdatedAt)
			}
		} else {
			limit := 10
			if len(parts) > 1 {
				fmt.Sscanf(parts[1], "%d", &limit)
			}
			sessions, err := sessMgr.List(limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				return true
			}
			if len(sessions) == 0 {
				fmt.Println("  No saved sessions.")
				return true
			}
			fmt.Println("  Recent Sessions:")
			for _, s := range sessions {
				fmt.Printf("  %s  %-40s  %d msgs  %s\n", s.ID[:8], s.Title, s.MessageCount, s.UpdatedAt)
			}
		}
	case "/resume":
		parts := strings.Fields(input)
		if len(parts) < 2 {
			fmt.Println("  Usage: /resume <session-id>")
			return true
		}
		loaded, err := sessMgr.Load(parts[1])
		if err != nil {
			// Try partial match
			sessions, _ := sessMgr.List(20)
			for _, s := range sessions {
				if strings.HasPrefix(s.ID, parts[1]) {
					loaded, err = sessMgr.Load(s.ID)
					break
				}
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				return true
			}
		}
		conv.ID = loaded.ID
		conv.Title = loaded.Title
		conv.Messages = loaded.Messages
		conv.Provider = loaded.Provider
		conv.Model = loaded.Model
		fmt.Printf("  Resumed: %s (%d messages)\n", loaded.Title, len(loaded.Messages))
	case "/skills":
		fmt.Println(skill.FormatList(skills))
	case "/agents":
		fmt.Println("  Available Sub-Agents:")
		fmt.Println("  ─────────────────────────────────────────")
		for _, a := range agent.ListSubAgents() {
			fmt.Printf("  %-20s %s\n", a.Name, a.Prompt[:min(60, len(a.Prompt))]+"...")
		}
	case "/plan":
		task := strings.TrimPrefix(input, "/plan ")
		if task == "/plan" || task == "" {
			fmt.Println("  Usage: /plan <task description>")
			return true
		}
		fmt.Printf("  📋 Analyzing in plan mode...\n")
		// Note: prov is not accessible here from handleCommand.
		// Plan mode is triggered via mode switching instead.
		cfg.PermissionMode = "plan"
		fmt.Printf("  Switched to plan mode. Your next message will be analyzed read-only.\n")
	case "/new":
		conv.ID = fmt.Sprintf("conv-%d", time.Now().UnixMilli())
		conv.Title = "New Conversation"
		conv.Messages = nil
		fmt.Println("  Started new conversation.")
	case "/upgrade":
		if err := cli.Upgrade(Version); err != nil {
			fmt.Fprintf(os.Stderr, "  Upgrade error: %v\n", err)
		}
	case "/clear":
		fmt.Print("\033c")
	case "/exit":
		fmt.Println("Goodbye! Happy coding.")
		os.Exit(0)
	default:
		return false
	}
	return true
}

func printUsageStats(result *types.GenerateResult, modelID string, show bool) {
	if !show || result.Usage.TotalTokens == 0 {
		return
	}
	spec := model.GetSpec(modelID)
	cost := float64(result.Usage.PromptTokens)*spec.PriceInput/1_000_000 +
		float64(result.Usage.CompletionTokens)*spec.PriceOutput/1_000_000

	fmt.Printf("  Tokens: %d in / %d out / %d total",
		result.Usage.PromptTokens, result.Usage.CompletionTokens, result.Usage.TotalTokens)
	if cost > 0 {
		fmt.Printf(" | Cost: $%.4f", cost)
	}
	fmt.Println()
}

func maskKey(key string) string {
	if key == "" {
		return "not set"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func printBanner() {
	fmt.Println(`
  ███╗   ██╗███████╗ ██████╗  ██████╗ ██████╗ ██████╗ ███████╗
  ████╗  ██║██╔════╝██╔═══██╗██╔════╝██╔═══██╗██╔══██╗██╔════╝
  ██╔██╗ ██║█████╗  ██║   ██║██║     ██║   ██║██║  ██║█████╗
  ██║╚██╗██║██╔══╝  ██║   ██║██║     ██║   ██║██║  ██║██╔══╝
  ██║ ╚████║███████╗╚██████╔╝╚██████╗╚██████╔╝██████╔╝███████╗
  ╚═╝  ╚═══╝╚══════╝ ╚═════╝  ╚═════╝ ╚═════╝ ╚═════╝╚══════╝`)
	fmt.Println("  AI coding agent — DeepSeek · Claude · GPT · Gemini · MiMo · Qwen · GLM")
	fmt.Println()
}
