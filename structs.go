package main

import (
	"github.com/charmbracelet/bubbles/list"
)

type Display struct {
	name        string
	connected   bool
	current     bool
	resolutions list.Model
}

type NameLine struct {
	name      string
	connected bool
	current   bool
}

type Model struct {
	displays    list.Model
	selected    bool
	display     Display
	resolutions []string
	screen      string
	current     string
	resolution  string
}

type Status string
