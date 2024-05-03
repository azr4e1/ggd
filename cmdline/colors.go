package cmdline

import (
	"github.com/charmbracelet/lipgloss"
)

var offsetStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#3C3C3C"))

var inputStyle = lipgloss.NewStyle().
	Italic(true).
	Foreground(lipgloss.Color("#7D56F4"))

var hexCodesStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#04B575"))
