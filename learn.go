package main

import (
	"fmt"
	"os/exec"
	"strings"
	"container/list"

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
	displays list
	selected int
	message string
}

func initialModel() model {
	cmd := exec.Command("xrandr", "-q")
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	lines := strings.Split(output, "\n")
	displays := list.New()

	var first bool = true
	var begin, end int = -1, -1
	var nl NameLine
	for i:= 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "connected") {
			
			if (first) {
				begin = i+1
				first = false
			} else {
				end = i
				display := Display{nl.name, nl.connected, nl.current, lines[begin:end]}
				displays.PushBack(display)
				begin = i+1
			}
			nl = extract_metadata(lines[i])
		}
	}

	for e := displays.Front(); e != nil; e = e.Next() {
		display := Display(e.Value.(Display))
		fmt.Println(display.name)
		for i:= 0; i < len(display.resolutions); i++ {
			res := get_res(display.resolutions[i])
		}
	}

	return model{
		displays: displays,
		selected: 0,
		message: lines[0],
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    // Is it a key press?
    case tea.KeyMsg:

        // Cool, what was the actual key pressed?
        switch msg.String() {

        // These keys should exit the program.
        case "ctrl+c", "q":
            return m, tea.Quit

        // The "up" and "k" keys move the cursor up
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }

        // The "down" and "j" keys move the cursor down
        case "down", "j":
            if m.cursor < m.Len() {
                m.cursor++
            }

        // The "enter" key and the spacebar (a literal space) toggle
        // the selected state for the item that the cursor is pointing at.
        case "enter", " ":
            // _, ok := m.selected[m.cursor]
            // if ok {
            //     delete(m.selected, m.cursor)
            // } else {
            //     m.selected[m.cursor] = struct{}{}
            // }
        }
    }

    // Return the updated model to the Bubble Tea runtime for processing.
    // Note that we're not returning a command.
    return m, nil
}

func (m model) View() string {
    s := "Which screen do you want to use?\n\n"

    // for i, choice := range m.choices {
    //     cursor := " " // no cursor
    //     if m.cursor == i {
    //         cursor = ">" // cursor!
    //     }

    //     checked := " " // not selected
    //     if _, ok := m.selected[i]; ok {
    //         checked = "x" // selected!
    //     }

    //     s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    // }

    // The footer
    s += "\nPress q to quit.\n"

    // Send the UI for rendering
    return s
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

