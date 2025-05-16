package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	AppName = "markdown-to-pdf"
)

type Config struct {
	Gotenberg struct {
		Endpoint string `toml:"endpoint"`
	} `toml:"gotenberg"`
}

func GetConfigPath() (string, error) {
	// Get XDG_CONFIG_HOME or default to ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	// Create app config directory
	appConfigDir := filepath.Join(configHome, AppName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appConfigDir, "config.toml"), nil
}

type configModel struct {
	endpoint string
	index    int
	err      error
}

func initialConfigModel() configModel {
	return configModel{
		endpoint: "http://localhost:3000",
		index:    0,
	}
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.index == 0 {
				m.index = 1
				return m, nil
			}
			return m, tea.Quit
		case "backspace":
			if m.index == 0 && len(m.endpoint) > 0 {
				m.endpoint = m.endpoint[:len(m.endpoint)-1]
			}
		default:
			if m.index == 0 && len(msg.String()) == 1 {
				m.endpoint += msg.String()
			}
		}
	}
	return m, nil
}

func (m configModel) View() string {
	s := strings.Builder{}
	s.WriteString("Welcome to Markdown to PDF Converter!\n\n")
	s.WriteString("Let's set up your configuration.\n\n")

	if m.index == 0 {
		s.WriteString("Enter Gotenberg endpoint (default: http://localhost:3000):\n")
		s.WriteString(m.endpoint)
		s.WriteString("\n\nPress Enter to continue, Ctrl+C to quit")
	} else {
		s.WriteString("Configuration saved successfully!\n")
		s.WriteString("Press Enter to start the application")
	}

	return s.String()
}

func createConfigWizard() (*Config, error) {
	p := tea.NewProgram(initialConfigModel())
	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run config wizard: %w", err)
	}

	model := m.(configModel)
	config := &Config{}
	config.Gotenberg.Endpoint = model.endpoint

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(config); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	return config, nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Run the wizard if config doesn't exist
		return createConfigWizard()
	}

	// Load existing config
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &config, nil
}
