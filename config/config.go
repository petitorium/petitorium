package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Theme ThemeConfig `mapstructure:"theme"`
}

type ThemeConfig struct {
	BackgroundColor  string       `mapstructure:"backgroundColor"`
	ForegroundColor  string       `mapstructure:"foregroundColor"`
	BorderColor      string       `mapstructure:"borderColor"`
	BorderFocusColor string       `mapstructure:"borderFocusColor"`
	TitleColor       string       `mapstructure:"titleColor"`
	Borders          BorderConfig `mapstructure:"borders"`
	BordersFocus     BorderConfig `mapstructure:"bordersFocus"`
}

type BorderConfig struct {
	TopLeft     string `mapstructure:"topLeft"`
	TopRight    string `mapstructure:"topRight"`
	BottomLeft  string `mapstructure:"bottomLeft"`
	BottomRight string `mapstructure:"bottomRight"`
	Horizontal  string `mapstructure:"horizontal"`
	Vertical    string `mapstructure:"vertical"`
}

var C AppConfig

// defaultConfigYAML is the default configuration template.
// This is what `petitorium init` will create.
const defaultConfigYAML = `
theme:
  backgroundColor: "#000000"
  foregroundColor: "#FFFFFF"
  borderColor: "#888888"
  borderFocusColor: "#FFFFFF"
  titleColor: "#FFFFFF"

  borders:
    topLeft: "╭"
    topRight: "╮"
    bottomLeft: "╰"
    bottomRight: "╯"
    horizontal: "─"
    vertical: "│"

  bordersFocus:
    topLeft: "╭"
    topRight: "╮"
    bottomLeft: "╰"
    bottomRight: "╯"
    horizontal: "─"
    vertical: "│"
`

func LoadConfig() error {
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(strings.NewReader(defaultConfigYAML)); err != nil {
		return err
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".config", "petitorium")
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	_ = viper.ReadInConfig()

	return viper.Unmarshal(&C)
}

func InitConfig() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "petitorium")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		return configFile, errors.New("configuration file already exists")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(configFile, []byte(defaultConfigYAML), 0644); err != nil {
		return "", err
	}

	return configFile, nil
}
