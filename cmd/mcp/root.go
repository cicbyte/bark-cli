package mcp

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "启动 MCP Server",
		Long: `以 stdio 模式启动 MCP Server，让 Claude Desktop、Claude Code 等 AI 客户端直接调用推送功能。

注册的 Tools:
  push              发送推送通知
  device_register   注册推送设备
  device_check      检查设备是否已注册
  server_ping       检查服务器连通性
  server_info       获取服务器信息`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runMCPServer(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
}
