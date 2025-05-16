package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"markdown-to-pdf/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cfg           *config.Config
)

type item struct {
	filename string
	modTime  time.Time
	path     string
}

type model struct {
	items      []item
	selected   int
	converting bool
	err        error
}

func initialModel() model {
	items := []item{}
	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			items = append(items, item{
				filename: info.Name(),
				modTime:  info.ModTime(),
				path:     path,
			})
		}
		return nil
	})
	return model{
		items: items,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			if len(m.items) > 0 && !m.converting {
				m.converting = true
				return m, m.convertToPDF
			}
		}
	case errorMsg:
		m.err = msg
		m.converting = false
		return m, nil
	case successMsg:
		m.converting = false
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit", m.err)
	}
	if m.converting {
		return fmt.Sprintf("Converting %s to PDF...\nPress q to quit", m.items[m.selected].filename)
	}
	if len(m.items) == 0 {
		return "No markdown files found. Press q to quit."
	}
	head := headerStyle.Render(fmt.Sprintf("%-30s | %-20s", "Filename", "Last Updated At"))
	sep := strings.Repeat("-", 55)
	rows := []string{head, sep}
	for i, it := range m.items {
		row := fmt.Sprintf("%-30s | %-20s", it.filename, it.modTime.Format("2006-01-02 15:04:05"))
		if i == m.selected {
			row = selectedStyle.Render("> " + row)
		} else {
			row = "  " + row
		}
		rows = append(rows, row)
	}
	rows = append(rows, "\nUse ↑/↓ to select, Enter to convert, q to quit.")
	return docStyle.Render(strings.Join(rows, "\n"))
}

type errorMsg error
type successMsg struct{}

func (m model) convertToPDF() tea.Msg {
	it := m.items[m.selected]

	// Create HTML template
	htmlTemplate := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Markdown to PDF</title>
  </head>
  <body>
    {{ toHTML "file.md" }}
  </body>
</html>`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the HTML template
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return errorMsg(err)
	}
	if _, err := part.Write([]byte(htmlTemplate)); err != nil {
		return errorMsg(err)
	}

	// Add the markdown file
	file, err := os.Open(it.path)
	if err != nil {
		return errorMsg(err)
	}
	defer file.Close()

	part, err = writer.CreateFormFile("files", "file.md")
	if err != nil {
		return errorMsg(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return errorMsg(err)
	}

	err = writer.Close()
	if err != nil {
		return errorMsg(err)
	}

	req, err := http.NewRequest("POST", cfg.Gotenberg.Endpoint+"/forms/chromium/convert/markdown", body)
	if err != nil {
		return errorMsg(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errorMsg(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errorMsg(fmt.Errorf("conversion failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes)))
	}

	outputPath := strings.TrimSuffix(it.path, filepath.Ext(it.path)) + ".pdf"
	out, err := os.Create(outputPath)
	if err != nil {
		return errorMsg(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errorMsg(err)
	}

	return successMsg{}
}

func main() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
