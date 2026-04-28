package cmd

import (
	"fmt"
	"os"

	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/log"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/cicbyte/bark-cli/internal/utils"
	"github.com/cicbyte/bark-cli/cmd/chat"
	"github.com/cicbyte/bark-cli/cmd/config"
	"github.com/cicbyte/bark-cli/cmd/device"
	"github.com/cicbyte/bark-cli/cmd/mcp"
	"github.com/cicbyte/bark-cli/cmd/push"
	"github.com/cicbyte/bark-cli/cmd/server"
	"github.com/cicbyte/bark-cli/cmd/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var globalFormat string

var rootCmd = &cobra.Command{
	Use:   "bark-cli",
	Short: "Bark 推送通知 CLI 工具",
	Long:  `bark-cli 是 bark-server 的命令行客户端，支持推送通知、设备管理、AI 对话等功能。`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalFormat, "format", "table", "输出格式 (table|json|jsonl)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		output.SetFormat(globalFormat)

		// Skip init for certain commands
		skipCommands := map[string]bool{
			"config": true,
			"help":   true,
			"completion": true,
		}
		if skipCommands[cmd.Name()] {
			return nil
		}
		// Check parent command
		if cmd.Parent() != nil && skipCommands[cmd.Parent().Name()] {
			return nil
		}

		// Initialize app dirs
		if err := utils.InitAppDirs(); err != nil {
			fmt.Printf("初始化目录失败: %v\n", err)
			os.Exit(1)
		}

		// Load config
		common.AppConfigModel = utils.ConfigInstance.LoadConfig()

		// Init logging
		if err := log.Init(utils.ConfigInstance.GetLogPath()); err != nil {
			fmt.Printf("日志初始化失败: %v\n", err)
			os.Exit(1)
		}

		return nil
	}

	rootCmd.AddCommand(config.GetCommand())
	rootCmd.AddCommand(push.GetCommand())
	rootCmd.AddCommand(device.GetCommand())
	rootCmd.AddCommand(server.GetCommand())
	rootCmd.AddCommand(chat.GetCommand())
	rootCmd.AddCommand(mcp.GetCommand())
	rootCmd.AddCommand(version.GetCommand())
}

var _ = zap.Field{}
