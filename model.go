package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type view int

const (
	fileListView view = iota
	fileViewerView
)

// Messages
type filesLoadedMsg struct {
	files    []fileEntry
	branch   string
	err      error
	scanTime time.Duration
}

type fileContentMsg struct {
	content  string
	filename string
	seq      int
	err      error
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Custom list delegate for colored badge rows
type fileDelegate struct {
	recentFiles map[string]bool
}

func (d fileDelegate) Height() int                             { return 1 }
func (d fileDelegate) Spacing() int                            { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	f, ok := item.(fileEntry)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	isRecent := d.recentFiles[f.path]
	maxWidth := m.Width()

	// Prefix: cursor/marker (2) + badge (3) + space (1) = 6 chars
	var prefix string
	if isRecent {
		prefix = recentMarkerStyle.Render("✦ ")
	} else if isSelected {
		prefix = cursorStyle.Render("> ")
	} else {
		prefix = "  "
	}
	badge := statusBadgeStyle(f.status).Render(statusLabel(f.status))

	// Split path into dir + filename
	dir, file := splitPath(f.path)

	// Truncate path to fit: maxWidth - 6 (prefix+badge+space)
	pathBudget := maxWidth - 7
	if pathBudget < 10 {
		pathBudget = 10
	}

	var pathStr string
	if dir != "" {
		fullPath := dir + "/" + file
		if len(fullPath) > pathBudget {
			fullPath = fullPath[:pathBudget-1] + "…"
		}
		// Re-split after truncation
		d2, f2 := splitPath(fullPath)
		if d2 != "" {
			if isSelected || isRecent {
				pathStr = pathDirStyle.Background(colorHighlight).Render(d2+"/") +
					pathFileStyle.Background(colorHighlight).Render(f2)
			} else {
				pathStr = pathDirStyle.Render(d2+"/") + pathFileStyle.Render(f2)
			}
		} else {
			if isSelected || isRecent {
				pathStr = pathFileStyle.Background(colorHighlight).Render(f2)
			} else {
				pathStr = pathFileStyle.Render(f2)
			}
		}
	} else {
		name := file
		if len(name) > pathBudget {
			name = name[:pathBudget-1] + "…"
		}
		if isSelected || isRecent {
			pathStr = pathFileStyle.Background(colorHighlight).Render(name)
		} else {
			pathStr = pathFileStyle.Render(name)
		}
	}

	row := prefix + badge + " " + pathStr

	if isSelected || isRecent {
		rowLen := lipgloss.Width(row)
		if rowLen < maxWidth {
			pad := selectedRowStyle.Render(strings.Repeat(" ", maxWidth-rowLen))
			row += pad
		}
	}

	fmt.Fprint(w, row)
}

func splitPath(path string) (dir, file string) {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return "", path
	}
	return path[:idx], path[idx+1:]
}

type model struct {
	currentView view
	list        list.Model
	viewport    viewport.Model
	width       int
	height      int
	diffMode    bool
	mdPreview   bool
	allFiles    bool
	currentFile string
	branch      string
	loadSeq     int
	autoRefresh bool
	ready       bool

	// Animation
	spinner  spinnerState
	owl      owlState
	showHelp bool // '?' toggles

	// Snapshot diffing & events
	prevSnapshot snapshot
	events       eventsRing
	recentFiles  map[string]bool // paths with recent changes (for row ✦ markers)

	// Horizontal scroll
	hScroll    int    // current horizontal offset in visible columns
	rawContent string // unshifted file content for re-applying offset

	// Header pulse
	headerPulse int // frames remaining (decremented by animTick)

	// Metrics
	lastScanTime time.Duration // how long the last git status call took
	lastScanAt   time.Time     // when last scan completed
}

func initialModel() model {
	l := list.New(nil, fileDelegate{}, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.Prompt = "/ "
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(colorFgDim).Padding(1, 2)
	l.Styles.TitleBar = lipgloss.NewStyle() // remove default bottom padding

	return model{
		currentView: fileListView,
		list:        l,
		branch:      "?",
		owl:         newOwlState(),
		events:      newEventsRing(5),
		recentFiles: map[string]bool{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadFiles(false), tickCmd(), animTickCmd())
}

func loadFiles(all bool) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		var files []fileEntry
		var err error
		if all {
			files, err = getAllFiles()
		} else {
			files, err = getChangedFiles()
		}
		branch := getCurrentBranch()
		elapsed := time.Since(start)
		return filesLoadedMsg{files: files, branch: branch, err: err, scanTime: elapsed}
	}
}

