package push

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cicbyte/bark-cli/internal/client"
	"github.com/cicbyte/bark-cli/internal/common"
	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	pushTitle    string
	pushSubtitle string
	pushBody     string
	pushSound    string
	pushLevel    string
	pushVolume   int
	pushBadge    int
	pushIcon     string
	pushGroup    string
	pushURL      string
	pushCopy     string
	pushArchive  bool
	pushCall     bool
	pushJSON     string
	pushServer   string
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [body]",
		Short: "发送推送通知",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPush,
	}

	cmd.Flags().StringVarP(&pushTitle, "title", "t", "", "通知标题")
	cmd.Flags().StringVarP(&pushSubtitle, "subtitle", "s", "", "通知副标题")
	cmd.Flags().StringVarP(&pushBody, "body", "b", "", "通知内容")
	cmd.Flags().StringVar(&pushServer, "server", "", "指定实例名称")
	cmd.Flags().StringVar(&pushSound, "sound", "1107", "提示音")
	cmd.Flags().StringVar(&pushLevel, "level", "", "通知级别 (critical/active/timeSensitive/passive)")
	cmd.Flags().IntVar(&pushVolume, "volume", 5, "提示音量 (0-10)")
	cmd.Flags().IntVar(&pushBadge, "badge", 0, "角标数字")
	cmd.Flags().StringVar(&pushIcon, "icon", "", "图标 URL")
	cmd.Flags().StringVar(&pushGroup, "group", "", "通知分组")
	cmd.Flags().StringVar(&pushURL, "url", "", "点击跳转 URL")
	cmd.Flags().StringVar(&pushCopy, "copy", "", "复制内容")
	cmd.Flags().BoolVar(&pushArchive, "archive", false, "保存通知")
	cmd.Flags().BoolVar(&pushCall, "call", false, "持续响铃 30 秒")
	cmd.Flags().StringVar(&pushJSON, "json", "", "从 JSON 字符串读取参数")

	return cmd
}

func runPush(cmd *cobra.Command, args []string) error {
	inst := resolveInstance(pushServer)
	if inst == nil {
		return nil
	}
	bark := newBarkClient(inst)

	req := &client.PushRequest{DeviceKey: inst.DeviceKey}

	if pushJSON != "" {
		var input map[string]any
		if err := json.Unmarshal([]byte(pushJSON), &input); err != nil {
			return fmt.Errorf("JSON 解析失败: %w", err)
		}
		applyJSONInput(req, input)
	}

	if pushTitle != "" {
		req.Title = pushTitle
	}
	if pushSubtitle != "" {
		req.Subtitle = pushSubtitle
	}
	if pushBody != "" {
		req.Body = pushBody
	}
	if len(args) > 0 && req.Body == "" {
		req.Body = args[0]
	}
	if req.Body == "" {
		stat, _ := os.Stdin.Stat()
		if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err == nil {
				req.Body = strings.TrimSpace(string(data))
			}
		}
	}
	if pushSound != "1107" {
		req.Sound = pushSound
	} else if req.Sound == "" {
		req.Sound = "1107"
	}
	if pushLevel != "" {
		req.Level = pushLevel
	}
	if pushVolume != 5 {
		req.Volume = pushVolume
	}
	if pushBadge != 0 {
		req.Badge = pushBadge
	}
	if pushIcon != "" {
		req.Icon = pushIcon
	}
	if pushGroup != "" {
		req.Group = pushGroup
	}
	if pushURL != "" {
		req.URL = pushURL
	}
	if pushCopy != "" {
		req.Copy = pushCopy
	}
	if pushArchive {
		req.IsArchive = "1"
	}
	if pushCall {
		req.Call = "1"
	}

	if req.Body == "" {
		return fmt.Errorf("请提供通知内容（使用 --body 或作为参数传入）")
	}

	if output.IsJSON("") {
		resp, err := bark.Push(req)
		if err != nil {
			return err
		}
		if output.IsJSONL("") {
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
		} else {
			output.PrintJSON(resp)
		}
		return nil
	}

	resp, err := bark.Push(req)
	if err != nil {
		output.Failed(fmt.Sprintf("推送失败: %v", err))
		return nil
	}
	if resp.Code != 200 {
		output.Failed(fmt.Sprintf("推送失败: %s", resp.Message))
		return nil
	}
	output.Success("推送成功")
	return nil
}

func applyJSONInput(req *client.PushRequest, input map[string]any) {
	if v, ok := input["device_key"].(string); ok {
		req.DeviceKey = v
	}
	if v, ok := input["title"].(string); ok {
		req.Title = v
	}
	if v, ok := input["subtitle"].(string); ok {
		req.Subtitle = v
	}
	if v, ok := input["body"].(string); ok {
		req.Body = v
	}
	if v, ok := input["sound"].(string); ok {
		req.Sound = v
	}
	if v, ok := input["level"].(string); ok {
		req.Level = v
	}
	if v, ok := input["volume"].(float64); ok {
		req.Volume = int(v)
	}
	if v, ok := input["badge"].(float64); ok {
		req.Badge = int(v)
	}
	if v, ok := input["icon"].(string); ok {
		req.Icon = v
	}
	if v, ok := input["group"].(string); ok {
		req.Group = v
	}
	if v, ok := input["url"].(string); ok {
		req.URL = v
	}
	if v, ok := input["copy"].(string); ok {
		req.Copy = v
	}
	if v, ok := input["isArchive"].(string); ok {
		req.IsArchive = v
	}
	if v, ok := input["call"].(string); ok {
		req.Call = v
	}
	if v, ok := input["image"].(string); ok {
		req.Image = v
	}
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

var _ = os.Stdout
var _ = strings.TrimSpace
