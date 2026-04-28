package server

import (
	"fmt"

	"github.com/cicbyte/bark-cli/internal/client"
	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/cicbyte/bark-cli/internal/utils"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "管理 bark 实例",
	}

	cmd.AddCommand(addCommand())
	cmd.AddCommand(listCommand())
	cmd.AddCommand(useCommand())
	cmd.AddCommand(removeCommand())
	cmd.AddCommand(pingCommand())
	cmd.AddCommand(infoCommand())
	return cmd
}

func addCommand() *cobra.Command {
	var url, deviceToken, token string
	var timeout int
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "添加 bark 实例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg := utils.ConfigInstance.LoadConfig()

			if cfg.GetServer(name) != nil {
				output.Failed(fmt.Sprintf("实例 %s 已存在", name))
				return nil
			}
			if url == "" {
				url = "https://api.day.app"
			}
			if timeout <= 0 {
				timeout = 10
			}

				inst := &models.ServerInstance{
					URL:         url,
					DeviceToken: deviceToken,
					Token:       token,
					Timeout:     timeout,
				}
				cfg.AddServer(name, inst)

				// 如果有 device_token，自动注册并回填 device_key
				if deviceToken != "" {
					bark := newBarkClient(inst)
					resp, err := bark.Register("", deviceToken)
					if err != nil {
						output.Failed(fmt.Sprintf("实例 %s 添加成功，但设备注册失败: %v", name, err))
					} else {
						inst.DeviceKey = resp.DeviceKey
						output.Success(fmt.Sprintf("实例 %s 已添加，设备已注册 (key: %s)", name, resp.DeviceKey))
					}
				} else {
					output.Success(fmt.Sprintf("实例 %s 已添加", name))
				}

				utils.ConfigInstance.SaveConfig(cfg)
				common.AppConfigModel = cfg
				return nil
			return nil
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "bark-server 地址（默认 https://api.day.app）")
	cmd.Flags().StringVar(&deviceToken, "device-token", "", "设备推送 token")
	cmd.Flags().StringVar(&token, "token", "", "API 认证 token")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "请求超时（秒，默认 10）")
	return cmd
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有实例",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := utils.ConfigInstance.LoadConfig()
			if cfg.Servers == nil || len(cfg.Servers.Instances) == 0 {
				output.Failed("暂无实例，请运行: bark-cli server add <name>")
				return nil
			}

			if output.IsJSON("") {
				type item struct {
					Name        string `json:"name"`
					URL         string `json:"url"`
					DeviceKey   string `json:"device_key"`
					Default     bool   `json:"default"`
				}
				items := make([]item, 0, len(cfg.Servers.Instances))
				for name, inst := range cfg.Servers.Instances {
					items = append(items, item{
						Name: name, URL: inst.URL, DeviceKey: inst.DeviceKey,
						Default: name == cfg.Servers.Default,
					})
				}
				output.PrintJSON(items)
				return nil
			}

			rows := make([][]string, 0, len(cfg.Servers.Instances))
			for name, inst := range cfg.Servers.Instances {
				mark := ""
				if name == cfg.Servers.Default {
					mark = output.Green(" *")
				}
				rows = append(rows, []string{
					name + mark,
					inst.URL,
					maskValue(inst.DeviceKey),
				})
			}
			output.PrintTable([]string{"名称", "地址", "Device Key"}, rows)
			fmt.Println(output.Dim("  * 默认实例"))
			return nil
		},
	}
}

func useCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "设置默认实例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg := utils.ConfigInstance.LoadConfig()

			if cfg.GetServer(name) == nil {
				output.Failed(fmt.Sprintf("实例 %s 不存在", name))
				return nil
			}

			cfg.SetDefault(name)
			utils.ConfigInstance.SaveConfig(cfg)
			common.AppConfigModel = cfg
			output.Success(fmt.Sprintf("默认实例已设为 %s", name))
			return nil
		},
	}
}

func removeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "删除实例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg := utils.ConfigInstance.LoadConfig()

			if cfg.GetServer(name) == nil {
				output.Failed(fmt.Sprintf("实例 %s 不存在", name))
				return nil
			}

			cfg.RemoveServer(name)
			utils.ConfigInstance.SaveConfig(cfg)
			common.AppConfigModel = cfg
			output.Success(fmt.Sprintf("实例 %s 已删除", name))
			return nil
		},
	}
}

func pingCommand() *cobra.Command {
	var serverName string
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "检查服务器连通性",
		RunE: func(cmd *cobra.Command, args []string) error {
			inst := resolveInstance(serverName)
			if inst == nil {
				return nil
			}
			bark := newBarkClient(inst)

			if err := bark.Ping(); err != nil {
				output.Failed(fmt.Sprintf("连接失败: %v", err))
				return nil
			}

			if output.IsJSON("") {
				output.PrintJSON(map[string]string{"status": "ok", "server": inst.URL})
				return nil
			}
			output.Success(fmt.Sprintf("连接成功 (%s)", inst.URL))
			return nil
		},
	}
	cmd.Flags().StringVar(&serverName, "server", "", "指定实例名称")
	return cmd
}

func infoCommand() *cobra.Command {
	var serverName string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "查看服务器信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			inst := resolveInstance(serverName)
			if inst == nil {
				return nil
			}
			bark := newBarkClient(inst)

			info, err := bark.Info()
			if err != nil {
				output.Failed(fmt.Sprintf("获取信息失败: %v", err))
				return nil
			}

			if output.IsJSON("") {
				output.PrintJSON(info)
				return nil
			}

			rows := [][]string{
				{"版本", info.Version},
				{"构建时间", info.Build},
				{"架构", info.Arch},
				{"Commit", info.Commit},
				{"设备数", fmt.Sprintf("%d", info.Devices)},
			}
			output.PrintTable([]string{"项目", "值"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&serverName, "server", "", "指定实例名称")
	return cmd
}

func resolveInstance(name string) *models.ServerInstance {
	cfg := common.AppConfigModel
	if name == "" {
		name = cfg.Servers.Default
	}
	inst := cfg.GetServer(name)
	if inst == nil {
		if name == "" {
			output.Failed("未设置默认实例，请运行: bark-cli server add <name>")
		} else {
			output.Failed(fmt.Sprintf("实例 %s 不存在", name))
		}
		return nil
	}
	return inst
}

func newBarkClient(inst *models.ServerInstance) *client.Client {
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

func maskValue(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return "****"
	}
	return v[:4] + "****" + v[len(v)-4:]
}
