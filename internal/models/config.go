package models

type AppConfig struct {
	Version string  `yaml:"version"`
	Servers *Servers `yaml:"servers"`
	AI      struct {
		Provider    string  `yaml:"provider"`
		BaseURL     string  `yaml:"base_url"`
		ApiKey      string  `yaml:"api_key"`
		Model       string  `yaml:"model"`
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     int     `yaml:"timeout"`
	} `yaml:"ai"`
	Output struct {
		Format string `yaml:"format"`
	} `yaml:"output"`
	Log struct {
		Level      string `yaml:"level"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

type ServerInstance struct {
	URL         string `yaml:"url"`
	DeviceToken string `yaml:"device_token"`
	DeviceKey   string `yaml:"device_key"`
	Token       string `yaml:"token"`
	Timeout     int    `yaml:"timeout"`
}

type Servers struct {
	Default   string                     `yaml:"default"`
	Instances map[string]*ServerInstance `yaml:"instances"`
}

func (c *AppConfig) GetDefaultServer() *ServerInstance {
	if c.Servers == nil || c.Servers.Default == "" {
		return nil
	}
	return c.Servers.Instances[c.Servers.Default]
}

func (c *AppConfig) GetServer(name string) *ServerInstance {
	if c.Servers == nil {
		return nil
	}
	return c.Servers.Instances[name]
}

func (c *AppConfig) ListServers() []string {
	if c.Servers == nil || len(c.Servers.Instances) == 0 {
		return nil
	}
	names := make([]string, 0, len(c.Servers.Instances))
	for name := range c.Servers.Instances {
		names = append(names, name)
	}
	return names
}

func (c *AppConfig) AddServer(name string, inst *ServerInstance) {
	if c.Servers == nil {
		c.Servers = &Servers{Instances: make(map[string]*ServerInstance)}
	}
	if c.Servers.Instances == nil {
		c.Servers.Instances = make(map[string]*ServerInstance)
	}
	c.Servers.Instances[name] = inst
	if c.Servers.Default == "" {
		c.Servers.Default = name
	}
}

func (c *AppConfig) RemoveServer(name string) {
	if c.Servers == nil || c.Servers.Instances == nil {
		return
	}
	delete(c.Servers.Instances, name)
	if c.Servers.Default == name {
		c.Servers.Default = ""
		for n := range c.Servers.Instances {
			c.Servers.Default = n
			break
		}
	}
}

func (c *AppConfig) SetDefault(name string) {
	if c.Servers == nil {
		c.Servers = &Servers{Instances: make(map[string]*ServerInstance)}
	}
	c.Servers.Default = name
}

func (c *AppConfig) GetOutputFormat() string {
	if c.Output.Format == "" {
		return "table"
	}
	return c.Output.Format
}
