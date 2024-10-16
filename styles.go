package main

import (
	"github.com/charmbracelet/lipgloss"
)

var border = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	MarginLeft(2).
	PaddingLeft(2).
	PaddingRight(3).
	PaddingTop(1).
	PaddingBottom(1)

var heading = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	PaddingLeft(2)

var footer = lipgloss.NewStyle().
	Faint(true).
	PaddingTop(2).
	PaddingLeft(2).
	PaddingBottom(1)

var alternative = lipgloss.NewStyle().
	Faint(true).
	Bold(false)

var active_alternative = lipgloss.NewStyle().
	Bold(true)
