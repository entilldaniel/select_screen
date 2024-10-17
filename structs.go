package main

type Display struct {
	name        string
	connected   bool
	current     bool
	resolutions []string
}

type NameLine struct {
	name      string
	connected bool
	current   bool
}

type Model struct {
	displays    []Display
	selected    int
	resolutions []string
	screen      string
	current     string
	resolution  string
}

type Status string
