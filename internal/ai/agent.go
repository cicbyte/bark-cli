package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

const maxIterations = 5

type Agent struct {
	client *openai.Client
	model  string
}

func NewAgent(client *openai.Client, model string) *Agent {
	return &Agent{client: client, model: model}
}

func (a *Agent) buildMessages(question string, history []ChatMessage) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: a.buildSystemPrompt()},
	}
	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{Role: msg.Role, Content: msg.Content})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: "user", Content: question})
	return messages
}

func (a *Agent) buildSystemPrompt() string {
	return `你是一个推送通知助手。你可以帮助用户发送推送通知、管理设备。

可用工具：
- push: 发送推送通知（参数: title, subtitle, body, sound, level, volume, badge, icon, group, url, copy, key）
- device_register: 注册设备（参数: device_token, key）
- device_check: 检查设备是否已注册（参数: device_key）
- server_ping: 检查服务器连通性
- server_info: 获取服务器信息

注意事项：
- 回答使用中文
- 使用 Markdown 格式
- 如果用户没有指定设备，使用默认设备
- 仅在用户明确要求时执行写操作
- 当前时间：` + fmt.Sprintf("%s", currentTime())
}

func (a *Agent) getTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "push",
				Description: "发送推送通知到 iOS 设备",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":    map[string]any{"type": "string", "description": "通知标题"},
						"subtitle": map[string]any{"type": "string", "description": "通知副标题"},
						"body":     map[string]any{"type": "string", "description": "通知内容"},
						"sound":    map[string]any{"type": "string", "description": "提示音"},
						"level":    map[string]any{"type": "string", "description": "通知级别", "enum": []string{"critical", "active", "timeSensitive", "passive"}},
						"volume":   map[string]any{"type": "integer", "description": "音量(0-10)"},
						"badge":    map[string]any{"type": "integer", "description": "角标数字"},
						"icon":     map[string]any{"type": "string", "description": "图标 URL"},
						"group":    map[string]any{"type": "string", "description": "通知分组"},
						"url":      map[string]any{"type": "string", "description": "点击跳转 URL"},
						"copy":     map[string]any{"type": "string", "description": "复制内容"},
						"key":      map[string]any{"type": "string", "description": "设备 key（可选，不传则使用默认）"},
						"archive":  map[string]any{"type": "boolean", "description": "是否保存通知"},
							"call":     map[string]any{"type": "boolean", "description": "持续响铃 30 秒"},
					},
					"required": []string{"body"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "device_register",
				Description: "注册新的推送设备",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"device_token": map[string]any{"type": "string", "description": "Apple 设备推送 token"},
						"key":          map[string]any{"type": "string", "description": "指定设备 key（可选，不传则自动生成）"},
					},
					"required": []string{"device_token"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "device_check",
				Description: "检查设备是否已注册",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"device_key": map[string]any{"type": "string", "description": "设备 key"},
					},
					"required": []string{"device_key"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "server_ping",
				Description: "检查 bark-server 是否可连通",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "server_info",
				Description: "获取 bark-server 服务器信息",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}
}

func (a *Agent) Ask(ctx context.Context, question string, history []ChatMessage, tools []openai.Tool) (*AskResponse, error) {
	if len(tools) == 0 {
		tools = a.getTools()
	}
	messages := a.buildMessages(question, history)

	for range maxIterations {
		resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, fmt.Errorf("AI 请求失败: %w", err)
		}

		choice := resp.Choices[0]
		if len(choice.Message.ToolCalls) == 0 {
			return &AskResponse{
				Answer:          choice.Message.Content,
				Model:           resp.Model,
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
			}, nil
		}

		messages = append(messages, choice.Message)
		for _, tc := range choice.Message.ToolCalls {
			result := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}
	return nil, fmt.Errorf("agent 超过最大迭代次数")
}

func (a *Agent) AskStream(ctx context.Context, question string, history []ChatMessage, tools []openai.Tool, cb StreamCallback) error {
	if len(tools) == 0 {
		tools = a.getTools()
	}
	messages := a.buildMessages(question, history)

	for range maxIterations {
		stream, err := a.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			cb(StreamEvent{Type: "error", Content: fmt.Sprintf("AI 请求失败: %v", err)})
			return err
		}

		var assistantContent string
		toolCallMap := make(map[int]*openai.ToolCall)

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				cb(StreamEvent{Type: "error", Content: fmt.Sprintf("流式读取错误: %v", err)})
				return err
			}

			if len(resp.Choices) == 0 {
				continue
			}
			delta := resp.Choices[0].Delta

			if delta.Content != "" {
				assistantContent += delta.Content
				cb(StreamEvent{Type: "content", Content: delta.Content})
			}

			for _, tc := range delta.ToolCalls {
				if tc.ID == "" {
					continue
				}
				idx := 0
				if tc.Index != nil {
					idx = int(*tc.Index)
				}
				if _, ok := toolCallMap[idx]; !ok {
					toolCallMap[idx] = &openai.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: openai.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				} else {
					toolCallMap[idx].Function.Arguments += tc.Function.Arguments
				}
			}
		}

		assistantMsg := openai.ChatCompletionMessage{Role: "assistant", Content: assistantContent}
		if len(toolCallMap) > 0 {
			assistantMsg.ToolCalls = make([]openai.ToolCall, 0, len(toolCallMap))
			for i := range len(toolCallMap) {
				assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, *toolCallMap[i])
			}
		}
		messages = append(messages, assistantMsg)

		if len(toolCallMap) == 0 {
			totalTokens := 0
			cb(StreamEvent{Type: "done", PromptTokens: totalTokens})
			return nil
		}

		for i := range len(toolCallMap) {
			tc := toolCallMap[i]
			cb(StreamEvent{Type: "tool_call", Tool: tc.Function.Name, Content: tc.Function.Arguments})
			result := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			cb(StreamEvent{Type: "tool_result", Tool: tc.Function.Name, Content: result})
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}
	return fmt.Errorf("agent 超过最大迭代次数")
}

func (a *Agent) executeTool(ctx context.Context, name, args string) string {
	switch name {
		case "push":
			var params PushParams
			json.Unmarshal([]byte(args), &params)
			return ExecutePush(ctx, params)

	case "device_register":
		var params struct {
			DeviceToken string `json:"device_token"`
			Key         string `json:"key"`
		}
		json.Unmarshal([]byte(args), &params)
		return ExecuteDeviceRegister(ctx, params.DeviceToken, params.Key)

	case "device_check":
		var params struct {
			DeviceKey string `json:"device_key"`
		}
		json.Unmarshal([]byte(args), &params)
		return ExecuteDeviceCheck(ctx, params.DeviceKey)

	case "server_ping":
		return ExecuteServerPing(ctx)

	case "server_info":
		return ExecuteServerInfo(ctx)

	default:
		return fmt.Sprintf("未知工具: %s", name)
	}
}
