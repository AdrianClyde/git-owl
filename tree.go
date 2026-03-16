package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// treeNode represents a file or directory in the hierarchical tree.
type treeNode struct {
	name     string      // segment name ("src", "main.go")
	path     string      // full repo-relative path
	isDir    bool
	status   string      // git status (files only)
	children []*treeNode
}

// treeEntry wraps a treeNode as a list.Item for display in the current directory.
type treeEntry struct {
	node *treeNode
}

func (t treeEntry) Title() string       { return t.node.name }
func (t treeEntry) Description() string { return "" }
func (t treeEntry) FilterValue() string { return t.node.path }

// buildTree constructs a tree from a flat list of file entries.
func buildTree(files []fileEntry) *treeNode {
	root := &treeNode{name: "", path: "", isDir: true}

	for _, f := range files {
		parts := strings.Split(f.path, "/")
		cur := root
		for i, part := range parts {
			isLast := i == len(parts)-1
			partPath := strings.Join(parts[:i+1], "/")

			// Find existing child
			var child *treeNode
			for _, c := range cur.children {
				if c.name == part {
					child = c
					break
				}
			}

			if child == nil {
				child = &treeNode{
					name:  part,
					path:  partPath,
					isDir: !isLast,
				}
				if isLast {
					child.status = f.status
				}
				cur.children = append(cur.children, child)
			}
			cur = child
		}
	}

	sortTree(root)
	return root
}

func sortTree(node *treeNode) {
	sort.Slice(node.children, func(i, j int) bool {
		a, b := node.children[i], node.children[j]
		if a.isDir != b.isDir {
			return a.isDir // dirs first
		}
		return a.name < b.name
	})
	for _, c := range node.children {
		if c.isDir {
			sortTree(c)
		}
	}
}

// childItems returns the immediate children of a node as list items.
func childItems(node *treeNode) []list.Item {
	items := make([]list.Item, len(node.children))
	for i, c := range node.children {
		items[i] = treeEntry{node: c}
	}
	return items
}

// allFileItems recursively collects all file (non-directory) nodes as list items.
func allFileItems(root *treeNode) []list.Item {
	var items []list.Item
	var walk func(node *treeNode)
	walk = func(node *treeNode) {
		for _, child := range node.children {
			if child.isDir {
				walk(child)
			} else {
				items = append(items, treeEntry{node: child})
			}
		}
	}
	walk(root)
	return items
}

// findNodeByPath finds a node by its full path.
func findNodeByPath(root *treeNode, path string) *treeNode {
	if path == "" || path == root.path {
		return root
	}
	var result *treeNode
	var walk func(node *treeNode)
	walk = func(node *treeNode) {
		for _, child := range node.children {
			if child.path == path {
				result = child
				return
			}
			if child.isDir {
				walk(child)
				if result != nil {
					return
				}
			}
		}
	}
	walk(root)
	return result
}

// parentPath returns the parent directory path, or "" for top-level.
func parentPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return ""
	}
	return path[:idx]
}

// treeDelegate renders tree entries — folders with icons, files with status badges.
type treeDelegate struct {
	recentFiles map[string]bool
}

func (d treeDelegate) Height() int                             { return 1 }
func (d treeDelegate) Spacing() int                            { return 0 }
func (d treeDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d treeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(treeEntry)
	if !ok {
		return
	}
	node := entry.node
	isSelected := index == m.Index()
	isRecent := d.recentFiles[node.path]
	maxWidth := m.Width()

	// Cursor/marker prefix
	var prefix string
	if isRecent {
		prefix = recentMarkerStyle.Render("✦ ")
	} else if isSelected {
		prefix = cursorStyle.Render("> ")
	} else {
		prefix = "  "
	}

	var nameStr string
	if node.isDir {
		icon := treeFolderCollapsedStyle.Render("▶")
		dirName := treeFolderNameStyle.Render(node.name + "/")
		if isSelected || isRecent {
			dirName = treeFolderNameStyle.Background(colorHighlight).Render(node.name + "/")
		}
		nameStr = icon + " " + dirName
	} else {
		badge := statusBadgeStyle(node.status).Render(statusLabel(node.status))
		fileName := pathFileStyle.Render(node.name)
		if isSelected || isRecent {
			fileName = pathFileStyle.Background(colorHighlight).Render(node.name)
		}
		nameStr = badge + " " + fileName
	}

	row := prefix + nameStr

	// Truncate to width
	rowLen := lipgloss.Width(row)
	if rowLen > maxWidth {
		row = row[:maxWidth]
	}

	if isSelected || isRecent {
		rowLen = lipgloss.Width(row)
		if rowLen < maxWidth {
			pad := selectedRowStyle.Render(strings.Repeat(" ", maxWidth-rowLen))
			row += pad
		}
	}

	fmt.Fprint(w, row)
}
