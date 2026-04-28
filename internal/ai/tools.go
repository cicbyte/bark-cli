package ai

import (
	"context"
	"fmt"
	"time"
)

type PushParams struct {
	Body     string `json:"body"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Sound    string `json:"sound"`
	Level    string `json:"level"`
	Volume   int    `json:"volume"`
	Badge    int    `json:"badge"`
	Icon     string `json:"icon"`
	Group    string `json:"group"`
	URL      string `json:"url"`
	Copy     string `json:"copy"`
	Key      string `json:"key"`
	Archive  bool   `json:"archive"`
	Call     bool   `json:"call"`
}

func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ExecutePush is implemented by cmd/chat/chat.go to avoid circular dependency
var (
	ExecutePush          func(ctx context.Context, params PushParams) string
	ExecuteDeviceRegister func(ctx context.Context, token, key string) string
	ExecuteDeviceCheck   func(ctx context.Context, key string) string
	ExecuteServerPing     func(ctx context.Context) string
	ExecuteServerInfo    func(ctx context.Context) string
)

func init() {
	// Default no-op implementations
	ExecutePush = func(ctx context.Context, params PushParams) string {
		return fmt.Sprintf("[push] 推送通知: title=%s body=%s", params.Title, params.Body)
	}
	ExecuteDeviceRegister = func(ctx context.Context, token, key string) string {
		return fmt.Sprintf("[device_register] 注册设备: token=%s key=%s", token[:min(8, len(token))]+"...", key)
	}
	ExecuteDeviceCheck = func(ctx context.Context, key string) string {
		return fmt.Sprintf("[device_check] 检查设备: key=%s", key)
	}
	ExecuteServerPing = func(ctx context.Context) string {
		return "[server_ping] 服务器连通性检查"
	}
	ExecuteServerInfo = func(ctx context.Context) string {
		return "[server_info] 获取服务器信息"
	}
}
