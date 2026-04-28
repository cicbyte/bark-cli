package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cicbyte/bark-cli/internal/client"
	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runMCPServer() error {
	cfg := common.AppConfigModel
	if cfg.GetDefaultServer() == nil {
		return fmt.Errorf("未配置实例，请先运行: bark-cli server add <name>")
	}

	bark := newBarkClient(cfg)

	s := server.NewMCPServer(
		"bark-cli",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("push",
		mcp.WithDescription("发送推送通知到 iOS 设备"),
		mcp.WithString("body", mcp.Required(), mcp.Description("通知内容")),
		mcp.WithString("title", mcp.Description("通知标题")),
		mcp.WithString("subtitle", mcp.Description("通知副标题")),
		mcp.WithString("sound", mcp.Description("提示音")),
		mcp.WithString("level", mcp.Description("通知级别: critical|active|timeSensitive|passive")),
		mcp.WithNumber("volume", mcp.Description("音量(0-10)")),
		mcp.WithNumber("badge", mcp.Description("角标数字")),
		mcp.WithString("icon", mcp.Description("图标 URL")),
		mcp.WithString("group", mcp.Description("通知分组")),
		mcp.WithString("url", mcp.Description("点击跳转 URL")),
		mcp.WithString("copy", mcp.Description("复制内容")),
		mcp.WithString("key", mcp.Description("设备 key，不传则使用默认")),
		mcp.WithString("archive", mcp.Description("是否保存通知: true|false")),
		mcp.WithString("call", mcp.Description("持续响铃 30 秒: true|false")),
		mcp.WithString("image", mcp.Description("通知图片 URL")),
	), makeHandler(bark, handlePush))

	s.AddTool(mcp.NewTool("device_register",
		mcp.WithDescription("注册新的推送设备"),
		mcp.WithString("device_token", mcp.Required(), mcp.Description("Apple 设备推送 token")),
		mcp.WithString("key", mcp.Description("指定设备 key，不传则自动生成")),
	), makeHandler(bark, handleDeviceRegister))

	s.AddTool(mcp.NewTool("device_check",
		mcp.WithDescription("检查设备是否已注册"),
		mcp.WithString("device_key", mcp.Required(), mcp.Description("设备 key")),
	), makeHandler(bark, handleDeviceCheck))

	s.AddTool(mcp.NewTool("server_ping",
		mcp.WithDescription("检查 bark-server 是否可连通"),
	), makeHandler(bark, handleServerPing))

	s.AddTool(mcp.NewTool("server_info",
		mcp.WithDescription("获取 bark-server 服务器信息"),
	), makeHandler(bark, handleServerInfo))

	return server.ServeStdio(s)
}

type handlerFunc func(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

func makeHandler(bark *client.Client, fn handlerFunc) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return fn(ctx, bark, req)
	}
}

func handlePush(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, err := req.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pushReq := &client.PushRequest{Body: body}
	if v := req.GetString("title", ""); v != "" {
		pushReq.Title = v
	}
	if v := req.GetString("subtitle", ""); v != "" {
		pushReq.Subtitle = v
	}
	if v := req.GetString("sound", ""); v != "" {
		pushReq.Sound = v
	}
	if v := req.GetString("level", ""); v != "" {
		pushReq.Level = v
	}
	if v := int(req.GetFloat("volume", 0)); v > 0 {
		pushReq.Volume = v
	}
	if v := int(req.GetFloat("badge", 0)); v > 0 {
		pushReq.Badge = v
	}
	if v := req.GetString("icon", ""); v != "" {
		pushReq.Icon = v
	}
	if v := req.GetString("group", ""); v != "" {
		pushReq.Group = v
	}
	if v := req.GetString("url", ""); v != "" {
		pushReq.URL = v
	}
	if v := req.GetString("copy", ""); v != "" {
		pushReq.Copy = v
	}
	if v := req.GetString("key", ""); v != "" {
		pushReq.DeviceKey = v
	}
	if v := req.GetString("archive", ""); v == "true" {
		pushReq.IsArchive = "1"
	}
	if v := req.GetString("call", ""); v == "true" {
			pushReq.Call = "1"
		}
		if v := req.GetString("image", ""); v != "" {
			pushReq.Image = v
		}

	resp, err := bark.Push(pushReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("推送失败: %v", err)), nil
	}
	if resp.Code != 200 {
		return mcp.NewToolResultError(fmt.Sprintf("推送失败: %s", resp.Message)), nil
	}

	result := map[string]any{"status": "ok", "message": "推送成功"}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleDeviceRegister(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token, err := req.RequireString("device_token")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	key := req.GetString("key", "")

	resp, err := bark.Register(key, token)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("注册失败: %v", err)), nil
	}

	result := map[string]any{"status": "ok", "device_key": resp.DeviceKey}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleDeviceCheck(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, err := req.RequireString("device_key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := bark.CheckDevice(key); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("设备未注册: %v", err)), nil
	}

	result := map[string]any{"status": "ok", "message": "设备已注册"}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleServerPing(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := bark.Ping(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("服务器不可达: %v", err)), nil
	}

	result := map[string]any{"status": "ok", "message": "服务器连接正常"}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleServerInfo(ctx context.Context, bark *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info, err := bark.Info()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取信息失败: %v", err)), nil
	}

	result := map[string]any{
		"version": info.Version,
		"arch":    info.Arch,
		"devices": info.Devices,
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
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
