package chat

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/bark-cli/internal/ai"
	"github.com/cicbyte/bark-cli/internal/client"
	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	chatInteractive bool
	chatNonStream   bool
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [question]",
		Short: "AI 对话，通过自然语言发送推送",
		RunE:  runChat,
	}

	cmd.Flags().BoolVarP(&chatInteractive, "interactive", "i", false, "多轮对话模式")
	cmd.Flags().BoolVar(&chatNonStream, "non-stream", false, "非流式输出")
	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg := common.AppConfigModel
	if cfg.AI.ApiKey == "" {
		return fmt.Errorf("AI 未配置，请先运行: bark-cli config set ai.api_key <key>")
	}

	svc := ai.NewAIService(cfg.AI.Provider, cfg.AI.BaseURL, cfg.AI.ApiKey, cfg.AI.Model)

	ai.ExecutePush = newPushExecutor(cfg)
	ai.ExecuteDeviceRegister = newDeviceRegisterExecutor(cfg)
	ai.ExecuteDeviceCheck = newDeviceCheckExecutor(cfg)
	ai.ExecuteServerPing = newServerPingExecutor(cfg)
	ai.ExecuteServerInfo = newServerInfoExecutor(cfg)

	ctx := cmd.Context()

	if chatInteractive {
		return runInteractive(svc, ctx, cmd)
	}

	if len(args) == 0 {
		return fmt.Errorf("请输入问题，或使用 --interactive 进入多轮对话模式")
	}

	question := strings.Join(args, " ")
	_, err := askWithHistory(svc, ctx, question, nil, chatNonStream)
	return err
}

func askWithHistory(svc *ai.AIService, ctx context.Context, question string, history []ai.ChatMessage, nonStream bool) (string, error) {
	if nonStream {
		resp, err := svc.Ask(ctx, question, history, nil)
		if err != nil {
			return "", fmt.Errorf("AI 请求失败: %w", err)
		}
		fmt.Print(output.RenderMarkdown(resp.Answer))
		if resp.PromptTokens > 0 || resp.CompletionTokens > 0 {
			fmt.Printf("\n%s\nTokens: %d prompt + %d completion | Model: %s\n",
				output.Dim("---"), resp.PromptTokens, resp.CompletionTokens, resp.Model)
		}
		return resp.Answer, nil
	}

	start := time.Now()
	var buf strings.Builder
	var promptTokens, completionTokens int

	err := svc.AskStream(ctx, question, history, nil, func(event ai.StreamEvent) {
		switch event.Type {
		case "content":
			buf.WriteString(event.Content)
		case "tool_call":
			fmt.Printf("  %s %s\n", output.Dim("▸"), output.Cyan(event.Tool))
		case "tool_result":
			fmt.Printf("  %s %s\n", output.Dim("✓"), event.Content)
		case "done":
			promptTokens = event.PromptTokens
			completionTokens = event.CompletionTokens
		case "error":
			fmt.Printf("  %s\n", output.Dim("✗ "+event.Content))
		}
	})

	if err != nil {
		fmt.Println()
		return "", fmt.Errorf("AI 请求失败: %w", err)
	}

	raw := buf.String()
	if raw == "" {
		return "", nil
	}

	fmt.Println()
	fmt.Print(output.RenderMarkdown(raw))

	if promptTokens > 0 || completionTokens > 0 {
		elapsed := time.Since(start)
		fmt.Printf("\n%s\n", output.Dim(fmt.Sprintf("Tokens: %d + %d · %.1fs",
			promptTokens, completionTokens, elapsed.Seconds())))
	}

	return raw, nil
}

