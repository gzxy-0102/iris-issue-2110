package config

type AuthConfiguration struct {
	Expires int64  `yaml:"Expires"`
	Secret  string `yaml:"Secret"`
}
