package client

import (
	"os"
	"testing"

	"github.com/cicbyte/bark-cli/internal/models"
	"github.com/cicbyte/bark-cli/internal/utils"
)

var (
	testServerURL   string
	testDeviceToken string
	testDeviceKey   string
	skipIntegration bool
)

func init() {
	testServerURL = os.Getenv("BARK_SERVER_URL")
	if testServerURL == "" {
		testServerURL = "https://api.day.app"
	}

	cfg := utils.ConfigInstance.LoadConfig()
	inst := cfg.GetDefaultServer()
	if inst != nil {
		testDeviceToken = inst.DeviceToken
	}

	if testDeviceToken == "" {
		skipIntegration = true
	}
}

func newTestClient() *Client {
	return New(testServerURL, 15, nil)
}

func TestPing(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()
	if err := c.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	t.Log("ping ok")
}

func TestInfo(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()
	info, err := c.Info()
	if err != nil {
		t.Fatalf("info failed: %v", err)
	}
	if info.Version == "" {
		t.Fatal("expected non-empty version")
	}
	t.Logf("version=%s arch=%s devices=%d", info.Version, info.Arch, info.Devices)
}

func TestRegisterAndCheck(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()

	resp, err := c.Register("", testDeviceToken)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.DeviceKey == "" {
		t.Fatal("expected non-empty device_key")
	}
	testDeviceKey = resp.DeviceKey
	t.Logf("registered: key=%s", testDeviceKey)

	if err := c.CheckDevice(testDeviceKey); err != nil {
		t.Fatalf("check device failed: %v", err)
	}
	t.Log("device check ok")
}

func TestPushSimple(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()

	resp, err := c.Register("", testDeviceToken)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	testDeviceKey = resp.DeviceKey

	result, err := c.Push(&PushRequest{
		DeviceKey: testDeviceKey,
		Title:     "bark-cli 测试",
		Body:      "集成测试推送 - TestPushSimple",
	})
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if result.Code != 200 {
		t.Fatalf("push returned code %d: %s", result.Code, result.Message)
	}
	t.Log("push ok")
}

func TestPushWithAllParams(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()

	resp, err := c.Register("", testDeviceToken)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	result, err := c.Push(&PushRequest{
		DeviceKey: resp.DeviceKey,
		Title:     "完整参数测试",
		Subtitle:  "子标题",
		Body:      "测试全部推送参数",
		Sound:     "1107",
		Level:     "active",
		Volume:    7,
		Badge:     1,
		Group:     "test",
		IsArchive: "1",
	})
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if result.Code != 200 {
		t.Fatalf("push returned code %d: %s", result.Code, result.Message)
	}
	t.Log("push with all params ok")
}

func TestCheckDevice_NotFound(t *testing.T) {
	if skipIntegration {
		t.Skip("skipping: default instance has no device_token")
	}
	c := newTestClient()
	err := c.CheckDevice("nonexistent_device_key_12345")
	if err == nil {
		t.Fatal("expected error for unregistered device")
	}
	t.Logf("correctly returned error: %v", err)
}

func TestPing_Connectivity(t *testing.T) {
	c := New("https://127.0.0.1:1", 3, nil)
	err := c.Ping()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	t.Logf("correctly detected unreachable: %v", err)
}

func TestConfigModel(t *testing.T) {
	cfg := &models.AppConfig{}
	cfg.Servers = &models.Servers{
		Default: "iphone",
		Instances: map[string]*models.ServerInstance{
			"iphone": {
				URL:         "https://api.day.app",
				DeviceToken: "test123",
				DeviceKey:   "key456",
				Timeout:     10,
			},
		},
	}

	inst := cfg.GetDefaultServer()
	if inst == nil || inst.URL != "https://api.day.app" {
		t.Fatal("GetDefaultServer failed")
	}

	inst2 := cfg.GetServer("iphone")
	if inst2 == nil || inst2.DeviceToken != "test123" {
		t.Fatal("GetServer failed")
	}

	names := cfg.ListServers()
	if len(names) != 1 || names[0] != "iphone" {
		t.Fatal("ListServers failed")
	}
}
