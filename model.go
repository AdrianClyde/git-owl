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
)

type view int

const (
	fileListView view = iota
	fileViewerView
)

// Messages
type filesLoadedMsg struct {
	files  []fileEntry
	branch string
	err    error
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
type fileDelegate struct{}

func (d fileDelegate) Height() int                             { return 1 }
func (d fileDelegate) Spacing() int                            { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	f, ok := item.(fileEntry)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	maxWidth := m.Width()

	// Prefix: cursor (2) + badge (3) + space (1) = 6 chars
	prefix := "  "
	if isSelected {
		prefix = cursorStyle.Render("> ")
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
			if isSelected {
				pathStr = pathDirStyle.Background(colorHighlight).Render(d2+"/") +
					pathFileStyle.Background(colorHighlight).Render(f2)
			} else {
				pathStr = pathDirStyle.Render(d2+"/") + pathFileStyle.Render(f2)
			}
		} else {
			if isSelected {
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
		if isSelected {
			pathStr = pathFileStyle.Background(colorHighlight).Render(name)
		} else {
			pathStr = pathFileStyle.Render(name)
		}
	}

	row := prefix + badge + " " + pathStr

	if isSelected {
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
	allFiles    bool
	currentFile string
	branch      string
	loadSeq     int
	autoRefresh bool
	ready       bool
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

	return model{
		currentView: fileListView,
		list:        l,
		branch:      "?",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadFiles(false), tickCmd())
}

func loadFiles(all bool) tea.Cmd {
	return func() tea.Msg {
		var files []fileEntry
		var err error
		if all {
			files, err = getAllFiles()
		} else {
			files, err = getChangedFiles()
		}
		branch := getCurrentBranch()
		return filesLoadedMsg{files: files, branch: branch, err: err}
	}
}

func loadFileContent(filename string, diffMode bool, status string, seq int) tea.Cmd {
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

	case filesLoadedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.branch = msg.branch
		items := make([]list.Item, len(msg.files))
		for i, f := range msg.files {
			items[i] = f
		}
		m.list.SetItems(items)
		return m, nil

	case fileContentMsg:
		if msg.seq != m.loadSeq {
			return m, nil
		}
		wasAutoRefresh := m.autoRefresh
		m.autoRefresh = false
		if msg.err != nil {
			m.viewport.SetContent(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.viewport.SetContent(msg.content)
		}
		if !wasAutoRefresh {
			m.viewport.GotoTop()
		}
		m.currentView = fileViewerView
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{tickCmd(), loadFiles(m.allFiles)}
		if m.currentView == fileViewerView && m.currentFile != "" {
			m.loadSeq++
			m.autoRefresh = true
			item, ok := m.list.SelectedItem().(fileEntry)
			status := ""
			if ok {
				status = item.status
			}
			cmds = append(cmds, loadFileContent(m.currentFile, m.diffMode, status, m.loadSeq))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.currentView == fileListView && m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
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

func (m model) updateFileList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "enter":
		item, ok := m.list.SelectedItem().(fileEntry)
		if !ok {
			return m, nil
		}
		m.currentFile = item.path
		m.loadSeq++
		innerW, innerH := m.innerSize()
		m.viewport = viewport.New(innerW, innerH-2)
		m.viewport.SetContent("Loading...")
		return m, loadFileContent(item.path, m.diffMode, item.status, m.loadSeq)

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
		m.currentView = fileListView
		return m, nil

	case "d":
		m.diffMode = !m.diffMode
		m.loadSeq++
		item, ok := m.list.SelectedItem().(fileEntry)
		status := ""
		if ok {
			status = item.status
		}
		return m, loadFileContent(m.currentFile, m.diffMode, status, m.loadSeq)

	case "g":
		m.viewport.GotoTop()
		return m, nil

	case "G":
		m.viewport.GotoBottom()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// innerSize returns the content width and height inside the bordered panel.
// Layout: status bar (1) + top border (1) + content + bottom border (1) + command bar (1) = 4 chrome lines
func (m model) innerSize() (int, int) {
	w := m.width - 4  // border left (1) + padding (1) + border right (1) + padding (1)
	h := m.height - 4 // status bar + border top + border bottom + cmd bar
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

	status := m.renderStatusBar()
	cmdbar := m.renderCmdBar()
	panel := m.renderPanel()

	return status + "\n" + panel + "\n" + cmdbar
}

// renderStatusBar builds the full-width top bar.
func (m model) renderStatusBar() string {
	logo := logoBadge.Render("git-watch")
	branch := " " + branchStyle.Render("\ue0a0 "+m.branch) // powerline branch icon
	count := fileCountStyle.Render(fmt.Sprintf("  %d files", len(m.list.Items())))

	left := logo + branch + count

	var badges []string
	if m.diffMode {
		badges = append(badges, diffBadgeStyle.Render("DIFF"))
	}
	if m.allFiles {
		badges = append(badges, allBadgeStyle.Render("ALL"))
	}
	right := strings.Join(badges, " ")

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := m.width - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(m.width).Render(bar)
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
			{"q", "uit"},
		}
	} else {
		hints = []hint{
			{"esc", "back"},
			{"d", "iff"},
			{"g/G", "top/btm"},
			{"j/k", "scroll"},
			{"q", "uit"},
		}
	}

	var parts []string
	for _, h := range hints {
		parts = append(parts,
			cmdKeyStyle.Render("<"+h.key+">")+cmdDescStyle.Render(h.desc))
	}
	bar := strings.Join(parts, cmdSepStyle.Render("  "))
	return cmdBarStyle.Width(m.width).Render(" " + bar)
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

	if m.diffMode {
		breadcrumb += " " + diffBadgeStyle.Render("DIFF")
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
