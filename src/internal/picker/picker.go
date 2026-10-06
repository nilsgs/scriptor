package picker

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"scriptor/internal/discovery"
)

const visibleRows = 8

type model struct {
	scripts       []discovery.Script
	filter        textinput.Model
	cursor        int
	height        int
	width         int
	title         lipgloss.Style
	selectedStyle lipgloss.Style
	muted         lipgloss.Style
	empty         lipgloss.Style
	selected      *discovery.Script
	done          bool
}

func newModel(scripts []discovery.Script, output io.Writer) model {
	renderer := lipgloss.NewRenderer(output)
	accent := renderer.NewStyle().Foreground(lipgloss.Color("6"))
	selected := renderer.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	input := textinput.New()
	input.Prompt = "Filter: "
	input.PromptStyle = accent
	input.Cursor.Style = selected
	input.Focus()
	return model{
		scripts: scripts, filter: input, height: visibleRows, width: 80,
		title: accent.Bold(true), selectedStyle: selected,
		muted: renderer.NewStyle().Foreground(lipgloss.Color("8")),
		empty: renderer.NewStyle().Foreground(lipgloss.Color("3")),
	}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

// match implements case-insensitive subsequence filtering of the full relative path.
func match(path, query string) bool {
	remaining := []rune(strings.ToLower(query))
	if len(remaining) == 0 {
		return true
	}
	for _, r := range strings.ToLower(path) {
		if r == remaining[0] {
			remaining = remaining[1:]
			if len(remaining) == 0 {
				return true
			}
		}
	}
	return false
}

func (m model) matches() []discovery.Script {
	result := []discovery.Script{}
	for _, script := range m.scripts {
		if match(script.Path, m.filter.Value()) {
			result = append(result, script)
		}
	}
	return result
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = min(visibleRows, max(1, msg.Height-6))
		m.width = max(1, msg.Width)
		m.filter.Width = max(1, msg.Width-10)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		case "up", "ctrl+p":
			m.cursor = max(0, m.cursor-1)
			return m, nil
		case "down", "ctrl+n":
			m.cursor = min(max(0, len(m.matches())-1), m.cursor+1)
			return m, nil
		case "enter":
			matches := m.matches()
			if len(matches) > 0 {
				selected := matches[m.cursor]
				m.selected = &selected
				m.done = true
				return m, tea.Quit
			}
			return m, nil
		}
	}
	before := m.filter.Value()
	var command tea.Cmd
	m.filter, command = m.filter.Update(msg)
	if before != m.filter.Value() {
		m.cursor = 0
	}
	return m, command
}

func (m model) View() string {
	if m.done {
		return ""
	}
	matches := m.matches()
	var out strings.Builder
	fmt.Fprintf(&out, "%s %s\n\n%s\n\n", m.title.Render("scriptor"), m.muted.Render(fmt.Sprintf("%d/%d scripts", len(matches), len(m.scripts))), m.filter.View())
	if len(matches) == 0 {
		out.WriteString(m.empty.Render("  No matching scripts.") + "\n")
	}
	start := max(0, m.cursor-m.height+1)
	for i := start; i < min(len(matches), start+m.height); i++ {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		// Quote control characters so filenames cannot emit terminal escape sequences.
		path := matches[i].Path
		if strings.ContainsAny(path, "\x1b\r\n\t") {
			path = strconv.Quote(path)
		}
		line := ansi.Truncate(prefix+path, m.width, "...")
		if i == m.cursor {
			line = m.selectedStyle.Render(line)
		}
		fmt.Fprintln(&out, line)
	}
	out.WriteString("\n" + m.muted.Render("Up/Down select | Enter run | Esc cancel"))
	return out.String()
}

func Select(scripts []discovery.Script, input io.Reader, output io.Writer) (*discovery.Script, error) {
	result, err := tea.NewProgram(newModel(scripts, output), tea.WithInput(input), tea.WithOutput(output)).Run()
	if err != nil {
		return nil, err
	}
	return result.(model).selected, nil
}
