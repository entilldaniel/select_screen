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

func main() {
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
			nl = test(lines[i])
		}
	}

	for e := displays.Front(); e != nil; e = e.Next() {
		display := Display(e.Value.(Display))
		fmt.Println(display.name)
		for i:= 0; i < len(display.resolutions); i++ {
			res := get_res(display.resolutions[i])
			fmt.Println(res)
		}
	}
}

func test(line string) NameLine {
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

