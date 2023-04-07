package config

type MonitorConfiguration struct {
	Enable bool        `yaml:"Enable"`
	Path   string      `yaml:"Path"`
	Auth   MonitorAuth `yaml:"Auth"`
}

type MonitorAuth struct {
	Enable   bool   `yaml:"Enable"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}
