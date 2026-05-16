package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/petitorium/petitorium/plugins"
)

type AppConfig struct {
	Theme               ThemeConfig          `mapstructure:"theme"`
	UI                  UIConfig             `mapstructure:"ui"`
	MethodColors        MethodColorsConfig   `mapstructure:"methodColors"`
	StatusColors        StatusColorsConfig   `mapstructure:"statusColors"`
	SyntaxTheme         string               `mapstructure:"syntaxTheme"`
	SelectedEnvironment string               `mapstructure:"selectedEnvironment"`
	ThemeOverrides      ThemeConfig          `mapstructure:"themeOverrides"` // Explicit user overrides for theme colors
	RequestTimeout      int                  `mapstructure:"requestTimeout"` // Timeout in seconds for HTTP requests
	MaxResponseHistory  int                  `mapstructure:"maxResponseHistory"`
	Plugins             plugins.PluginConfig `mapstructure:"plugins"`
	Shortcuts           ShortcutsConfig      `mapstructure:"shortcuts"`
	DisableVersionCheck bool                 `mapstructure:"disableVersionCheck"` // Disable latest version check
}

type ThemeConfig struct {
	BackgroundColor             string       `mapstructure:"backgroundColor"`
	ForegroundColor             string       `mapstructure:"foregroundColor"`
	BorderColor                 string       `mapstructure:"borderColor"`
	BorderFocusColor            string       `mapstructure:"borderFocusColor"`
	TitleColor                  string       `mapstructure:"titleColor"`
	SelectionBackground         string       `mapstructure:"selectionBackground"`
	TreeSelectionBackground     string       `mapstructure:"treeSelectionBackground"`
	SelectedBackground          string       `mapstructure:"selectedBackground"`
	SelectedForeground          string       `mapstructure:"selectedForeground"`
	ActiveTabColor              string       `mapstructure:"activeTabColor"`
	ButtonBackgroundColor       string       `mapstructure:"buttonBackgroundColor"`
	ButtonSelectedColor         string       `mapstructure:"buttonSelectedColor"`
	DropdownFocusedBackground   string       `mapstructure:"dropdownFocusedBackground"`
	InputBackgroundColor        string       `mapstructure:"inputBackgroundColor"`        // Background color for input fields (slightly lighter than background for contrast)
	InputBackgroundLighterColor string       `mapstructure:"inputBackgroundLighterColor"` // Lighter variant of input background for additional contrast states
	LabelColor                  string       `mapstructure:"labelColor"`                  // Color for form labels
	ValueColor                  string       `mapstructure:"valueColor"`                  // Color for form values
	Borders                     BorderConfig `mapstructure:"borders"`
	BordersFocus                BorderConfig `mapstructure:"bordersFocus"`
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
	CollectionExpansion           string `mapstructure:"collectionExpansion"`           // "closed", "expanded", "remember"
	CollectionIcon                string `mapstructure:"collectionIcon"`                // icon to display before closed collection names
	CollectionExpandedIcon        string `mapstructure:"collectionExpandedIcon"`        // icon to display before expanded collection names
	SelectedRequestIcon           string `mapstructure:"selectedRequestIcon"`           // icon to display before selected request names
	SelectedRequestIconColor      string `mapstructure:"selectedRequestIconColor"`      // color for selected request icon
	HeaderRemoveIcon              string `mapstructure:"headerRemoveIcon"`              // icon for removing headers
	MultipartRemoveIcon           string `mapstructure:"multipartRemoveIcon"`           // icon for removing multipart fields
	ConfigButtonIcon              string `mapstructure:"configButtonIcon"`              // icon for environment config button
	DropdownIndicator             string `mapstructure:"dropdownIndicator"`             // indicator for dropdown
	ButtonFlashDuration           int    `mapstructure:"buttonFlashDuration"`           // duration of button flash in milliseconds
	FileBrowserFolderIcon         string `mapstructure:"fileBrowserFolderIcon"`         // icon for folders in file browser
	FileBrowserFolderExpandedIcon string `mapstructure:"fileBrowserFolderExpandedIcon"` // icon for expanded folders in file browser
	FileBrowserFileIcon           string `mapstructure:"fileBrowserFileIcon"`           // icon for files in file browser
	CopyResponseIcon              string `mapstructure:"copyResponseIcon"`              // icon for copy response button
	ExportResponseIcon            string `mapstructure:"exportResponseIcon"`            // icon for export/save response button
	CheckboxOn                    string `mapstructure:"checkboxOn"`                    // checkbox enabled state icon
	CheckboxOff                   string `mapstructure:"checkboxOff"`                   // checkbox disabled state icon
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

type StatusColorsConfig struct {
	Success         string `mapstructure:"success"`         // 2xx background
	SuccessText     string `mapstructure:"successText"`     // 2xx text
	Redirection     string `mapstructure:"redirection"`     // 3xx background
	RedirectionText string `mapstructure:"redirectionText"` // 3xx text
	ClientError     string `mapstructure:"clientError"`     // 4xx background
	ClientErrorText string `mapstructure:"clientErrorText"` // 4xx text
	ServerError     string `mapstructure:"serverError"`     // 5xx background
	ServerErrorText string `mapstructure:"serverErrorText"` // 5xx text
	Default         string `mapstructure:"default"`         // Unknown background
	DefaultText     string `mapstructure:"defaultText"`     // Unknown text
}

type ShortcutsConfig struct {
	JumpToWorkspace   string `mapstructure:"jumpToWorkspace"`
	JumpToEnvironment string `mapstructure:"jumpToEnvironment"`
	JumpToCollections string `mapstructure:"jumpToCollections"`
	JumpToURLBar      string `mapstructure:"jumpToURLBar"`
	JumpToRequest     string `mapstructure:"jumpToRequest"`
	JumpToResponse    string `mapstructure:"jumpToResponse"`
	SendRequest       string `mapstructure:"sendRequest"`
}

var C AppConfig

// defaultConfigYAML is the default configuration template.
// This is what `petitorium init` will create.
const defaultConfigYAML = `theme:
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

# You can override specific theme colors here. These will take precedence over the generated theme.
themeOverrides:
  # backgroundColor: "#102529"
  # foregroundColor: "#e4e4e4"
  # borderColor: "#95CEDA"
  # borderFocusColor: "#FF9F77"
  # titleColor: "#EBEBEB"
  # selectionBackground: "#1B4248"
  # treeSelectionBackground: "#7AA2F7"
  # selectedBackground: "#1B4248"
  # selectedForeground: "#e4e4e4"
  # activeTabColor: "#FF9F77"
  # buttonBackgroundColor: "#1B4248"
  # buttonSelectedColor: "#FFD700"
  # dropdownFocusedBackground: "#636DA6"
  # inputBackgroundColor: "#153035"
  # labelColor: "#95CEDA"
  # valueColor: "#e4e4e4"

ui:
  collectionExpansion: "closed"       # "closed", "expanded", or "remember"
  collectionIcon: ""
  collectionExpandedIcon: ""
  selectedRequestIcon: ""
  selectedRequestIconColor: "#c8d3f5"
  headerRemoveIcon: ""
  multipartRemoveIcon: ""
  configButtonIcon: "⚙"
  dropdownIndicator: "▼"
  buttonFlashDuration: 100
  fileBrowserFolderIcon: "📁"
  fileBrowserFolderExpandedIcon: "📂"
  fileBrowserFileIcon: "📄"
  copyResponseIcon: "📋"
  exportResponseIcon: "💾"
  checkboxOn: "●"
  checkboxOff: "○"

# Syntax highlighting theme (chroma themes)
# Popular options: github-dark, dracula, monokai, solarized-dark, nord, one-dark, vim, github
# Run 'petitorium themes' to see all available themes
syntaxTheme: "tokyonight-night"

# Selected environment name (defaults to "Base")
selectedEnvironment: "Base"

# HTTP request timeout in seconds (default: 60)
requestTimeout: 60

# Maximum number of response history entries per request (default: 25, 0 = no limit)
maxResponseHistory: 25

# Disable latest version check
disableVersionCheck: false

plugins:
  registry_url: "http://localhost:8080"
  enabled: []
  installed: {}
  config: {}

methodColors:
  GET: "#6EA5A0"
  POST: "#FF00FF"
  PUT: "#FF9F77"
  PATCH: "#FF9F77"
  DELETE: "#FB4F49"
  OPTIONS: "#FFA500"
  HEAD: "#800080"
  default: "#888888"

statusColors:
  success: "#28a745"          # 2xx background - Green
  successText: "#ffffff"      # 2xx text - White
  redirection: "#fd7e14"      # 3xx background - Orange
  redirectionText: "#ffffff"  # 3xx text - White
  clientError: "#dc3545"      # 4xx background - Red
  clientErrorText: "#ffffff"  # 4xx text - White
  serverError: "#8b0000"      # 5xx background - Dark Red
  serverErrorText: "#ffffff"  # 5xx text - White
  default: "#636DA6"          # Unknown background - Default theme color
  defaultText: "#ffffff"      # Unknown text - White

shortcuts:
  jumpToWorkspace: "ctrl+w"
  jumpToEnvironment: "ctrl+e"
  jumpToCollections: "ctrl+r"
  jumpToURLBar: "ctrl+u"
  jumpToRequest: "ctrl+b"
  jumpToResponse: "ctrl+s"
  sendRequest: "ctrl+j"
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

	if err := viper.Unmarshal(&C); err != nil {
		return err
	}

	// FORCE READ FROM VIPER if unmarshal failed for this specific field
	if C.Plugins.RegistryURL == "" || C.Plugins.RegistryURL == "https://api.petitorium.dev" {
		C.Plugins.RegistryURL = viper.GetString("plugins.registry_url")
	}

	// Last resort fallback
	if C.Plugins.RegistryURL == "" {
		C.Plugins.RegistryURL = "http://localhost:8080/api/v1"
	}

	if C.UI.CheckboxOn == "" {
		C.UI.CheckboxOn = "●"
	}
	if C.UI.CheckboxOff == "" {
		C.UI.CheckboxOff = "○"
	}

	if C.Shortcuts.JumpToWorkspace == "" {
		C.Shortcuts.JumpToWorkspace = "ctrl+w"
	}
	if C.Shortcuts.JumpToEnvironment == "" {
		C.Shortcuts.JumpToEnvironment = "ctrl+e"
	}
	if C.Shortcuts.JumpToCollections == "" {
		C.Shortcuts.JumpToCollections = "ctrl+r"
	}
	if C.Shortcuts.JumpToURLBar == "" {
		C.Shortcuts.JumpToURLBar = "ctrl+u"
	}
	if C.Shortcuts.JumpToRequest == "" {
		C.Shortcuts.JumpToRequest = "ctrl+b"
	}
	if C.Shortcuts.JumpToResponse == "" {
		C.Shortcuts.JumpToResponse = "ctrl+s"
	}
	if C.Shortcuts.SendRequest == "" {
		C.Shortcuts.SendRequest = "ctrl+j"
	}

	return nil
}

func SaveConfig(config *AppConfig) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "petitorium")
	configFile := filepath.Join(configDir, "config.yaml")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	data = fixNerdFontEscapes(data)

	return os.WriteFile(configFile, data, 0644)
}

func fixNerdFontEscapes(data []byte) []byte {
	replacements := map[string]string{
		"\"\\U000F024B\"": "\"\uf024b\"",
		"\"\\U000F024C\"": "\"\uf024c\"",
		"\"\\U000F0370\"": "\"\uf0370\"",
		"\"\\U000F02D7\"": "\"\uf02d7\"",
		"\"\\U000F02DC\"": "\"\uf02dc\"",
		"\"\\U000F02E3\"": "\"\uf02e3\"",
		"\"\\U000F031D\"": "\"\uf031d\"",
		"\"\\U000F031F\"": "\"\uf031f\"",
		"\"\\U000F0321\"": "\"\uf0321\"",
		"\"\\U000F0323\"": "\"\uf0323\"",
		"\"\\U000F0325\"": "\"\uf0325\"",
		"\"\\U000F0337\"": "\"\uf0337\"",
		"\"\\U000F0447\"": "\"\uf0447\"",
		"\"\\U000F04AB\"": "\"\uf04ab\"",
		"\"\\U000F04B9\"": "\"\uf04b9\"",
		"\"\\U000F04BB\"": "\"\uf04bb\"",
		"\"\\U000F0513\"": "\"\uf0513\"",
		"\"\\U000F0531\"": "\"\uf0531\"",
		"\"\\U000F054F\"": "\"\uf054f\"",
		"\"\\U000F0555\"": "\"\uf0555\"",
		"\"\\U000F0557\"": "\"\uf0557\"",
		"\"\\U000F0589\"": "\"\uf0589\"",
		"\"\\U000F058A\"": "\"\uf058a\"",
		"\"\\U000F0593\"": "\"\uf0593\"",
		"\"\\U000F0594\"": "\"\uf0594\"",
		"\"\\U000F05A4\"": "\"\uf05a4\"",
		"\"\\U000F05A5\"": "\"\uf05a5\"",
		"\"\\U000F05A6\"": "\"\uf05a6\"",
		"\"\\U000F05A7\"": "\"\uf05a7\"",
		"\"\\U000F05B7\"": "\"\uf05b7\"",
		"\"\\U000F05B8\"": "\"\uf05b8\"",
		"\"\\U000F05BA\"": "\"\uf05ba\"",
		"\"\\U000F05BB\"": "\"\uf05bb\"",
		"\"\\U000F05BC\"": "\"\uf05bc\"",
		"\"\\U000F05BD\"": "\"\uf05bd\"",
		"\"\\U000F05BE\"": "\"\uf05be\"",
		"\"\\U000F05BF\"": "\"\uf05bf\"",
		"\"\\U000F05C0\"": "\"\uf05c0\"",
		"\"\\U000F05C1\"": "\"\uf05c1\"",
		"\"\\U000F05C2\"": "\"\uf05c2\"",
		"\"\\U000F05C3\"": "\"\uf05c3\"",
		"\"\\U000F05C4\"": "\"\uf05c4\"",
		"\"\\U000F05C5\"": "\"\uf05c5\"",
		"\"\\U000F05C6\"": "\"\uf05c6\"",
		"\"\\U000F05C7\"": "\"\uf05c7\"",
		"\"\\U000F05C8\"": "\"\uf05c8\"",
		"\"\\U000F05C9\"": "\"\uf05c9\"",
		"\"\\U000F05CA\"": "\"\uf05ca\"",
		"\"\\U000F05CB\"": "\"\uf05cb\"",
		"\"\\U000F05CC\"": "\"\uf05cc\"",
		"\"\\U000F05CD\"": "\"\uf05cd\"",
		"\"\\U000F05CE\"": "\"\uf05ce\"",
		"\"\\U000F05CF\"": "\"\uf05cf\"",
		"\"\\U000F05D0\"": "\"\uf05d0\"",
		"\"\\U000F05D1\"": "\"\uf05d1\"",
		"\"\\U000F05D2\"": "\"\uf05d2\"",
		"\"\\U000F05D3\"": "\"\uf05d3\"",
		"\"\\U000F05D4\"": "\"\uf05d4\"",
		"\"\\U000F06E1\"": "\"\uf06e1\"",
		"\"\\U000F076D\"": "\"\uf076d\"",
		"\"\\U000F07C9\"": "\"\uf07c9\"",
		"\"\\U000F0800\"": "\"\uf0800\"",
		"\"\\U000F0801\"": "\"\uf0801\"",
		"\"\\U000F0802\"": "\"\uf0802\"",
		"\"\\U000F0803\"": "\"\uf0803\"",
		"\"\\U000F0804\"": "\"\uf0804\"",
		"\"\\U000F0805\"": "\"\uf0805\"",
		"\"\\U000F0806\"": "\"\uf0806\"",
		"\"\\U000F0807\"": "\"\uf0807\"",
		"\"\\U000F0808\"": "\"\uf0808\"",
		"\"\\U000F0809\"": "\"\uf0809\"",
		"\"\\U000F080A\"": "\"\uf080a\"",
		"\"\\U000F080B\"": "\"\uf080b\"",
		"\"\\U000F080C\"": "\"\uf080c\"",
		"\"\\U000F080D\"": "\"\uf080d\"",
		"\"\\U000F080E\"": "\"\uf080e\"",
		"\"\\U000F080F\"": "\"\uf080f\"",
		"\"\\U000F0810\"": "\"\uf0810\"",
		"\"\\U000F0811\"": "\"\uf0811\"",
		"\"\\U000F0812\"": "\"\uf0812\"",
		"\"\\U000F0813\"": "\"\uf0813\"",
		"\"\\U000F0814\"": "\"\uf0814\"",
		"\"\\U000F0815\"": "\"\uf0815\"",
		"\"\\U000F0816\"": "\"\uf0816\"",
		"\"\\U000F0817\"": "\"\uf0817\"",
		"\"\\U000F0818\"": "\"\uf0818\"",
		"\"\\U000F0819\"": "\"\uf0819\"",
		"\"\\U000F081A\"": "\"\uf081a\"",
		"\"\\U000F081B\"": "\"\uf081b\"",
		"\"\\U000F081C\"": "\"\uf081c\"",
		"\"\\U000F081D\"": "\"\uf081d\"",
		"\"\\U000F081E\"": "\"\uf081e\"",
		"\"\\U000F081F\"": "\"\uf081f\"",
		"\"\\U000F0820\"": "\"\uf0820\"",
		"\"\\U000F0821\"": "\"\uf0821\"",
		"\"\\U000F0822\"": "\"\uf0822\"",
		"\"\\U000F0823\"": "\"\uf0823\"",
		"\"\\U000F0824\"": "\"\uf0824\"",
		"\"\\U000F0825\"": "\"\uf0825\"",
		"\"\\U000F0826\"": "\"\uf0826\"",
		"\"\\U000F0827\"": "\"\uf0827\"",
		"\"\\U000F0828\"": "\"\uf0828\"",
		"\"\\U000F0829\"": "\"\uf0829\"",
		"\"\\U000F082A\"": "\"\uf082a\"",
		"\"\\U000F082B\"": "\"\uf082b\"",
		"\"\\U000F082C\"": "\"\uf082c\"",
		"\"\\U000F082D\"": "\"\uf082d\"",
		"\"\\U000F082E\"": "\"\uf082e\"",
		"\"\\U000F082F\"": "\"\uf082f\"",
		"\"\\U000F0830\"": "\"\uf0830\"",
		"\"\\U000F0831\"": "\"\uf0831\"",
		"\"\\U000F0832\"": "\"\uf0832\"",
		"\"\\U000F0833\"": "\"\uf0833\"",
		"\"\\U000F0834\"": "\"\uf0834\"",
		"\"\\U000F0835\"": "\"\uf0835\"",
		"\"\\U000F0836\"": "\"\uf0836\"",
		"\"\\U000F0837\"": "\"\uf0837\"",
		"\"\\U000F0838\"": "\"\uf0838\"",
		"\"\\U000F0839\"": "\"\uf0839\"",
		"\"\\U000F083A\"": "\"\uf083a\"",
		"\"\\U000F083B\"": "\"\uf083b\"",
		"\"\\U000F083C\"": "\"\uf083c\"",
		"\"\\U000F083D\"": "\"\uf083d\"",
		"\"\\U000F083E\"": "\"\uf083e\"",
		"\"\\U000F083F\"": "\"\uf083f\"",
		"\"\\U000F0840\"": "\"\uf0840\"",
		"\"\\U000F0841\"": "\"\uf0841\"",
		"\"\\U000F0842\"": "\"\uf0842\"",
		"\"\\U000F0843\"": "\"\uf0843\"",
		"\"\\U000F0844\"": "\"\uf0844\"",
		"\"\\U000F0845\"": "\"\uf0845\"",
		"\"\\U000F0846\"": "\"\uf0846\"",
		"\"\\U000F0847\"": "\"\uf0847\"",
		"\"\\U000F0848\"": "\"\uf0848\"",
		"\"\\U000F0849\"": "\"\uf0849\"",
		"\"\\U000F084A\"": "\"\uf084a\"",
		"\"\\U000F084B\"": "\"\uf084b\"",
		"\"\\U000F084C\"": "\"\uf084c\"",
		"\"\\U000F084D\"": "\"\uf084d\"",
		"\"\\U000F084E\"": "\"\uf084e\"",
		"\"\\U000F084F\"": "\"\uf084f\"",
		"\"\\U000F0850\"": "\"\uf0850\"",
		"\"\\U000F0851\"": "\"\uf0851\"",
		"\"\\U000F0852\"": "\"\uf0852\"",
		"\"\\U000F0853\"": "\"\uf0853\"",
		"\"\\U000F0854\"": "\"\uf0854\"",
		"\"\\U000F0855\"": "\"\uf0855\"",
		"\"\\U000F0856\"": "\"\uf0856\"",
		"\"\\U000F0857\"": "\"\uf0857\"",
		"\"\\U000F0858\"": "\"\uf0858\"",
		"\"\\U000F0860\"": "\"\uf0860\"",
		"\"\\U000F0861\"": "\"\uf0861\"",
		"\"\\U000F0862\"": "\"\uf0862\"",
		"\"\\U000F0863\"": "\"\uf0863\"",
		"\"\\U000F0864\"": "\"\uf0864\"",
		"\"\\U000F0865\"": "\"\uf0865\"",
		"\"\\U000F0866\"": "\"\uf0866\"",
		"\"\\U000F0867\"": "\"\uf0867\"",
		"\"\\U000F0868\"": "\"\uf0868\"",
		"\"\\U000F0869\"": "\"\uf0869\"",
		"\"\\U000F086A\"": "\"\uf086a\"",
		"\"\\U000F086B\"": "\"\uf086b\"",
		"\"\\U000F086C\"": "\"\uf086c\"",
		"\"\\U000F086D\"": "\"\uf086d\"",
		"\"\\U000F086E\"": "\"\uf086e\"",
		"\"\\U000F086F\"": "\"\uf086f\"",
		"\"\\U000F0870\"": "\"\uf0870\"",
		"\"\\U000F0871\"": "\"\uf0871\"",
		"\"\\U000F0872\"": "\"\uf0872\"",
		"\"\\U000F0873\"": "\"\uf0873\"",
		"\"\\U000F0874\"": "\"\uf0874\"",
		"\"\\U000F0875\"": "\"\uf0875\"",
		"\"\\U000F0876\"": "\"\uf0876\"",
		"\"\\U000F0877\"": "\"\uf0877\"",
		"\"\\U000F0878\"": "\"\uf0878\"",
		"\"\\U000F0879\"": "\"\uf0879\"",
		"\"\\U000F087A\"": "\"\uf087a\"",
		"\"\\U000F087B\"": "\"\uf087b\"",
		"\"\\U000F087C\"": "\"\uf087c\"",
		"\"\\U000F087D\"": "\"\uf087d\"",
		"\"\\U000F087E\"": "\"\uf087e\"",
		"\"\\U000F087F\"": "\"\uf087f\"",
		"\"\\U000F0880\"": "\"\uf0880\"",
		"\"\\U000F0881\"": "\"\uf0881\"",
		"\"\\U000F0882\"": "\"\uf0882\"",
		"\"\\U000F0883\"": "\"\uf0883\"",
		"\"\\U000F0884\"": "\"\uf0884\"",
		"\"\\U000F0885\"": "\"\uf0885\"",
		"\"\\U000F0886\"": "\"\uf0886\"",
		"\"\\U000F0887\"": "\"\uf0887\"",
		"\"\\U000F0888\"": "\"\uf0888\"",
		"\"\\U000F0889\"": "\"\uf0889\"",
		"\"\\U000F088A\"": "\"\uf088a\"",
		"\"\\U000F088B\"": "\"\uf088b\"",
		"\"\\U000F088C\"": "\"\uf088c\"",
		"\"\\U000F088D\"": "\"\uf088d\"",
		"\"\\U000F088E\"": "\"\uf088e\"",
		"\"\\U000F088F\"": "\"\uf088f\"",
		"\"\\U000F0890\"": "\"\uf0890\"",
		"\"\\U000F0891\"": "\"\uf0891\"",
		"\"\\U000F0892\"": "\"\uf0892\"",
		"\"\\U000F0893\"": "\"\uf0893\"",
		"\"\\U000F0894\"": "\"\uf0894\"",
		"\"\\U000F0895\"": "\"\uf0895\"",
		"\"\\U000F0896\"": "\"\uf0896\"",
		"\"\\U000F0897\"": "\"\uf0897\"",
		"\"\\U000F0898\"": "\"\uf0898\"",
		"\"\\U000F0899\"": "\"\uf0899\"",
		"\"\\U000F089A\"": "\"\uf089a\"",
		"\"\\U000F089B\"": "\"\uf089b\"",
		"\"\\U000F089C\"": "\"\uf089c\"",
		"\"\\U000F089D\"": "\"\uf089d\"",
		"\"\\U000F089E\"": "\"\uf089e\"",
		"\"\\U000F089F\"": "\"\uf089f\"",
		"\"\\U000F08A0\"": "\"\uf08a0\"",
		"\"\\U000F08A1\"": "\"\uf08a1\"",
		"\"\\U000F08A2\"": "\"\uf08a2\"",
		"\"\\U000F08A3\"": "\"\uf08a3\"",
		"\"\\U000F08A4\"": "\"\uf08a4\"",
		"\"\\U000F08A5\"": "\"\uf08a5\"",
		"\"\\U000F08A6\"": "\"\uf08a6\"",
		"\"\\U000F08A7\"": "\"\uf08a7\"",
		"\"\\U000F08A8\"": "\"\uf08a8\"",
		"\"\\U000F08A9\"": "\"\uf08a9\"",
		"\"\\U000F08AA\"": "\"\uf08aa\"",
		"\"\\U000F08AB\"": "\"\uf08ab\"",
		"\"\\U000F08AC\"": "\"\uf08ac\"",
		"\"\\U000F08AD\"": "\"\uf08ad\"",
		"\"\\U000F08AE\"": "\"\uf08ae\"",
		"\"\\U000F08AF\"": "\"\uf08af\"",
		"\"\\U000F08B0\"": "\"\uf08b0\"",
		"\"\\U000F08B1\"": "\"\uf08b1\"",
		"\"\\U000F08B2\"": "\"\uf08b2\"",
		"\"\\U000F08B3\"": "\"\uf08b3\"",
		"\"\\U000F08B4\"": "\"\uf08b4\"",
		"\"\\U000F08B5\"": "\"\uf08b5\"",
		"\"\\U000F08B6\"": "\"\uf08b6\"",
		"\"\\U000F08B7\"": "\"\uf08b7\"",
		"\"\\U000F08B8\"": "\"\uf08b8\"",
		"\"\\U000F08B9\"": "\"\uf08b9\"",
		"\"\\U000F08BA\"": "\"\uf08ba\"",
		"\"\\U000F08BB\"": "\"\uf08bb\"",
		"\"\\U000F08BC\"": "\"\uf08bc\"",
		"\"\\U000F08BD\"": "\"\uf08bd\"",
		"\"\\U000F08BE\"": "\"\uf08be\"",
		"\"\\U000F08BF\"": "\"\uf08bf\"",
		"\"\\U000F08C0\"": "\"\uf08c0\"",
		"\"\\U000F08C1\"": "\"\uf08c1\"",
		"\"\\U000F08C2\"": "\"\uf08c2\"",
		"\"\\U000F08C3\"": "\"\uf08c3\"",
		"\"\\U000F08C4\"": "\"\uf08c4\"",
		"\"\\U000F08C5\"": "\"\uf08c5\"",
		"\"\\U000F08C6\"": "\"\uf08c6\"",
		"\"\\U000F08C7\"": "\"\uf08c7\"",
		"\"\\U000F08C8\"": "\"\uf08c8\"",
		"\"\\U000F08C9\"": "\"\uf08c9\"",
		"\"\\U000F08CA\"": "\"\uf08ca\"",
		"\"\\U000F08CB\"": "\"\uf08cb\"",
		"\"\\U000F08CC\"": "\"\uf08cc\"",
		"\"\\U000F08CD\"": "\"\uf08cd\"",
		"\"\\U000F08CE\"": "\"\uf08ce\"",
		"\"\\U000F08CF\"": "\"\uf08cf\"",
		"\"\\U000F08D0\"": "\"\uf08d0\"",
		"\"\\U000F08D1\"": "\"\uf08d1\"",
		"\"\\U000F08D2\"": "\"\uf08d2\"",
		"\"\\U000F08D3\"": "\"\uf08d3\"",
		"\"\\U000F08D4\"": "\"\uf08d4\"",
		"\"\\U000F08D5\"": "\"\uf08d5\"",
		"\"\\U000F08D6\"": "\"\uf08d6\"",
		"\"\\U000F08D7\"": "\"\uf08d7\"",
		"\"\\U000F08D8\"": "\"\uf08d8\"",
		"\"\\U000F08D9\"": "\"\uf08d9\"",
		"\"\\U000F08DA\"": "\"\uf08da\"",
		"\"\\U000F08DB\"": "\"\uf08db\"",
		"\"\\U000F08DC\"": "\"\uf08dc\"",
		"\"\\U000F08DD\"": "\"\uf08dd\"",
		"\"\\U000F08DE\"": "\"\uf08de\"",
		"\"\\U000F08DF\"": "\"\uf08df\"",
		"\"\\U000F08E0\"": "\"\uf08e0\"",
		"\"\\U000F08E1\"": "\"\uf08e1\"",
		"\"\\U000F08E2\"": "\"\uf08e2\"",
		"\"\\U000F08E3\"": "\"\uf08e3\"",
		"\"\\U000F08E4\"": "\"\uf08e4\"",
		"\"\\U000F08E5\"": "\"\uf08e5\"",
		"\"\\U000F08E6\"": "\"\uf08e6\"",
		"\"\\U000F08E7\"": "\"\uf08e7\"",
		"\"\\U000F08E8\"": "\"\uf08e8\"",
		"\"\\U000F08E9\"": "\"\uf08e9\"",
		"\"\\U000F08EA\"": "\"\uf08ea\"",
		"\"\\U000F08EB\"": "\"\uf08eb\"",
		"\"\\U000F08EC\"": "\"\uf08ec\"",
		"\"\\U000F08ED\"": "\"\uf08ed\"",
		"\"\\U000F08EE\"": "\"\uf08ee\"",
		"\"\\U000F08EF\"": "\"\uf08ef\"",
		"\"\\U000F08F0\"": "\"\uf08f0\"",
		"\"\\U000F08F1\"": "\"\uf08f1\"",
		"\"\\U000F08F2\"": "\"\uf08f2\"",
		"\"\\U000F08F3\"": "\"\uf08f3\"",
		"\"\\U000F08F4\"": "\"\uf08f4\"",
		"\"\\U000F08F5\"": "\"\uf08f5\"",
		"\"\\U000F08F6\"": "\"\uf08f6\"",
		"\"\\U000F08F7\"": "\"\uf08f7\"",
		"\"\\U000F08F8\"": "\"\uf08f8\"",
		"\"\\U000F08F9\"": "\"\uf08f9\"",
		"\"\\U000F08FA\"": "\"\uf08fa\"",
		"\"\\U000F08FB\"": "\"\uf08fb\"",
		"\"\\U000F08FC\"": "\"\uf08fc\"",
		"\"\\U000F08FD\"": "\"\uf08fd\"",
		"\"\\U000F08FE\"": "\"\uf08fe\"",
		"\"\\U000F08FF\"": "\"\uf08ff\"",
		"\"\\U000F0B4B\"": "\"\uf0b4b\"",
		"\"\\U000F0B53\"": "\"\uf0b53\"",
		"\"\\U000F0B54\"": "\"\uf0b54\"",
		"\"\\U000F0B55\"": "\"\uf0b55\"",
		"\"\\U000F0B67\"": "\"\uf0b67\"",
		"\"\\U000F0B70\"": "\"\uf0b70\"",
		"\"\\U000F0B71\"": "\"\uf0b71\"",
		"\"\\U000F0B72\"": "\"\uf0b72\"",
		"\"\\U000F0C00\"": "\"\uf0c00\"",
		"\"\\U000F0C25\"": "\"\uf0c25\"",
		"\"\\U000F0C26\"": "\"\uf0c26\"",
		"\"\\U000F0C27\"": "\"\uf0c27\"",
		"\"\\U000F0C28\"": "\"\uf0c28\"",
		"\"\\U000F0C29\"": "\"\uf0c29\"",
		"\"\\U000F0C2A\"": "\"\uf0c2a\"",
		"\"\\U000F0C2B\"": "\"\uf0c2b\"",
		"\"\\U000F0C2C\"": "\"\uf0c2c\"",
		"\"\\U000F0C2D\"": "\"\uf0c2d\"",
		"\"\\U000F0C2E\"": "\"\uf0c2e\"",
		"\"\\U000F0C2F\"": "\"\uf0c2f\"",
		"\"\\U000F0C30\"": "\"\uf0c30\"",
		"\"\\U000F0C31\"": "\"\uf0c31\"",
		"\"\\U000F0C32\"": "\"\uf0c32\"",
		"\"\\U000F0C33\"": "\"\uf0c33\"",
		"\"\\U000F0C34\"": "\"\uf0c34\"",
		"\"\\U000F0C35\"": "\"\uf0c35\"",
		"\"\\U000F0C36\"": "\"\uf0c36\"",
		"\"\\U000F0C37\"": "\"\uf0c37\"",
		"\"\\U000F0C38\"": "\"\uf0c38\"",
		"\"\\U000F0C39\"": "\"\uf0c39\"",
		"\"\\U000F0C3A\"": "\"\uf0c3a\"",
		"\"\\U000F0C3B\"": "\"\uf0c3b\"",
		"\"\\U000F0C3C\"": "\"\uf0c3c\"",
		"\"\\U000F0C3D\"": "\"\uf0c3d\"",
		"\"\\U000F0C3E\"": "\"\uf0c3e\"",
		"\"\\U000F0C3F\"": "\"\uf0c3f\"",
		"\"\\U000F0C40\"": "\"\uf0c40\"",
		"\"\\U000F0C41\"": "\"\uf0c41\"",
		"\"\\U000F0C42\"": "\"\uf0c42\"",
		"\"\\U000F0C43\"": "\"\uf0c43\"",
		"\"\\U000F0C44\"": "\"\uf0c44\"",
		"\"\\U000F0C45\"": "\"\uf0c45\"",
		"\"\\U000F0C46\"": "\"\uf0c46\"",
		"\"\\U000F0C47\"": "\"\uf0c47\"",
		"\"\\U000F0C48\"": "\"\uf0c48\"",
		"\"\\U000F0C49\"": "\"\uf0c49\"",
		"\"\\U000F0C4A\"": "\"\uf0c4a\"",
		"\"\\U000F0C4B\"": "\"\uf0c4b\"",
		"\"\\U000F0C4C\"": "\"\uf0c4c\"",
		"\"\\U000F0C4D\"": "\"\uf0c4d\"",
		"\"\\U000F0C4E\"": "\"\uf0c4e\"",
		"\"\\U000F0C4F\"": "\"\uf0c4f\"",
		"\"\\U000F0C50\"": "\"\uf0c50\"",
		"\"\\U000F0C51\"": "\"\uf0c51\"",
		"\"\\U000F0C52\"": "\"\uf0c52\"",
		"\"\\U000F0C53\"": "\"\uf0c53\"",
		"\"\\U000F0C54\"": "\"\uf0c54\"",
		"\"\\U000F0C55\"": "\"\uf0c55\"",
		"\"\\U000F0C56\"": "\"\uf0c56\"",
		"\"\\U000F0C57\"": "\"\uf0c57\"",
		"\"\\U000F0C58\"": "\"\uf0c58\"",
		"\"\\U000F0C59\"": "\"\uf0c59\"",
		"\"\\U000F0C5A\"": "\"\uf0c5a\"",
		"\"\\U000F0C5B\"": "\"\uf0c5b\"",
		"\"\\U000F0C5C\"": "\"\uf0c5c\"",
		"\"\\U000F0C5D\"": "\"\uf0c5d\"",
		"\"\\U000F0C5E\"": "\"\uf0c5e\"",
		"\"\\U000F0C5F\"": "\"\uf0c5f\"",
		"\"\\U000F0C60\"": "\"\uf0c60\"",
		"\"\\U000F0C61\"": "\"\uf0c61\"",
		"\"\\U000F0C62\"": "\"\uf0c62\"",
		"\"\\U000F0C63\"": "\"\uf0c63\"",
		"\"\\U000F0C64\"": "\"\uf0c64\"",
		"\"\\U000F0C65\"": "\"\uf0c65\"",
		"\"\\U000F0C66\"": "\"\uf0c66\"",
		"\"\\U000F0C67\"": "\"\uf0c67\"",
		"\"\\U000F0C68\"": "\"\uf0c68\"",
		"\"\\U000F0C69\"": "\"\uf0c69\"",
		"\"\\U000F0C6A\"": "\"\uf0c6a\"",
		"\"\\U000F0C6B\"": "\"\uf0c6b\"",
		"\"\\U000F0C6C\"": "\"\uf0c6c\"",
		"\"\\U000F0C6D\"": "\"\uf0c6d\"",
		"\"\\U000F0C6E\"": "\"\uf0c6e\"",
		"\"\\U000F0C6F\"": "\"\uf0c6f\"",
		"\"\\U000F0C70\"": "\"\uf0c70\"",
		"\"\\U000F0C71\"": "\"\uf0c71\"",
		"\"\\U000F0C72\"": "\"\uf0c72\"",
		"\"\\U000F0C73\"": "\"\uf0c73\"",
		"\"\\U000F0C74\"": "\"\uf0c74\"",
		"\"\\U000F0C75\"": "\"\uf0c75\"",
		"\"\\U000F0C76\"": "\"\uf0c76\"",
		"\"\\U000F0C77\"": "\"\uf0c77\"",
		"\"\\U000F0C78\"": "\"\uf0c78\"",
		"\"\\U000F0C79\"": "\"\uf0c79\"",
		"\"\\U000F0C7A\"": "\"\uf0c7a\"",
		"\"\\U000F0C7B\"": "\"\uf0c7b\"",
		"\"\\U000F0C7C\"": "\"\uf0c7c\"",
		"\"\\U000F0C7D\"": "\"\uf0c7d\"",
		"\"\\U000F0C7E\"": "\"\uf0c7e\"",
		"\"\\U000F0C7F\"": "\"\uf0c7f\"",
		"\"\\U000F0C80\"": "\"\uf0c80\"",
		"\"\\U000F0C81\"": "\"\uf0c81\"",
		"\"\\U000F0C82\"": "\"\uf0c82\"",
		"\"\\U000F0C83\"": "\"\uf0c83\"",
		"\"\\U000F0C84\"": "\"\uf0c84\"",
		"\"\\U000F0C85\"": "\"\uf0c85\"",
		"\"\\U000F0C86\"": "\"\uf0c86\"",
		"\"\\U000F0C87\"": "\"\uf0c87\"",
		"\"\\U000F0C88\"": "\"\uf0c88\"",
		"\"\\U000F0C89\"": "\"\uf0c89\"",
		"\"\\U000F0C8A\"": "\"\uf0c8a\"",
		"\"\\U000F0C8B\"": "\"\uf0c8b\"",
		"\"\\U000F0C8C\"": "\"\uf0c8c\"",
		"\"\\U000F0C8D\"": "\"\uf0c8d\"",
		"\"\\U000F0C8E\"": "\"\uf0c8e\"",
		"\"\\U000F0C8F\"": "\"\uf0c8f\"",
		"\"\\U000F0C90\"": "\"\uf0c90\"",
		"\"\\U000F0C91\"": "\"\uf0c91\"",
		"\"\\U000F0C92\"": "\"\uf0c92\"",
		"\"\\U000F0C93\"": "\"\uf0c93\"",
		"\"\\U000F0C94\"": "\"\uf0c94\"",
		"\"\\U000F0C95\"": "\"\uf0c95\"",
		"\"\\U000F0C96\"": "\"\uf0c96\"",
		"\"\\U000F0C97\"": "\"\uf0c97\"",
		"\"\\U000F0C98\"": "\"\uf0c98\"",
		"\"\\U000F0C99\"": "\"\uf0c99\"",
		"\"\\U000F0C9A\"": "\"\uf0c9a\"",
		"\"\\U000F0C9B\"": "\"\uf0c9b\"",
		"\"\\U000F0C9C\"": "\"\uf0c9c\"",
		"\"\\U000F0C9D\"": "\"\uf0c9d\"",
		"\"\\U000F0C9E\"": "\"\uf0c9e\"",
		"\"\\U000F0C9F\"": "\"\uf0c9f\"",
		"\"\\U000F0CA0\"": "\"\uf0ca0\"",
		"\"\\U000F0CA1\"": "\"\uf0ca1\"",
		"\"\\U000F0CA2\"": "\"\uf0ca2\"",
		"\"\\U000F0CA3\"": "\"\uf0ca3\"",
		"\"\\U000F0CA4\"": "\"\uf0ca4\"",
		"\"\\U000F0CA5\"": "\"\uf0ca5\"",
		"\"\\U000F0CA6\"": "\"\uf0ca6\"",
		"\"\\U000F0CA7\"": "\"\uf0ca7\"",
		"\"\\U000F0CA8\"": "\"\uf0ca8\"",
		"\"\\U000F0CA9\"": "\"\uf0ca9\"",
		"\"\\U000F0CAA\"": "\"\uf0caa\"",
		"\"\\U000F0CAB\"": "\"\uf0cab\"",
		"\"\\U000F0CAC\"": "\"\uf0cac\"",
		"\"\\U000F0CAD\"": "\"\uf0cad\"",
		"\"\\U000F0CAE\"": "\"\uf0cae\"",
		"\"\\U000F0CAF\"": "\"\uf0caf\"",
		"\"\\U000F0CB0\"": "\"\uf0cb0\"",
		"\"\\U000F0CB1\"": "\"\uf0cb1\"",
		"\"\\U000F0CB2\"": "\"\uf0cb2\"",
		"\"\\U000F0CB3\"": "\"\uf0cb3\"",
		"\"\\U000F0CB4\"": "\"\uf0cb4\"",
		"\"\\U000F0CB5\"": "\"\uf0cb5\"",
		"\"\\U000F0CB6\"": "\"\uf0cb6\"",
		"\"\\U000F0CB7\"": "\"\uf0cb7\"",
		"\"\\U000F0CB8\"": "\"\uf0cb8\"",
		"\"\\U000F0CB9\"": "\"\uf0cb9\"",
		"\"\\U000F0CBA\"": "\"\uf0cba\"",
		"\"\\U000F0CBB\"": "\"\uf0cbb\"",
		"\"\\U000F0CBC\"": "\"\uf0cbc\"",
		"\"\\U000F0CBD\"": "\"\uf0cbd\"",
		"\"\\U000F0CBE\"": "\"\uf0cbe\"",
		"\"\\U000F0CBF\"": "\"\uf0cbf\"",
		"\"\\U000F0CC0\"": "\"\uf0cc0\"",
		"\"\\U000F0CC1\"": "\"\uf0cc1\"",
		"\"\\U000F0CC2\"": "\"\uf0cc2\"",
		"\"\\U000F0CC3\"": "\"\uf0cc3\"",
		"\"\\U000F0CC4\"": "\"\uf0cc4\"",
		"\"\\U000F0CC5\"": "\"\uf0cc5\"",
		"\"\\U000F0CC6\"": "\"\uf0cc6\"",
		"\"\\U000F0CC7\"": "\"\uf0cc7\"",
		"\"\\U000F0CC8\"": "\"\uf0cc8\"",
		"\"\\U000F0CC9\"": "\"\uf0cc9\"",
		"\"\\U000F0CCA\"": "\"\uf0cca\"",
		"\"\\U000F0CCB\"": "\"\uf0ccb\"",
		"\"\\U000F0CCC\"": "\"\uf0ccc\"",
		"\"\\U000F0CCD\"": "\"\uf0ccd\"",
		"\"\\U000F0CCE\"": "\"\uf0cce\"",
		"\"\\U000F0CCF\"": "\"\uf0ccf\"",
		"\"\\U000F0CD0\"": "\"\uf0cd0\"",
		"\"\\U000F0CD1\"": "\"\uf0cd1\"",
		"\"\\U000F0CD2\"": "\"\uf0cd2\"",
		"\"\\U000F0CD3\"": "\"\uf0cd3\"",
		"\"\\U000F0CD4\"": "\"\uf0cd4\"",
		"\"\\U000F0CD5\"": "\"\uf0cd5\"",
		"\"\\U000F0CD6\"": "\"\uf0cd6\"",
		"\"\\U000F0CD7\"": "\"\uf0cd7\"",
		"\"\\U000F0CD8\"": "\"\uf0cd8\"",
		"\"\\U000F0CD9\"": "\"\uf0cd9\"",
		"\"\\U000F0CDA\"": "\"\uf0cda\"",
		"\"\\U000F0CDB\"": "\"\uf0cdb\"",
		"\"\\U000F0CDC\"": "\"\uf0cdc\"",
		"\"\\U000F0CDD\"": "\"\uf0cdd\"",
		"\"\\U000F0CDE\"": "\"\uf0cde\"",
		"\"\\U000F0CDF\"": "\"\uf0cdf\"",
		"\"\\U000F0CE0\"": "\"\uf0ce0\"",
		"\"\\U000F0CE1\"": "\"\uf0ce1\"",
		"\"\\U000F0CE2\"": "\"\uf0ce2\"",
		"\"\\U000F0CE3\"": "\"\uf0ce3\"",
		"\"\\U000F0CE4\"": "\"\uf0ce4\"",
		"\"\\U000F0CE5\"": "\"\uf0ce5\"",
		"\"\\U000F0CE6\"": "\"\uf0ce6\"",
		"\"\\U000F0CE7\"": "\"\uf0ce7\"",
		"\"\\U000F0CE8\"": "\"\uf0ce8\"",
		"\"\\U000F0CE9\"": "\"\uf0ce9\"",
		"\"\\U000F0CEA\"": "\"\uf0cea\"",
		"\"\\U000F0CEB\"": "\"\uf0ceb\"",
		"\"\\U000F0CEC\"": "\"\uf0cec\"",
		"\"\\U000F0CED\"": "\"\uf0ced\"",
		"\"\\U000F0CEE\"": "\"\uf0cee\"",
		"\"\\U000F0CEF\"": "\"\uf0cef\"",
		"\"\\U000F0CF0\"": "\"\uf0cf0\"",
		"\"\\U000F0CF1\"": "\"\uf0cf1\"",
		"\"\\U000F0CF2\"": "\"\uf0cf2\"",
		"\"\\U000F0CF3\"": "\"\uf0cf3\"",
		"\"\\U000F0CF4\"": "\"\uf0cf4\"",
		"\"\\U000F0CF5\"": "\"\uf0cf5\"",
		"\"\\U000F0CF6\"": "\"\uf0cf6\"",
		"\"\\U000F0CF7\"": "\"\uf0cf7\"",
		"\"\\U000F0CF8\"": "\"\uf0cf8\"",
		"\"\\U000F0CF9\"": "\"\uf0cf9\"",
		"\"\\U000F0CFA\"": "\"\uf0cfa\"",
		"\"\\U000F0CFB\"": "\"\uf0cfb\"",
		"\"\\U000F0CFC\"": "\"\uf0cfc\"",
		"\"\\U000F0CFD\"": "\"\uf0cfd\"",
		"\"\\U000F0CFE\"": "\"\uf0cfe\"",
		"\"\\U000F0CFF\"": "\"\uf0cff\"",
	}

	result := string(data)
	for escaped, literal := range replacements {
		result = strings.ReplaceAll(result, escaped, literal)
	}
	return []byte(result)
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
