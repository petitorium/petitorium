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
	Theme        ThemeConfig        `mapstructure:"theme"`
	UI           UIConfig           `mapstructure:"ui"`
	MethodColors MethodColorsConfig `mapstructure:"methodColors"`
	SyntaxTheme  string             `mapstructure:"syntaxTheme"`
}

type ThemeConfig struct {
	BackgroundColor           string       `mapstructure:"backgroundColor"`
	ForegroundColor           string       `mapstructure:"foregroundColor"`
	BorderColor               string       `mapstructure:"borderColor"`
	BorderFocusColor          string       `mapstructure:"borderFocusColor"`
	TitleColor                string       `mapstructure:"titleColor"`
	SelectionBackground       string       `mapstructure:"selectionBackground"`
	ActiveTabColor            string       `mapstructure:"activeTabColor"`
	ButtonSelectedColor       string       `mapstructure:"buttonSelectedColor"`
	DropdownFocusedBackground string       `mapstructure:"dropdownFocusedBackground"`
	Borders                   BorderConfig `mapstructure:"borders"`
	BordersFocus              BorderConfig `mapstructure:"bordersFocus"`
}

type BorderConfig struct {
	TopLeft     string `mapstructure:"topLeft"`
	TopRight    string `mapstructure:"topRight"`
	BottomLeft  string `mapstructure:"bottomLeft"`
	BottomRight string `mapstructure:"bottomRight"`
	Horizontal  string `mapstructure:"horizontal"`
	Vertical    string `mapstructure:"vertical"`
}

type UIConfig struct {
	CollectionExpansion      string `mapstructure:"collectionExpansion"`      // "closed", "expanded", "remember"
	CollectionIcon           string `mapstructure:"collectionIcon"`           // icon to display before closed collection names
	CollectionExpandedIcon   string `mapstructure:"collectionExpandedIcon"`   // icon to display before expanded collection names
	SelectedRequestIcon      string `mapstructure:"selectedRequestIcon"`      // icon to display before selected request names
	SelectedRequestIconColor string `mapstructure:"selectedRequestIconColor"` // color for selected request icon
	HeaderRemoveIcon         string `mapstructure:"headerRemoveIcon"`         // icon for removing headers
}

type MethodColorsConfig struct {
	GET     string `mapstructure:"GET"`
	POST    string `mapstructure:"POST"`
	PUT     string `mapstructure:"PUT"`
	PATCH   string `mapstructure:"PATCH"`
	DELETE  string `mapstructure:"DELETE"`
	OPTIONS string `mapstructure:"OPTIONS"`
	HEAD    string `mapstructure:"HEAD"`
	Default string `mapstructure:"default"`
}

var C AppConfig

// defaultConfigYAML is the default configuration template.
// This is what `petitorium init` will create.
const defaultConfigYAML = `theme:
  backgroundColor: "#102529"
  foregroundColor: "#e4e4e4"
  borderColor: "#95CEDA"
  borderFocusColor: "#FF9F77"
  titleColor: "#EBEBEB"
  selectionBackground: "#1B4248"  # background color for selected items
  activeTabColor: "#FF9F77"       # color for active tab indicator
  buttonSelectedColor: "#FFD700"  # color for selected buttons
  dropdownFocusedBackground: "#636DA6"  # background color for focused dropdown

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

ui:
  collectionExpansion: "closed"       # "closed", "expanded", or "remember"
  collectionIcon: ""
  collectionExpandedIcon: ""
  selectedRequestIcon: ""
  selectedRequestIconColor: "#c8d3f5"
  headerRemoveIcon: "✕"

# Syntax highlighting theme (chroma themes)
# Popular options: github-dark, dracula, monokai, solarized-dark, nord, one-dark, vim, github
# Run 'petitorium themes' to see all available themes
syntaxTheme: "tokyonight-night"

methodColors:
  GET: "#6EA5A0"
  POST: "#FF00FF"
  PUT: "#FF9F77"
  PATCH: "#FF9F77"
  DELETE: "#FB4F49"
  OPTIONS: "#FFA500"
  HEAD: "#800080"
  default: "#888888"
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