func loadFileContent(filename string, diffMode, mdPreview bool, status string, seq int) tea.Cmd {
	return func() tea.Msg {
		if diffMode && status != "??" {
			diff, err := getDiff(filename)
			if err != nil {
				return fileContentMsg{err: err, filename: filename, seq: seq}
			}
			if strings.TrimSpace(diff) != "" {
				highlighted := highlightDiff(diff)
				return fileContentMsg{content: highlighted, filename: filename, seq: seq}
			}
		}

		if status == "D" {
			diff, err := getDiff(filename)
			if err == nil && strings.TrimSpace(diff) != "" {
				highlighted := highlightDiff(diff)
				return fileContentMsg{content: highlighted, filename: filename, seq: seq}
			}
			return fileContentMsg{content: "(file deleted)", filename: filename, seq: seq}
		}

		content, err := readFile(filename)
		if err != nil {
			return fileContentMsg{err: err, filename: filename, seq: seq}
		}

		if isBinary(content) {
			return fileContentMsg{content: "(binary file)", filename: filename, seq: seq}
		}

		if mdPreview && strings.HasSuffix(strings.ToLower(filename), ".md") {
			rendered := renderMarkdown(content, 80)
			return fileContentMsg{content: rendered, filename: filename, seq: seq}
		}

		highlighted := highlightContent(content, filename)
		return fileContentMsg{content: highlighted, filename: filename, seq: seq}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		innerW, innerH := m.innerSize()
		m.list.SetSize(innerW, innerH)
		if m.currentView == fileViewerView {
			m.viewport.Width = innerW
			m.viewport.Height = innerH - 2 // breadcrumb + separator
		}
		m.ready = true
		return m, nil

	case animTickMsg:
		cmd := animTickCmd() // always reschedule
		m.spinner.tick()
		m.owl.tick()
		if m.headerPulse > 0 {
			m.headerPulse--
		}
		m.recentFiles = m.events.recentPaths(2 * time.Second)
		return m, cmd

	case filesLoadedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.branch = msg.branch
		m.lastScanAt = time.Now()
		m.lastScanTime = msg.scanTime

		// Snapshot diffing
		if m.prevSnapshot.files != nil {
			changes := m.prevSnapshot.diff(msg.files)
			if len(changes) > 0 {
				m.events.push(changes)
				m.headerPulse = 6 // ~600ms at 100ms ticks
			}
		}
		m.prevSnapshot = newSnapshot(msg.files)

		// Don't update items while user is actively filtering — it resets the filter
		if m.list.FilterState() != list.Filtering {
			items := make([]list.Item, len(msg.files))
			for i, f := range msg.files {
				items[i] = f
			}
			m.list.SetItems(items)
		}
		return m, nil

	case fileContentMsg:
		if msg.seq != m.loadSeq {
			return m, nil
		}
		wasAutoRefresh := m.autoRefresh
		m.autoRefresh = false
		if msg.err != nil {
			m.rawContent = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.rawContent = msg.content
		}
		innerW, _ := m.innerSize()
		showNums := !m.diffMode && !m.mdPreview
		m.viewport.SetContent(applyHScroll(m.rawContent, m.hScroll, innerW, showNums))
		if !wasAutoRefresh {
			m.viewport.GotoTop()
		}
		m.currentView = fileViewerView
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{tickCmd()}
		cmds = append(cmds, loadFiles(m.allFiles))
		if m.currentView == fileViewerView && m.currentFile != "" {
			m.loadSeq++
			m.autoRefresh = true
			item, ok := m.list.SelectedItem().(fileEntry)
			status := ""
			if ok {
				status = item.status
			}
			cmds = append(cmds, loadFileContent(m.currentFile, m.diffMode, m.mdPreview, status, m.loadSeq))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.currentView == fileListView && m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// Global keybindings
		if mdl, cmd, handled := m.handleGlobalKey(msg.String()); handled {
			return mdl, cmd
		}

		switch m.currentView {
		case fileListView:
			return m.updateFileList(msg)
		case fileViewerView:
			return m.updateFileViewer(msg)
		}
	}

	if m.currentView == fileListView {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) handleGlobalKey(key string) (model, tea.Cmd, bool) {
	switch key {
	case "?":
		m.showHelp = !m.showHelp
		return m, nil, true
	}
	return m, nil, false
}

func (m model) updateFileList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		if m.showHelp {
			m.showHelp = false
		}
		return m, nil

	case "enter":
		item, ok := m.list.SelectedItem().(fileEntry)
		if !ok {
			return m, nil
		}
		m.currentFile = item.path
		m.hScroll = 0
		m.mdPreview = strings.HasSuffix(strings.ToLower(item.path), ".md")
		m.loadSeq++
		innerW, innerH := m.innerSize()
		m.viewport = viewport.New(innerW, innerH-2)
		m.viewport.SetContent("Loading...")
		return m, loadFileContent(item.path, m.diffMode, m.mdPreview, item.status, m.loadSeq)

	case "t":
		m.allFiles = !m.allFiles
		return m, loadFiles(m.allFiles)

	case "d":
		m.diffMode = !m.diffMode
		return m, nil

	case "r":
		return m, loadFiles(m.allFiles)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateFileViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		m.currentView = fileListView
		return m, nil

	case "d":
		m.diffMode = !m.diffMode
		m.hScroll = 0
		m.loadSeq++
		item, ok := m.list.SelectedItem().(fileEntry)
		status := ""
		if ok {
			status = item.status
		}
		return m, loadFileContent(m.currentFile, m.diffMode, m.mdPreview, status, m.loadSeq)

	case "p":
		if strings.HasSuffix(strings.ToLower(m.currentFile), ".md") {
			m.mdPreview = !m.mdPreview
			m.hScroll = 0
			m.loadSeq++
			item, ok := m.list.SelectedItem().(fileEntry)
			status := ""
			if ok {
				status = item.status
			}
			return m, loadFileContent(m.currentFile, m.diffMode, m.mdPreview, status, m.loadSeq)
		}
		return m, nil

	case "g":
		m.viewport.GotoTop()
		return m, nil

	case "G":
		m.viewport.GotoBottom()
		return m, nil

	case "h", "left":
		m.hScroll -= 4
		if m.hScroll < 0 {
			m.hScroll = 0
		}
		innerW, _ := m.innerSize()
		showNums := !m.diffMode && !m.mdPreview
		yoff := m.viewport.YOffset
		m.viewport.SetContent(applyHScroll(m.rawContent, m.hScroll, innerW, showNums))
		m.viewport.SetYOffset(yoff)
		return m, nil

	case "l", "right":
		m.hScroll += 4
		innerW, _ := m.innerSize()
		showNums := !m.diffMode && !m.mdPreview
		yoff := m.viewport.YOffset
		m.viewport.SetContent(applyHScroll(m.rawContent, m.hScroll, innerW, showNums))
		m.viewport.SetYOffset(yoff)
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// applyHScroll shifts each line of content horizontally using ANSI-aware truncation.
// When lineNums is true, a fixed line-number gutter is prepended to each line.
func applyHScroll(content string, offset, width int, lineNums bool) string {
	// Replace tabs with spaces so width counting matches terminal rendering.
	content = strings.ReplaceAll(content, "\t", "    ")

	lines := strings.Split(content, "\n")

	// Calculate gutter width for line numbers
	gutterW := 0
	var numStyle lipgloss.Style
	if lineNums {
		digits := len(fmt.Sprintf("%d", len(lines)))
		if digits < 3 {
			digits = 3
		}
		numStyle = lineNumStyle.Width(digits)
		gutterW = digits + 1 + 1 // digits + bar + space
	}

	contentW := width - gutterW
	if contentW < 10 {
		contentW = 10
	}

	for i, line := range lines {
		if offset > 0 {
			line = ansi.TruncateLeft(line, offset, "")
		}
		if contentW > 0 {
			line = ansi.Truncate(line, contentW, "")
		}
		if lineNums {
			num := fmt.Sprintf("%d", i+1)
			line = numStyle.Render(num) + lineBarStyle.Render("│") + " " + line
		}
		// Safety: ensure final composed line fits within width
		if width > 0 {
			line = ansi.Truncate(line, width, "")
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// innerSize returns the content width and height inside the bordered panel.
// Layout: header (2) + top border (1) + content + bottom border (1) + command bar (1) = 5 chrome lines
func (m model) innerSize() (int, int) {
	w := m.width - 4 // border left (1) + padding (1) + border right (1) + padding (1)
	h := m.height - 5 // header (2) + border top + border bottom + cmd bar
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	return w, h
}

// ── View ────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Update delegate with recent files before rendering
	m.list.SetDelegate(fileDelegate{recentFiles: m.recentFiles})

	header := m.renderHeader()
	cmdbar := m.renderCmdBar()
	panel := m.renderPanel()

	if m.showHelp {
		return m.renderWithHelpOverlay(header, panel, cmdbar)
	}

	return header + "\n" + panel + "\n" + cmdbar
}

// renderCmdBar builds the bottom command hint bar.
func (m model) renderCmdBar() string {
	type hint struct{ key, desc string }

	var hints []hint
	if m.currentView == fileListView {
		hints = []hint{
			{"enter", "view"},
			{"d", "iff"},
			{"t", "oggle all"},
			{"r", "efresh"},
			{"/", "filter"},
		}
	} else {
		hints = []hint{
			{"esc", "back"},
			{"d", "iff"},
			{"p", "review"},
			{"g/G", "top/btm"},
			{"j/k", "scroll"},
			{"h/l", "pan"},
		}
	}

	// Global hints
	hints = append(hints,
		hint{"?", "help"},
		hint{"q", "uit"},
	)

	var parts []string
	for _, h := range hints {
		parts = append(parts,
			cmdKeyStyle.Render("<"+h.key+">")+cmdDescStyle.Render(h.desc))
	}

	bar := strings.Join(parts, cmdSepStyle.Render("  "))
	// Align with panel content: 1 char centering margin + 1 char border = 2
	return cmdBarStyle.Width(m.width).Render("  " + bar)
}

// renderPanel wraps the main content in a rounded border.
func (m model) renderPanel() string {
	focused := m.currentView == fileViewerView
	innerW, innerH := m.innerSize()

	var content string
	switch m.currentView {
	case fileListView:
		content = m.list.View()
	case fileViewerView:
		content = m.renderFileViewer(innerW)
	}

	border := panelBorder(focused, innerW, innerH)

	// Pulse: change top border color when headerPulse > 0
	if m.headerPulse > 0 {
		border = border.BorderForeground(colorCyan)
	}

	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, border.Render(content))
}

// renderFileViewer produces breadcrumb header + separator + viewport content.
func (m model) renderFileViewer(width int) string {
	// Breadcrumb
	parts := strings.Split(m.currentFile, "/")
	var crumbs []string
	for i, p := range parts {
		if i == len(parts)-1 {
			crumbs = append(crumbs, breadcrumbFileStyle.Render(p))
		} else {
			crumbs = append(crumbs, breadcrumbDirStyle.Render(p))
		}
	}
	breadcrumb := strings.Join(crumbs, breadcrumbSepStyle.Render(" / "))

	// Status badge if available
	if item, ok := m.list.SelectedItem().(fileEntry); ok {
		breadcrumb += " " + statusBadgeStyle(item.status).Render(statusLabel(item.status))
	}

	if m.diffMode {
		breadcrumb += " " + diffBadgeStyle.Render("DIFF")
	}
	if m.mdPreview {
		breadcrumb += " " + previewBadgeStyle.Render("PREVIEW")
	}

	// Scroll percentage right-aligned
	pct := fmt.Sprintf("%.0f%%", m.viewport.ScrollPercent()*100)
	pctStr := scrollPctStyle.Render(pct)
	crumbW := lipgloss.Width(breadcrumb)
	pctW := lipgloss.Width(pctStr)
	headerGap := width - crumbW - pctW
	if headerGap < 1 {
		headerGap = 1
	}
	header := breadcrumb + strings.Repeat(" ", headerGap) + pctStr

	// Separator
	sep := separatorStyle.Render(strings.Repeat("─", width))

	return header + "\n" + sep + "\n" + m.viewport.View()
}
