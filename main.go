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

func (display Display) FilterValue() string { return display.name }

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

type resolution string

func (r resolution) FilterValue() string { return "" }

type resolutionItemDelegate struct{}

func (r resolutionItemDelegate) Height() int                             { return 1 }
func (r resolutionItemDelegate) Spacing() int                            { return 0 }
func (r resolutionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (r resolutionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(resolution)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

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
				resolutions := []list.Item{}
				for _, res := range lines[begin:end] {
					resolutions = append(resolutions, resolution(get_res(res)))
				}
				displayResolutions := list.New(resolutions, resolutionItemDelegate{}, 32, 14)
				displayResolutions.Title = fmt.Sprintf("Select resoultion for %s", nl.name)
				display := Display{nl.name, nl.connected, nl.current, displayResolutions}
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
	l.SetShowStatusBar(true)
	l.Title = current
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	return Model{
		displays:   l,
		selected:   false,
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
		return m, nil
	case Model:
		m = msg
	case Status:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "b":
			m.selected = false
			return m, nil
		case "enter", " ":
			if m.selected {

			} else {
				var si = m.displays.Index()
				var sn = m.displays.SelectedItem().(Display)
				m.display = sn
				m.selected = true
				statusCmd := m.displays.NewStatusMessage(fmt.Sprintf("%d", si))
				return m, tea.Cmd(statusCmd)
			}
		}
	}

	var cmd tea.Cmd
	if !m.selected {
		m.displays, cmd = m.displays.Update(msg)
	} else {
		m.display.resolutions, cmd = m.display.resolutions.Update(msg)
	}
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
	if !m.selected {
		return "\n" + m.displays.View()
	} else {
		return "\n" + m.display.resolutions.View()
	}
}

func main() {
	p := tea.NewProgram(Model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
