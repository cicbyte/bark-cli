package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/cicbyte/bark-cli/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理配置",
	}

	cmd.AddCommand(listCommand())
	cmd.AddCommand(getCommand())
	cmd.AddCommand(setCommand())
	return cmd
}

// --- list ---

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := utils.ConfigInstance.LoadConfig()

			if output.IsJSONL("") {
				out, err := json.Marshal(cfg)
				if err == nil {
					fmt.Println(string(out))
				}
				return nil
			}

			if output.IsJSON("") {
				output.PrintJSON(cfg)
				return nil
			}

			type entry struct {
				key       string
				section   string
				value     string
				sensitive bool
			}

			entries := []entry{}

			if cfg.Servers != nil {
				entries = append(entries,
					entry{key: "servers.default", section: "Servers", value: cfg.Servers.Default},
				)
				for name, inst := range cfg.Servers.Instances {
					mark := ""
					if name == cfg.Servers.Default {
						mark = " *"
					}
					entries = append(entries,
						entry{key: fmt.Sprintf("servers.instances.%s", name), section: "Servers", value: fmt.Sprintf("url=%s device_key=%s%s", inst.URL, maskValue(inst.DeviceKey), mark)},
					)
				}
			}

			entries = append(entries,
				entry{key: "ai.provider", section: "AI", value: cfg.AI.Provider},
				entry{key: "ai.base_url", section: "AI", value: cfg.AI.BaseURL},
				entry{key: "ai.model", section: "AI", value: cfg.AI.Model},
				entry{key: "ai.api_key", section: "AI", value: cfg.AI.ApiKey, sensitive: true},
				entry{key: "output.format", section: "Output", value: cfg.GetOutputFormat()},
				entry{key: "log.level", section: "Log", value: cfg.Log.Level},
			)

			headers := []string{"KEY", "VALUE"}
			rows := make([][]string, 0, len(entries))
			currentSection := ""

			for _, e := range entries {
				if e.section != currentSection {
					currentSection = e.section
					rows = append(rows, []string{fmt.Sprintf("[%s]", currentSection), ""})
				}

				displayVal := e.value
				if e.sensitive {
					displayVal = maskValue(e.value)
				}
				if displayVal == "" {
					displayVal = "(未设置)"
				}

				rows = append(rows, []string{e.key, displayVal})
			}

			output.PrintTable(headers, rows)
			return nil
		},
	}
}

// --- get ---

var getShowFlag bool

func getCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "查看配置项的值",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}
	cmd.Flags().BoolVar(&getShowFlag, "show", false, "显示敏感字段的明文值")
	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	cfg := utils.ConfigInstance.LoadConfig()
	key := args[0]

	value, ok, sensitive := getConfigValue(cfg, key)
	if !ok {
		return fmt.Errorf("未知配置项 '%s'\n使用 'bark-cli config list' 查看所有配置项", key)
	}

	if value == "" {
		fmt.Printf("%s: (未设置)\n", key)
		return nil
	}

	if sensitive && !getShowFlag {
		fmt.Printf("%s: %s\n", key, maskValue(value))
		fmt.Println("使用 --show 查看明文")
		return nil
	}

	fmt.Printf("%s: %s\n", key, value)
	return nil
}

// --- set ---

var sensitiveKeys = map[string]bool{
	"ai.api_key": true,
}

func setCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> [value]",
		Short: "设置配置项的值",
		Long: `设置指定配置项的值。敏感字段（如 api_key）不提供 value 时交互式输入（不回显）。

示例:
  bark-cli config set ai.provider openai
  bark-cli config set ai.model gpt-4o
  bark-cli config set ai.api_key sk-xxx
  bark-cli config set ai.api_key
  bark-cli config set log.level debug
  bark-cli config set output.format json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runSet,
	}
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	cfg := utils.ConfigInstance.LoadConfig()

	if _, ok, _ := getConfigValue(cfg, key); !ok {
		return fmt.Errorf("未知配置项 '%s'\n使用 'bark-cli config list' 查看所有配置项", key)
	}

	var value string

	if len(args) >= 2 {
		value = args[1]
	} else if sensitiveKeys[key] {
		fmt.Printf("请输入 %s: ", key)
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("读取输入失败")
		}
		value = string(raw)
	} else {
		fmt.Printf("请输入 %s: ", key)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		value = strings.TrimSpace(line)
	}

	if value == "" {
		return fmt.Errorf("值不能为空")
	}

	if err := setConfigValue(cfg, key, value); err != nil {
		return err
	}

	utils.ConfigInstance.SaveConfig(cfg)
	fmt.Printf("%s 已更新\n", key)
	return nil
}

// --- helpers ---

func getConfigValue(cfg *models.AppConfig, key string) (string, bool, bool) {
	switch key {
	case "ai.provider":
		return cfg.AI.Provider, true, false
	case "ai.base_url":
		return cfg.AI.BaseURL, true, false
	case "ai.model":
		return cfg.AI.Model, true, false
	case "ai.api_key":
		return cfg.AI.ApiKey, true, true
	case "ai.max_tokens":
		return strconv.Itoa(cfg.AI.MaxTokens), true, false
	case "ai.temperature":
		return fmt.Sprintf("%.2f", cfg.AI.Temperature), true, false
	case "ai.timeout":
		return strconv.Itoa(cfg.AI.Timeout), true, false
	case "output.format":
		return cfg.GetOutputFormat(), true, false
	case "log.level":
		return cfg.Log.Level, true, false
	case "log.max_size":
		return strconv.Itoa(cfg.Log.MaxSize), true, false
	case "log.max_backups":
		return strconv.Itoa(cfg.Log.MaxBackups), true, false
	case "log.max_age":
		return strconv.Itoa(cfg.Log.MaxAge), true, false
	case "log.compress":
		return strconv.FormatBool(cfg.Log.Compress), true, false
	default:
		return "", false, false
	}
}

func setConfigValue(cfg *models.AppConfig, key, value string) error {
	switch key {
	case "ai.provider":
		cfg.AI.Provider = value
	case "ai.base_url":
		cfg.AI.BaseURL = value
	case "ai.model":
		cfg.AI.Model = value
	case "ai.api_key":
		cfg.AI.ApiKey = value
	case "ai.max_tokens":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		cfg.AI.MaxTokens = v
	case "ai.temperature":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("无效的浮点数值: %s", value)
		}
		cfg.AI.Temperature = v
	case "ai.timeout":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		cfg.AI.Timeout = v
	case "output.format":
		cfg.Output.Format = value
	case "log.level":
		cfg.Log.Level = value
	case "log.max_size":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		cfg.Log.MaxSize = v
	case "log.max_backups":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		cfg.Log.MaxBackups = v
	case "log.max_age":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		cfg.Log.MaxAge = v
	case "log.compress":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("无效的布尔值: %s (true/false)", value)
		}
		cfg.Log.Compress = v
	}
	return nil
}

func maskValue(v string) string {
	if v == "" {
		return ""
	}
	runes := []rune(v)
	if len(runes) <= 8 {
		return "******"
	}
	return string(runes[:4]) + "****" + string(runes[len(runes)-4:])
}
