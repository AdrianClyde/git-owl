package main

import "github.com/charmbracelet/lipgloss"

// Tokyo Night–inspired palette
var (
	// Base
	colorBg      = lipgloss.Color("#1a1b26")
	colorFg      = lipgloss.Color("#a9b1d6")
	colorFgDim   = lipgloss.Color("#565f89")
	colorFgBright = lipgloss.Color("#c0caf5")

	// Accents
	colorBlue   = lipgloss.Color("#7aa2f7")
	colorCyan   = lipgloss.Color("#7dcfff")
	colorPurple = lipgloss.Color("#bb9af7")
	colorOrange = lipgloss.Color("#ff9e64")

	// Git status
	colorAdded     = lipgloss.Color("#9ece6a") // green
	colorModified  = lipgloss.Color("#7aa2f7") // blue
	colorDeleted   = lipgloss.Color("#f7768e") // red
	colorRenamed   = lipgloss.Color("#ff9e64") // orange
	colorUntracked = lipgloss.Color("#565f89") // dim gray

	// Surfaces
	colorSurface    = lipgloss.Color("#24283b")
	colorSurfaceDim = lipgloss.Color("#1f2335")
	colorHighlight  = lipgloss.Color("#292e42")
	colorBorderDim  = lipgloss.Color("#3b4261")
	colorBorderFocus = lipgloss.Color("#7aa2f7")
)

// ── Status bar ──────────────────────────────────────────────

var (
	logoBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(colorBlue).
			Padding(0, 1)

	branchStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	fileCountStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	statusBarStyle = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorFg)

	headerLine2Style = lipgloss.NewStyle().
				Foreground(colorFg)

	diffBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(colorOrange).
			Padding(0, 1)

	allBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(colorCyan).
			Padding(0, 1)

	previewBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1a1b26")).
				Background(colorCyan).
				Padding(0, 1)
)

// ── Command bar ─────────────────────────────────────────────

var (
	cmdBarStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	cmdKeyStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	cmdDescStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	cmdSepStyle = lipgloss.NewStyle().
			Foreground(colorBorderDim)
)

// ── Panel border ────────────────────────────────────────────

func panelBorder(focused bool, width, height int) lipgloss.Style {
	bc := colorBorderDim
	if focused {
		bc = colorBorderFocus
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(bc).
		Width(width).
		Height(height)
}

// ── File list delegate styles ───────────────────────────────

func statusBadgeStyle(status string) lipgloss.Style {
	bg := statusColorForStatus(status)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1a1b26")).
		Background(bg).
		Width(3).
		Align(lipgloss.Center)
}

var (
	pathDirStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	pathFileStyle = lipgloss.NewStyle().
			Foreground(colorFgBright).
			Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Background(colorHighlight)

	cursorStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)
)

// ── File viewer ─────────────────────────────────────────────

var (
	breadcrumbDirStyle = lipgloss.NewStyle().
				Foreground(colorFgDim)

	breadcrumbFileStyle = lipgloss.NewStyle().
				Foreground(colorBlue).
				Bold(true)

	breadcrumbSepStyle = lipgloss.NewStyle().
				Foreground(colorBorderDim)

	scrollPctStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	separatorStyle = lipgloss.NewStyle().
			Foreground(colorBorderDim)

	lineNumStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Align(lipgloss.Right).
			Width(4)

	lineBarStyle = lipgloss.NewStyle().
			Foreground(colorBorderDim)
)

// ── Animation & header ──────────────────────────────────────

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	owlStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	recentMarkerStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	headerDimStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	headerAccentStyle = lipgloss.NewStyle().
				Foreground(colorCyan)

	headerPulseLineStyle = lipgloss.NewStyle().
				Foreground(colorCyan)

	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBlue).
				Background(colorSurfaceDim).
				Padding(1, 2)

	dirtyIndicatorStyle = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true)

	cleanIndicatorStyle = lipgloss.NewStyle().
				Foreground(colorAdded).
				Bold(true)
)

// ── Filter prompt ───────────────────────────────────────────

var filterPromptStyle = lipgloss.NewStyle().
	Foreground(colorCyan)

// ── Helpers ─────────────────────────────────────────────────

func statusColorForStatus(status string) lipgloss.Color {
	switch status {
	case "A":
		return colorAdded
	case "M":
		return colorModified
	case "D":
		return colorDeleted
	case "R":
		return colorRenamed
	case "??":
		return colorUntracked
	default:
		return colorFgDim
	}
}

func statusLabel(status string) string {
	switch status {
	case "M":
		return "MOD"
	case "A":
		return "ADD"
	case "D":
		return "DEL"
	case "R":
		return "REN"
	case "??":
		return " ? "
	default:
		return "   "
	}
}
