package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func TestDelegateRender(t *testing.T) {
	entries := []struct {
		status string
		path   string
	}{
		{"M", "src/client/components/hud/HUD.tsx"},
		{"A", "short.go"},
		{"??", "src/very/deeply/nested/directory/structure/file.ts"},
		{"D", "deleted-file.txt"},
	}

	widths := []int{40, 60, 80, 120}

	d := fileDelegate{}

	for _, width := range widths {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			items := make([]list.Item, len(entries))
			for i, e := range entries {
				items[i] = fileEntry{status: e.status, path: e.path}
			}

			l := list.New(items, d, width, 20)
			l.SetShowStatusBar(false)
			l.SetShowTitle(false)
			l.SetShowHelp(false)

			output := l.View()
			lines := strings.Split(output, "\n")

			for i, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				visWidth := lipgloss.Width(line)
				if visWidth > width {
					t.Errorf("line %d exceeds width %d (got %d): %q", i, width, visWidth, line)
				}
			}
		})
	}
}

func TestDelegateRenderContent(t *testing.T) {
	// Verify that rendered output contains the full path (not truncated/corrupted)
	f := fileEntry{status: "M", path: "src/client/components/hud/HUD.tsx"}

	d := fileDelegate{}
	items := []list.Item{f}
	l := list.New(items, d, 80, 20)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	output := l.View()

	// The rendered output should contain key parts of the path
	// Strip ANSI for content check
	plain := stripAnsi(output)

	if !strings.Contains(plain, "src/client") {
		t.Errorf("rendered output missing 'src/client', got:\n%s", plain)
	}
	if !strings.Contains(plain, "HUD.tsx") {
		t.Errorf("rendered output missing 'HUD.tsx', got:\n%s", plain)
	}
	if !strings.Contains(plain, "MOD") {
		t.Errorf("rendered output missing 'MOD' badge, got:\n%s", plain)
	}
}

func TestDelegateSelectedRow(t *testing.T) {
	// First item is selected by default — verify path integrity
	f := fileEntry{status: "M", path: "src/client/components/hud/HUD.tsx"}

	d := fileDelegate{}
	items := []list.Item{f}
	l := list.New(items, d, 80, 20)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	output := l.View()
	lines := strings.Split(output, "\n")

	// Find the line with content
	var contentLine string
	for _, line := range lines {
		if strings.TrimSpace(stripAnsi(line)) != "" {
			contentLine = line
			break
		}
	}

	plain := stripAnsi(contentLine)

	// Selected row should have > cursor and full path
	if !strings.Contains(plain, ">") {
		t.Errorf("selected row missing '>' cursor: %q", plain)
	}
	if !strings.Contains(plain, "src/") {
		t.Errorf("selected row has corrupted path (missing 'src/'): %q", plain)
	}

	// Check visual width doesn't exceed list width
	visWidth := lipgloss.Width(contentLine)
	if visWidth > 80 {
		t.Errorf("selected row visual width %d exceeds 80: %q", visWidth, contentLine)
	}
}

func TestDelegateRenderDump(t *testing.T) {
	// Diagnostic: dump raw render output for visual inspection
	f := fileEntry{status: "M", path: "src/client/components/hud/HUD.tsx"}

	d := fileDelegate{}
	items := []list.Item{f}

	for _, width := range []int{40, 60, 80} {
		l := list.New(items, d, width, 10)
		l.SetShowStatusBar(false)
		l.SetShowTitle(false)
		l.SetShowHelp(false)

		output := l.View()
		lines := strings.Split(output, "\n")
		for i, line := range lines {
			if strings.TrimSpace(stripAnsi(line)) != "" {
				plain := stripAnsi(line)
				t.Logf("[width=%d] line %d (vis=%d plain=%d): plain=%q",
					width, i, lipgloss.Width(line), len(plain), plain)
			}
		}
	}
}

func TestPanelWidthMath(t *testing.T) {
	// Simulate what renderPanel does and check total width
	termWidth := 80
	innerW := termWidth - 4 // from innerSize()

	// The list renders rows at innerW width
	row := "> MOD " + strings.Repeat("x", innerW-7) + " "
	t.Logf("row visual width: %d", lipgloss.Width(row))

	// panelBorder sets Width(innerW) + border chars
	border := panelBorder(false, innerW, 10)
	rendered := border.Render(row)

	renderedLines := strings.Split(rendered, "\n")
	for i, line := range renderedLines {
		vw := lipgloss.Width(line)
		if vw > termWidth {
			t.Errorf("panel line %d exceeds terminal width %d (got %d)", i, termWidth, vw)
		}
		if i == 0 || i == len(renderedLines)-1 {
			t.Logf("border line %d width: %d", i, vw)
		} else if strings.TrimSpace(stripAnsi(line)) != "" {
			t.Logf("content line %d width: %d, plain: %q", i, vw, stripAnsi(line)[:min(60, len(stripAnsi(line)))])
		}
	}
}

