package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Theme       string            `yaml:"theme"`
	Colors      ColorConfig       `yaml:"colors"`
	DateFormat  string            `yaml:"date_format"`
	Timezone    string            `yaml:"timezone"`
	SortBy      string            `yaml:"sort_by"`
	SortOrder   string            `yaml:"sort_order"`
	Editor      EditorConfig      `yaml:"editor"`
	Keybindings KeybindingsConfig `yaml:"keybindings"`
	Debug       bool              `yaml:"debug"`
	LogFile     string            `yaml:"log_file"`
}

type ColorConfig struct {
	Primary    string `yaml:"primary"`
	Secondary  string `yaml:"secondary"`
	Error      string `yaml:"error"`
	Success    string `yaml:"success"`
	Warning    string `yaml:"warning"`
	Muted      string `yaml:"muted"`
	Text       string `yaml:"text"`
	Background string `yaml:"background"`
	Selected   string `yaml:"selected"`
}

type EditorConfig struct {
	Mode        string `yaml:"mode"`
	TabWidth    int    `yaml:"tab_width"`
	WordWrap    bool   `yaml:"word_wrap"`
	LineNumbers bool   `yaml:"line_numbers"`
}

type KeybindingsConfig struct {
	Quit    string `yaml:"quit"`
	New     string `yaml:"new"`
	Edit    string `yaml:"edit"`
	Delete  string `yaml:"delete"`
	Search  string `yaml:"search"`
	Refresh string `yaml:"refresh"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Theme:      "dracula",
		DateFormat: "relative",
		SortBy:     "display_time",
		SortOrder:  "desc",
		Editor: EditorConfig{
			Mode:        "vim",
			TabWidth:    4,
			WordWrap:    true,
			LineNumbers: false,
		},
		Keybindings: KeybindingsConfig{
			Quit:    "q",
			New:     "n",
			Edit:    "e",
			Delete:  "d",
			Search:  "/",
			Refresh: "r",
		},
		Colors: ColorConfig{
			Primary:    "#7D56F4",
			Secondary:  "#00D9FF",
			Error:      "#FF5555",
			Success:    "#50FA7B",
			Warning:    "#FFB86C",
			Muted:      "#6272A4",
			Text:       "#F8F8F2",
			Background: "#282A36",
			Selected:   "#44475A",
		},
		Debug:   false,
		LogFile: "",
	}
}

func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "memento"), nil
}

func DataDir() (string, error) {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "memento"), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func LoadSettings() (*Settings, error) {
	settings := DefaultSettings()

	configPath, err := ConfigPath()
	if err != nil {
		return settings, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}

	if err := yaml.Unmarshal(data, settings); err != nil {
		return DefaultSettings(), err
	}

	settings.applyDefaults()

	return settings, nil
}

func SaveSettings(settings *Settings) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func CreateDefaultConfig() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	defaultConfig := `# Memento Configuration

theme: dracula

colors:
  primary: "#7D56F4"
  secondary: "#00D9FF"
  error: "#FF5555"
  success: "#50FA7B"
  warning: "#FFB86C"
  muted: "#6272A4"
  text: "#F8F8F2"
  background: "#282A36"
  selected: "#44475A"

date_format: relative

# Timezone for displaying dates (e.g., "America/New_York", "Europe/London", "Local")
timezone: Local

sort_by: display_time
sort_order: desc

editor:
  mode: vim
  tab_width: 4
  word_wrap: true
  line_numbers: false

keybindings:
  quit: "q"
  new: "n"
  edit: "e"
  delete: "d"
  search: "/"
  refresh: "r"

debug: false
`
	return os.WriteFile(configPath, []byte(defaultConfig), 0600)
}

func (s *Settings) applyDefaults() {
	defaults := DefaultSettings()

	if s.Theme == "" {
		s.Theme = defaults.Theme
	}
	if s.DateFormat == "" {
		s.DateFormat = defaults.DateFormat
	}
	if s.SortBy == "" {
		s.SortBy = defaults.SortBy
	}
	if s.SortOrder == "" {
		s.SortOrder = defaults.SortOrder
	}
	if s.Editor.Mode == "" {
		s.Editor.Mode = defaults.Editor.Mode
	}
	if s.Editor.TabWidth == 0 {
		s.Editor.TabWidth = defaults.Editor.TabWidth
	}
	if s.Keybindings.Quit == "" {
		s.Keybindings.Quit = defaults.Keybindings.Quit
	}
	if s.Keybindings.New == "" {
		s.Keybindings.New = defaults.Keybindings.New
	}
	if s.Keybindings.Edit == "" {
		s.Keybindings.Edit = defaults.Keybindings.Edit
	}
	if s.Keybindings.Delete == "" {
		s.Keybindings.Delete = defaults.Keybindings.Delete
	}
	if s.Keybindings.Search == "" {
		s.Keybindings.Search = defaults.Keybindings.Search
	}
	if s.Keybindings.Refresh == "" {
		s.Keybindings.Refresh = defaults.Keybindings.Refresh
	}

	if s.Theme == "custom" {
		if s.Colors.Primary == "" {
			s.Colors = defaults.Colors
		}
	}

	if s.Debug && s.LogFile == "" {
		if dataDir, err := DataDir(); err == nil {
			s.LogFile = filepath.Join(dataDir, "debug.log")
		}
	}
}

func (s *Settings) IsVimMode() bool {
	return s.Editor.Mode == "vim"
}

func (s *Settings) GetTimezone() *time.Location {
	if s.Timezone == "" || s.Timezone == "Local" {
		return time.Local
	}
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return time.Local
	}
	return loc
}
