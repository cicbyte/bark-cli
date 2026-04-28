package device

import (
	"fmt"
	"strings"

	"github.com/cicbyte/bark-cli/internal/client"
	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/cicbyte/bark-cli/internal/utils"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "管理推送设备",
	}

	cmd.AddCommand(listCommand())
	cmd.AddCommand(checkCommand())
	return cmd
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有设备实例",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := utils.ConfigInstance.LoadConfig()
			if cfg.Servers == nil || len(cfg.Servers.Instances) == 0 {
				output.Failed("暂无实例，请运行: bark-cli server add <name> --device-token <token>")
				return nil
			}

			if output.IsJSON("") {
				type item struct {
					Name        string `json:"name"`
					URL         string `json:"url"`
					DeviceKey   string `json:"device_key"`
					DeviceToken string `json:"device_token"`
					Default     bool   `json:"default"`
				}
				items := make([]item, 0, len(cfg.Servers.Instances))
				for name, inst := range cfg.Servers.Instances {
					items = append(items, item{
						Name: name, URL: inst.URL, DeviceKey: inst.DeviceKey,
						DeviceToken: maskToken(inst.DeviceToken),
						Default: name == cfg.Servers.Default,
					})
				}
				if output.IsJSONL("") {
					output.PrintJSONL(items)
				} else {
					output.PrintJSON(items)
				}
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
					maskKey(inst.DeviceKey),
					maskToken(inst.DeviceToken),
				})
			}
			output.PrintTable([]string{"名称", "地址", "Device Key", "Device Token"}, rows)
			fmt.Println(output.Dim("  * 默认实例"))
			return nil
		},
	}
}

func checkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check <instance|device_key>",
		Short: "检查设备是否已注册",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := common.AppConfigModel
			input := args[0]
			key := ""
			var checkInst *models.ServerInstance

			if inst := cfg.GetServer(input); inst != nil {
				key = inst.DeviceKey
				checkInst = inst
			}

			if key == "" {
				for _, inst := range cfg.Servers.Instances {
					if inst.DeviceKey == input {
						key = inst.DeviceKey
						checkInst = inst
						break
					}
				}
			}

			if key == "" {
				for _, inst := range cfg.Servers.Instances {
					if inst.DeviceToken == input {
						key = inst.DeviceKey
						checkInst = inst
						break
					}
				}
			}

			if key == "" {
				key = input
			}

			if key == "" {
				output.Failed("未找到匹配的设备")
				return nil
			}

			bark := newBarkClientForCheck(cfg, checkInst)
			if err := bark.CheckDevice(key); err != nil {
				output.Failed(fmt.Sprintf("设备 %s 未注册", maskKey(key)))
				return nil
			}

			if output.IsJSON("") {
				output.PrintJSON(map[string]string{"device_key": key, "registered": "true"})
				return nil
			}
			output.Success(fmt.Sprintf("设备 %s 已注册", maskKey(key)))
			return nil
		},
	}
}

func newBarkClientForCheck(cfg *models.AppConfig, inst *models.ServerInstance) *client.Client {
	if inst != nil {
		return newBarkClient(inst)
	}
	defaultInst := cfg.GetDefaultServer()
	if defaultInst != nil {
		return newBarkClient(defaultInst)
	}
	return client.New("https://api.day.app", 10, nil)
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

func maskKey(v string) string {
	if v == "" {
		return output.Dim("(未注册)")
	}
	if len(v) <= 8 {
		return "****"
	}
	return v[:4] + "****" + v[len(v)-4:]
}

func maskToken(v string) string {
	if v == "" {
		return output.Dim("(未设置)")
	}
	if len(v) <= 8 {
		return "****"
	}
	return v[:4] + "****" + v[len(v)-4:]
}

var _ = strings.TrimSpace