func TestFullRenderPipeline(t *testing.T) {
	// Simulate the full render pipeline: list → border → final output
	termWidth := 80
	termHeight := 24
	innerW := termWidth - 4
	innerH := termHeight - 4

	f := fileEntry{status: "M", path: "src/client/components/hud/HUD.tsx"}
	items := []list.Item{f}
	l := list.New(items, fileDelegate{}, innerW, innerH)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	listOutput := l.View()
	listLines := strings.Split(listOutput, "\n")
	for i, line := range listLines {
		vw := lipgloss.Width(line)
		if vw > innerW {
			t.Errorf("list line %d exceeds innerW %d (got %d): plain=%q",
				i, innerW, vw, stripAnsi(line))
		}
	}

	// Now wrap in border like renderPanel does
	border := panelBorder(false, innerW, innerH)
	panelOutput := border.Render(listOutput)
	panelLines := strings.Split(panelOutput, "\n")
	for i, line := range panelLines {
		vw := lipgloss.Width(line)
		if vw > termWidth {
			t.Errorf("panel line %d exceeds termWidth %d (got %d): plain=%q",
				i, termWidth, vw, stripAnsi(line))
		}
	}

	// Final: PlaceHorizontal
	final := lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, panelOutput)
	finalLines := strings.Split(final, "\n")
	for i, line := range finalLines {
		vw := lipgloss.Width(line)
		if vw > termWidth {
			t.Errorf("final line %d exceeds termWidth %d (got %d): plain=%q",
				i, termWidth, vw, stripAnsi(line))
		}
		if strings.TrimSpace(stripAnsi(line)) != "" && i < 5 {
			t.Logf("final line %d (vis=%d): plain=%q", i, vw, stripAnsi(line))
		}
	}
}

func TestParseporcelainPreservesLeadingStatus(t *testing.T) {
	// git status --porcelain uses column 0 for index status and column 1 for
	// worktree status. A space in column 0 means "not staged". TrimSpace on the
	// full output used to strip that leading space from the first line, corrupting
	// the path (e.g. " M src/foo" → "M src/foo" → path parsed as "rc/foo").
	cases := []struct {
		name     string
		raw      string
		wantPath string
		wantStatus string
	}{
		{
			name:       "worktree modified, not staged (leading space)",
			raw:        " M src/client/components/hud/HUD.tsx\n",
			wantPath:   "src/client/components/hud/HUD.tsx",
			wantStatus: "M",
		},
		{
			name:       "staged added",
			raw:        "A  newfile.go\n",
			wantPath:   "newfile.go",
			wantStatus: "A",
		},
		{
			name:       "untracked",
			raw:        "?? some/untracked/file.ts\n",
			wantPath:   "some/untracked/file.ts",
			wantStatus: "??",
		},
		{
			name:       "multiple files, first has leading space status",
			raw:        " M src/first.ts\n M src/second.ts\nA  src/third.ts\n",
			wantPath:   "src/first.ts",
			wantStatus: "M",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Replicate the parsing logic from getChangedFiles
			var files []fileEntry
			for _, line := range strings.Split(strings.TrimRight(tc.raw, "\n"), "\n") {
				if len(line) < 4 {
					continue
				}
				status := strings.TrimSpace(line[:2])
				path := strings.TrimSpace(line[3:])
				if idx := strings.Index(path, " -> "); idx != -1 {
					path = path[idx+4:]
				}
				files = append(files, fileEntry{status: status, path: path})
			}

			if len(files) == 0 {
				t.Fatal("parsed zero files")
			}
			if files[0].path != tc.wantPath {
				t.Errorf("path = %q, want %q", files[0].path, tc.wantPath)
			}
			if files[0].status != tc.wantStatus {
				t.Errorf("status = %q, want %q", files[0].status, tc.wantStatus)
			}
		})
	}
}

func TestParseporcelainTrimSpaceBug(t *testing.T) {
	// Regression: prove that TrimSpace on the full output corrupts the first path
	raw := " M src/client/components/hud/HUD.tsx\n M src/second.ts\n"

	// BAD: old behavior with TrimSpace
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path == "rc/client/components/hud/HUD.tsx" {
			// This is the bug — TrimSpace ate the leading space, shifting the parse
			t.Log("confirmed: TrimSpace produces corrupted path 'rc/client/...'")
		}
		break // only check first line
	}

	// GOOD: fixed behavior with TrimRight
	for _, line := range strings.Split(strings.TrimRight(raw, "\n"), "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path != "src/client/components/hud/HUD.tsx" {
			t.Errorf("TrimRight parse got %q, want %q", path, "src/client/components/hud/HUD.tsx")
		}
		break
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// stripAnsi removes ANSI escape sequences for content testing
func stripAnsi(s string) string {
	var buf bytes.Buffer
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEsc = false
			}
			continue
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}
