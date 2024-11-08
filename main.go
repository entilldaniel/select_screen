package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Init() tea.Cmd {
	return createModel
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func (display Display) FilterValue() string { return "" }

type displayItemDelegate struct{}

func (d displayItemDelegate) Height() int                             { return 1 }
func (d displayItemDelegate) Spacing() int                            { return 0 }
func (d displayItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d displayItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Display)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func createModel() tea.Msg {
	cmd := exec.Command("xrandr", "-q")
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	lines := strings.Split(output, "\n")

	displays := []list.Item{}
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

	l := list.New(displays, displayItemDelegate{}, 20, 14)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	return Model{
		displays:   l,
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
	case tea.WindowSizeMsg:
		m.displays.SetWidth(msg.Width)
		return m, nil
	case Model:
		m = msg
	case Status:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter", " ":
			// if m.screen == "" {
			// 	m.screen = m.displays[m.selected].name
			// 	m.resolutions = m.displays[m.selected].resolutions
			// } else {
			// 	m.resolution = get_res(m.resolutions[m.selected])
			// 	return m, change_resolution(m)
			// }

			//m.selected = 0
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.displays, cmd = m.displays.Update(msg)
	return m, cmd
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
		return "\n" + m.displays.View()
		// s += heading.Render("Which screen do you want to use?")
		// s += "\n\n"

		// alternatives := ""
		// for i := 0; i < len(m.displays); i++ {
		// 	display := m.displays[i]
		// 	current := ""
		// 	if m.displays[i].current {
		// 		current = "(current)"
		// 	}

		// 	if m.selected == i {
		// 		alternatives += fmt.Sprintf(active_alternative.Render("> %s %s"), display.name, current)
		// 	} else {
		// 		alternatives += fmt.Sprintf(alternative.Render("  %s %s"), display.name, current)
		// 	}

		// 	if i < len(m.displays)-1 {
		// 		alternatives += "\n"
		// 	}
		//}

		//s += border.Render(alternatives)
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
