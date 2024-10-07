package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Display struct {
	name string
	connected bool
	current bool
	resolutions []string
}

type NameLine struct {
	name string
	connected bool
	current bool
}

type model struct {
	displays []Display
	selected int
	message string
	resolutions []string
	screen string
	current string
	resolution string
}

func initialModel() model {
	cmd := exec.Command("xrandr", "-q")
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	lines := strings.Split(output, "\n")
	
	var displays []Display
	var first bool = true
	var begin, end int = -1, -1
	var nl NameLine
	var current string

	for i:= 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "connected") {
			
			if (first) {
				begin = i+1
				first = false
			} else {
				end = i
				display := Display{nl.name, nl.connected, nl.current, lines[begin:end]}
				if (display.connected) {
					displays = append(displays, display)
				}
				
				if (display.current) {
					current = display.name
				}

				begin = i+1
			}
			nl = extract_metadata(lines[i])
		}
	}

	return model{
		displays: displays,
		selected: 0,
		message: lines[0],
		screen: "",
		current: current,
		resolution: "",
	}
}

func (m model) Init()  (tea.Model, tea.Cmd) {
	return m, nil
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
            if m.screen == "" && m.selected < len(m.displays)-1 {
                m.selected++
            } else if m.selected < len(m.resolutions)-1 {
				m.selected++
			}

        case "enter", " ":
			if (m.screen == "") {
				m.screen = m.displays[m.selected].name
				m.resolutions = m.displays[m.selected].resolutions
			} else {
				m.resolution = get_res(m.resolutions[m.selected])
				change_resolution(m)
				return m, tea.Quit
			}
			
			m.selected = 0
        }
    }

    return m, nil
}

func (m model) View() string {
    s := "Which screen do you want to use?\n\n"
	for i:= 0; i < len(m.displays); i++ {
		display := m.displays[i]
		cursor := "  "
		if (m.selected == i) {
			cursor = "> "
		}
		s += fmt.Sprintf("%s %s\n", cursor, display.name)
	}
	
	if(m.screen != "") {
		s = "Which resolution do you want?\n\n"
		for i:= 0; i < len(m.resolutions); i++ {
			resolution := get_res(m.resolutions[i])
			cursor := "  "
			if (m.selected == i) {
				cursor = "> "
			}
			s += fmt.Sprintf("%s %s\n", cursor, resolution)
		}
	}

    s += "\nPress q to quit.\n"
    return s
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}

func change_resolution(m model) {
	change := exec.Command("xrandr", "--output", m.screen, "--mode", m.resolution, "--fb", m.resolution, "--primary")
	change.Run()

	if (m.current != m.screen) {
		turnOff := exec.Command("xrandr", "--output", m.current, "--off")
		turnOff.Run()

		moveDesktop := exec.Command("bspc", "desktop", m.current, "--to-monitor", m.screen)
		moveDesktop.Run()
	}
}

func extract_metadata(line string) NameLine {
	parts := strings.Split(line, " ")
	name := parts[0]
	connected := parts[1] == "connected"
	current := parts[2] == "primary"

	return NameLine{name, connected, current}
}

func get_res(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")
	return parts[0]
}

