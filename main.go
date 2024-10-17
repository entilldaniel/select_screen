package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return createModel
}

func createModel() tea.Msg {
	cmd := exec.Command("xrandr", "-q")
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	lines := strings.Split(output, "\n")

	var displays []Display
	var first bool = true
	var begin, end int = -1, -1
	var nl NameLine
	var current string

	for i := 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "connected") {

			if first {
				begin = i + 1
				first = false
			} else {
				end = i
				display := Display{nl.name, nl.connected, nl.current, lines[begin:end]}
				if display.connected {
					displays = append(displays, display)
				}

				if display.current {
					current = display.name
				}

				begin = i + 1
			}
			nl = extract_metadata(lines[i])
		}
	}

	return Model{
		displays:   displays,
		selected:   0,
		screen:     "",
		current:    current,
		resolution: "",
	}
}

func extract_metadata(line string) NameLine {
	parts := strings.Split(line, " ")
	name := parts[0]
	connected := parts[1] == "connected"
	current := parts[2] == "primary"

	return NameLine{name, connected, current}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case Model:
		m = msg
	case Status:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.screen == "" && m.selected < len(m.displays)-1 {
				m.selected++
			} else if m.selected < len(m.resolutions)-1 {
				m.selected++
			}

		case "enter", " ":
			if m.screen == "" {
				m.screen = m.displays[m.selected].name
				m.resolutions = m.displays[m.selected].resolutions
			} else {
				m.resolution = get_res(m.resolutions[m.selected])
				return m, change_resolution(m)
			}

			m.selected = 0
		}
	}

	return m, nil
}

func change_resolution(m Model) tea.Cmd {
	return func() tea.Msg {
		exec.Command("xrandr", "--output", m.screen, "--mode", m.resolution, "--fb", m.resolution, "--primary").Run()

		if m.current != m.screen {
			exec.Command("xrandr", "--output", m.current, "--off").Run()
			exec.Command("bspc", "desktop", m.current, "--to-monitor", m.screen).Run()
			return Status("Changed resolution and moved desktop")
		}
		return Status("Changed resolution")
	}
}

func get_res(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")
	return parts[0]
}

func (m Model) View() string {
	s := ""
	if m.screen == "" {
		s += heading.Render("Which screen do you want to use?")
		s += "\n\n"

		alternatives := ""
		for i := 0; i < len(m.displays); i++ {
			display := m.displays[i]
			current := ""
			if m.displays[i].current {
				current = "(current)"
			}

			if m.selected == i {
				alternatives += fmt.Sprintf(active_alternative.Render("> %s %s"), display.name, current)
			} else {
				alternatives += fmt.Sprintf(alternative.Render("  %s %s"), display.name, current)
			}

			if i < len(m.displays)-1 {
				alternatives += "\n"
			}
		}

		s += border.Render(alternatives)
	} else {
		s = heading.Render("Which resolution do you want?")
		s += "\n\n"

		alternatives := ""
		for i := 0; i < len(m.resolutions); i++ {
			resolution := get_res(m.resolutions[i])
			if m.selected == i {
				alternatives += active_alternative.Render(fmt.Sprintf("> %s", resolution))
			} else {
				alternatives += alternative.Render(fmt.Sprintf("  %s", resolution))
			}
			if i < len(m.resolutions)-1 {
				alternatives += "\n"
			}

		}

		s += border.Render(alternatives)
	}

	s += footer.Render("Press q to quit.\n")
	return s
}

func main() {
	p := tea.NewProgram(Model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
