package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Init() tea.Cmd {
	return nil
}

var (
	docStyle          = lipgloss.NewStyle().Padding(1, 2)
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

	str := fmt.Sprintf("%-6d%s", index+1, i.name)
	if i.current {
		str = fmt.Sprintf("%s (current)", str)
	}

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

	str := fmt.Sprintf("%-6d%s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func createModel() Model {
	cmd := exec.Command("xrandr", "-q")
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	lines := strings.Split(output, "\n")

	var displays []list.Item
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

				displayResolutions := list.New(resolutions, resolutionItemDelegate{}, 32, 20)
				displayResolutions.Title = fmt.Sprintf("Resolution for %s", nl.name)
				displayResolutions.SetShowHelp(false)
				displayResolutions.SetShowPagination(false)
				displayResolutions.SetShowStatusBar(false)

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

	l := list.New(displays, displayItemDelegate{}, 32, 20)
	l.Title = fmt.Sprintf("Current: %s", current)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)

	return Model{
		displays:   l,
		selected:   false,
		current:    current,
		resolution: "",
	}
}

func extract_metadata(line string) NameLine {
	parts := strings.Split(line, " ")
	name := parts[0]
	connected := parts[1] == "connected"
	current := parts[2] == "primary"

	return NameLine{name, "1920x1080", connected, current}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Println("WindowSizeMsg fired.")
		h, v := docStyle.GetFrameSize()
		var cmds []tea.Cmd
		for _, disp := range m.displays.Items() {
			switch d := disp.(type) {
			case Display:
				log.Println("Setting size width: ", msg.Width, " and height: ", msg.Height)
				d.resolutions.SetSize(msg.Width-h, msg.Height-v)
				var resCmd tea.Cmd
				d.resolutions, resCmd = d.resolutions.Update(msg)
				cmds = append(cmds, resCmd)
			default:
				m.displays.Title = "Not that type."
			}
		}

		var dispCmd tea.Cmd
		m.displays.SetSize(msg.Width-h, msg.Height-v)
		m.displays, dispCmd = m.displays.Update(msg)
		cmds = append(cmds, dispCmd)

		return m, tea.Batch(cmds...)

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
			var updateCmd tea.Cmd
			if m.selected {
				m.display.resolutions, updateCmd = m.display.resolutions.Update(msg)
				return m, tea.Batch(updateCmd, change_resolution(m))
			} else {
				var si = m.displays.Index()
				var sn = m.displays.SelectedItem().(Display)
				m.display = sn
				m.selected = true
				m.display.resolutions, updateCmd = m.display.resolutions.Update(msg)

				statusCmd := m.displays.NewStatusMessage(fmt.Sprintf("%d", si))
				return m, tea.Batch(updateCmd, statusCmd)
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
	var sd = m.displays.SelectedItem().(Display)
	resolution := string(m.display.resolutions.SelectedItem().(resolution))
	new := sd.name

	log.Println(fmt.Sprintf("new display: %s, current: %s, res: %s, old res: %s", new, m.current, resolution, m.resolution))
	return func() tea.Msg {
		exec.Command("xrandr", "--output", new, "--mode", resolution, "--fb", resolution, "--primary").Run()

		if m.current != new {
			exec.Command("xrandr", "--output", m.current, "--off").Run()
			exec.Command("bspc", "desktop", m.current, "--to-monitor", new).Run()
			return Status("Changed resolution and moved desktop")
		}
		exec.Command("bspc", "wm", "-r").Run()
		return Status("Changed resolution")
	}
}

func get_res(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")
	return parts[0]
}

func (m Model) View() string {
	if !m.selected {
		return docStyle.Render(m.displays.View())
	} else {
		return docStyle.Render(m.display.resolutions.View())
	}
}

func main() {
	f, _ := tea.LogToFile("/tmp/display-select.log", "")

	p := tea.NewProgram(createModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}

	if err := f.Close(); err != nil {
		panic(1)
	}
}
