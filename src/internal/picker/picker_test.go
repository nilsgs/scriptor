package picker

import (
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"scriptor/internal/discovery"
)

func key(m model, message tea.KeyMsg) model {
	updated, _ := m.Update(message)
	return updated.(model)
}

func TestFilterSelectAndCancel(t *testing.T) {
	m := newModel([]discovery.Script{{Path: "scripts/build.ps1"}, {Path: "scripts/build.sh"}, {Path: "tools/check.py"}}, io.Discard)
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("BLD")})
	if len(m.matches()) != 2 {
		t.Fatalf("matches: %v", m.matches())
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.selected == nil || m.selected.Path != "scripts/build.sh" {
		t.Fatalf("selection: %v", m.selected)
	}
	if m.View() != "" {
		t.Fatal("picker should be cleared on exit")
	}
	for _, cancel := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m = key(newModel(m.scripts, io.Discard), tea.KeyMsg{Type: cancel})
		if !m.done || m.selected != nil {
			t.Fatal("cancel selected a script")
		}
	}
}

func TestFilterResetsCursorAndEmptyEnterDoesNothing(t *testing.T) {
	m := newModel([]discovery.Script{{Path: "a.sh"}, {Path: "b.py"}}, io.Discard)
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.cursor != 0 || len(m.matches()) != 1 {
		t.Fatal("filter did not reset selection")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("no-match")})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.done || m.selected != nil {
		t.Fatal("empty results should not exit")
	}
}