func runInteractive(svc *ai.AIService, ctx context.Context, cmd *cobra.Command) error {
	fmt.Println(output.Bold(" 多轮对话模式"))
	fmt.Println(output.Dim("────────────────────────────────────────────────"))
	fmt.Println(output.Dim("  输入问题开始对话，输入 /quit 退出，/clear 清除上下文"))

	var history []ai.ChatMessage
	reader := bufio.NewReader(cmd.InOrStdin())

	for {
		fmt.Print(output.Bold("  user > "))
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if input == "/quit" || input == "/exit" || input == "/q" {
			fmt.Println(output.Dim("  再见!"))
			break
		}
		if input == "/clear" {
			history = nil
			fmt.Println(output.Dim("  上下文已清除"))
			fmt.Println()
			continue
		}

		fmt.Println()
		resp, err := askWithHistory(svc, ctx, input, history, chatNonStream)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  错误: %v\n", err)
			fmt.Println()
			continue
		}

		if resp != "" {
			history = append(history, ai.ChatMessage{Role: "user", Content: input})
			history = append(history, ai.ChatMessage{Role: "assistant", Content: resp})
		}
		fmt.Println()
	}
	return nil
}

func newPushExecutor(cfg *models.AppConfig) func(ctx context.Context, params ai.PushParams) string {
	return func(ctx context.Context, params ai.PushParams) string {
		bark := newBarkClient(cfg)
		req := &client.PushRequest{
			Title: params.Title, Subtitle: params.Subtitle, Body: params.Body,
			Sound: params.Sound, Level: params.Level, Volume: params.Volume,
			Badge: params.Badge, Icon: params.Icon, Group: params.Group,
			URL: params.URL, Copy: params.Copy,
		}
		if params.Key != "" {
			req.DeviceKey = params.Key
		} else {
			req.DeviceKey = cfg.GetDefaultServer().DeviceKey
		}
		if params.Archive {
			req.IsArchive = "1"
		}
		if params.Call {
			req.Call = "1"
		}
		resp, err := bark.Push(req)
		if err != nil {
			return fmt.Sprintf("推送失败: %v", err)
		}
		if resp.Code != 200 {
			return fmt.Sprintf("推送失败: %s", resp.Message)
		}
		return "推送成功"
	}
}

func newDeviceRegisterExecutor(cfg *models.AppConfig) func(ctx context.Context, token, key string) string {
	return func(ctx context.Context, token, key string) string {
		bark := newBarkClient(cfg)
		resp, err := bark.Register(key, token)
		if err != nil {
			return fmt.Sprintf("注册失败: %v", err)
		}
		return fmt.Sprintf("设备注册成功, device_key: %s", resp.DeviceKey)
	}
}

func newDeviceCheckExecutor(cfg *models.AppConfig) func(ctx context.Context, key string) string {
	return func(ctx context.Context, key string) string {
		bark := newBarkClient(cfg)
		err := bark.CheckDevice(key)
		if err != nil {
			return fmt.Sprintf("设备未注册: %v", err)
		}
		return "设备已注册"
	}
}

func newServerPingExecutor(cfg *models.AppConfig) func(ctx context.Context) string {
	return func(ctx context.Context) string {
		bark := newBarkClient(cfg)
		err := bark.Ping()
		if err != nil {
			return fmt.Sprintf("服务器不可达: %v", err)
		}
		return "服务器连接正常"
	}
}

func newServerInfoExecutor(cfg *models.AppConfig) func(ctx context.Context) string {
	return func(ctx context.Context) string {
		bark := newBarkClient(cfg)
		info, err := bark.Info()
		if err != nil {
			return fmt.Sprintf("获取信息失败: %v", err)
		}
		return fmt.Sprintf("版本: %s, 架构: %s, 设备数: %d", info.Version, info.Arch, info.Devices)
	}
}

func newBarkClient(cfg *models.AppConfig) *client.Client {
	inst := cfg.GetDefaultServer()
	if inst == nil {
		return nil
	}
	timeout := inst.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	auth := &client.AuthConfig{}
	if inst.Token != "" {
		auth.Username = "bark-cli"
		auth.Password = inst.Token
	}
	return client.New(inst.URL, timeout, auth)
}

var _ = openai.Tool{}
