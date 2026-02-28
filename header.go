package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader produces a 2-line animated header, full terminal width.
func (m model) renderHeader() string {
	// Align with panel content: 1 char centering margin + 1 char border = 2
	indent := "  "
	rightPad := "   "

	logo := logoBadge.Render("Git Owl")

	// Spinner
	var spin string
	if m.paused {
		spin = pausedStyle.Render("⏸")
	} else if !m.animating {
		spin = spinnerStyle.Render("●")
	} else {
		spin = spinnerStyle.Render(m.spinner.view())
	}

	branch := branchStyle.Render("⏵ " + m.branch)

	// Dirty indicator + file count
	fileCount := len(m.list.Items())
	var dirty string
	if fileCount > 0 && !m.allFiles {
		dirty = dirtyIndicatorStyle.Render("●")
	} else {
		dirty = cleanIndicatorStyle.Render("✓")
	}
	count := fileCountStyle.Render(fmt.Sprintf("%d files", fileCount))

	// ── Line 1: logo ... badges + owl top ──
	line1Left := indent + logo

	// Build right side: badges first, then owl (so owl is always flush-right)
	var line1RightParts []string
	if m.diffMode {
		line1RightParts = append(line1RightParts, diffBadgeStyle.Render("DIFF"))
	}
	if m.allFiles {
		line1RightParts = append(line1RightParts, allBadgeStyle.Render("ALL"))
	}
	if !m.minimal {
		line1RightParts = append(line1RightParts, owlStyle.Render(owlTop()))
	}
	line1Right := strings.Join(line1RightParts, " ")
	if line1Right != "" {
		line1Right += rightPad
	}

	gap1 := m.width - lipgloss.Width(line1Left) - lipgloss.Width(line1Right)
	if gap1 < 1 {
		gap1 = 1
	}
	bar1 := line1Left + strings.Repeat(" ", gap1) + line1Right

	// ── Line 2: spinner + branch + files ... owl bottom ──
	line2Left := indent + spin + "  " + branch + "  " + dirty + " " + count

	var line2Right string
	if !m.minimal {
		line2Right = owlStyle.Render(m.owl.owlBottom()) + rightPad
	}

	gap2 := m.width - lipgloss.Width(line2Left) - lipgloss.Width(line2Right)
	if gap2 < 1 {
		gap2 = 1
	}
	bar2 := line2Left + strings.Repeat(" ", gap2) + line2Right

	rendered1 := headerLine2Style.Width(m.width).Render(bar1)
	rendered2 := headerLine2Style.Width(m.width).Render(bar2)
	return rendered1 + "\n" + rendered2
}

// renderWithHelpOverlay renders the help overlay centered over the panel.
func (m model) renderWithHelpOverlay(header, panel, cmdbar string) string {
	helpContent := m.renderHelpContent()
	overlay := helpOverlayStyle.Render(helpContent)

	panelHeight := lipgloss.Height(panel)

	placed := lipgloss.Place(
		m.width, panelHeight,
		lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)

	return header + "\n" + placed + "\n" + cmdbar
}

// renderHelpContent produces the help text.
func (m model) renderHelpContent() string {
	title := headerAccentStyle.Render("Keybindings") + "\n\n"

	type binding struct{ key, desc string }
	bindings := []binding{
		{"enter", "Open file viewer"},
		{"esc", "Back to file list"},
		{"d", "Toggle diff mode"},
		{"t", "Toggle all files"},
		{"r", "Refresh file list"},
		{"/", "Filter files"},
		{"space", "Pause/resume refresh"},
		{"a", "Toggle animations"},
		{"m", "Toggle minimal mode"},
		{"?", "Toggle this help"},
		{"g/G", "Go to top/bottom"},
		{"j/k", "Scroll up/down"},
		{"q", "Quit"},
	}

	var lines []string
	for _, b := range bindings {
		key := cmdKeyStyle.Render(fmt.Sprintf("%8s", b.key))
		desc := headerDimStyle.Render("  " + b.desc)
		lines = append(lines, key+desc)
	}

	return title + strings.Join(lines, "\n")
}
