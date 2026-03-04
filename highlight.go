package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	mermaidCmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

var (
	style     = styles.Get("monokai")
	formatter = formatters.Get("terminal256")

	mermaidBlockRe = regexp.MustCompile("(?s)```mermaid\\n(.*?)```")
)

func highlight(content, filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return content
	}
	return buf.String()
}

func highlightDiff(diff string) string {
	lexer := lexers.Get("diff")
	if lexer == nil {
		return diff
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, diff)
	if err != nil {
		return diff
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return diff
	}
	return buf.String()
}

func boolPtr(b bool) *bool    { return &b }
func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint    { return &u }

// Tokyo Night–inspired Glamour style to match the app theme.
var mdStyle = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#a9b1d6"),
		},
		Margin: uintPtr(2),
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H1OPEN",
			Suffix: "H1CLOSE",
			Color:  stringPtr("#ffffff"),
			Bold:   boolPtr(true),
			Upper:  boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H2OPEN",
			Suffix: "H2CLOSE",
			Color:  stringPtr("#c0caf5"),
			Bold:   boolPtr(true),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H3OPEN",
			Suffix: "H3CLOSE",
			Color:  stringPtr("#a9b1d6"),
			Bold:   boolPtr(true),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H4OPEN",
			Suffix: "H4CLOSE",
			Color:  stringPtr("#a9b1d6"),
			Bold:   boolPtr(true),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H5OPEN",
			Suffix: "H5CLOSE",
			Color:  stringPtr("#7aa2f7"),
			Bold:   boolPtr(true),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "H6OPEN",
			Suffix: "H6CLOSE",
			Color:  stringPtr("#565f89"),
			Bold:   boolPtr(true),
		},
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#565f89"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	Paragraph: ansi.StyleBlock{},
	List: ansi.StyleList{
		LevelIndent: 2,
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#a9b1d6"),
			},
		},
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
		Color:       stringPtr("#a9b1d6"),
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#7aa2f7"),
	},
	Task: ansi.StyleTask{
		Ticked:   "[✓] ",
		Unticked: "[ ] ",
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#9ece6a"),
		},
	},
	Strong: ansi.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#c0caf5"),
	},
	Emph: ansi.StylePrimitive{
		Italic: boolPtr(true),
		Color:  stringPtr("#bb9af7"),
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
		Color:      stringPtr("#565f89"),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#3b4261"),
		Format: "\nHRPLACEHOLDER\n",
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#7dcfff"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#7aa2f7"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#bb9af7"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#bb9af7"),
		Format: "🖼  {{.text}}",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#9ece6a"),
			BackgroundColor: stringPtr("#24283b"),
			Prefix:          " ",
			Suffix:          " ",
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#a9b1d6"),
			},
			Margin: uintPtr(2),
		},
		Chroma: &ansi.Chroma{
			Text:                ansi.StylePrimitive{Color: stringPtr("#a9b1d6")},
			Error:               ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			Comment:             ansi.StylePrimitive{Color: stringPtr("#565f89"), Italic: boolPtr(true)},
			CommentPreproc:      ansi.StylePrimitive{Color: stringPtr("#565f89")},
			Keyword:             ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			KeywordReserved:     ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			KeywordNamespace:    ansi.StylePrimitive{Color: stringPtr("#7dcfff")},
			KeywordType:         ansi.StylePrimitive{Color: stringPtr("#2ac3de")},
			Operator:            ansi.StylePrimitive{Color: stringPtr("#89ddff")},
			Punctuation:         ansi.StylePrimitive{Color: stringPtr("#a9b1d6")},
			Name:                ansi.StylePrimitive{Color: stringPtr("#c0caf5")},
			NameBuiltin:         ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
			NameTag:             ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			NameAttribute:       ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			NameClass:           ansi.StylePrimitive{Color: stringPtr("#2ac3de")},
			NameConstant:        ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			NameDecorator:       ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			NameFunction:        ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
			LiteralNumber:       ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			LiteralString:       ansi.StylePrimitive{Color: stringPtr("#9ece6a")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#89ddff")},
			GenericDeleted:      ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			GenericInserted:     ansi.StylePrimitive{Color: stringPtr("#9ece6a")},
			GenericEmph:         ansi.StylePrimitive{Italic: boolPtr(true)},
			GenericStrong:       ansi.StylePrimitive{Bold: boolPtr(true)},
			GenericSubheading:   ansi.StylePrimitive{Color: stringPtr("#7dcfff")},
			Background:         ansi.StylePrimitive{BackgroundColor: stringPtr("#1a1b26")},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#a9b1d6"),
			},
		},
		CenterSeparator: stringPtr("┼"),
		ColumnSeparator: stringPtr("│"),
		RowSeparator:    stringPtr("─"),
	},
	DefinitionTerm: ansi.StylePrimitive{
		Color: stringPtr("#7aa2f7"),
		Bold:  boolPtr(true),
	},
	DefinitionDescription: ansi.StylePrimitive{
		Color:       stringPtr("#a9b1d6"),
		BlockPrefix: "  ",
	},
}

var (
	hrLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3b4261"))

	// Heading background styles — gradient from bright to dim.
	headingStyles = map[string]lipgloss.Style{
		"H1": lipgloss.NewStyle().Background(lipgloss.Color("#7aa2f7")).Foreground(lipgloss.Color("#ffffff")).Bold(true),
		"H2": lipgloss.NewStyle().Background(lipgloss.Color("#3d59a1")).Foreground(lipgloss.Color("#c0caf5")).Bold(true),
		"H3": lipgloss.NewStyle().Background(lipgloss.Color("#2e3a5e")).Foreground(lipgloss.Color("#a9b1d6")).Bold(true),
		"H4": lipgloss.NewStyle().Background(lipgloss.Color("#283350")).Foreground(lipgloss.Color("#a9b1d6")).Bold(true),
		"H5": lipgloss.NewStyle().Background(lipgloss.Color("#24283b")).Foreground(lipgloss.Color("#7aa2f7")).Bold(true),
		"H6": lipgloss.NewStyle().Background(lipgloss.Color("#1f2335")).Foreground(lipgloss.Color("#565f89")).Bold(true),
	}
)

func renderMarkdown(content string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(mdStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(out, "\n")
}

// addTableRowSeparators is a no-op placeholder. Glamour renders tables
// with a header separator (┼/─) and tightly packed data rows by default.
func addTableRowSeparators(out string) string {
	return out
}

// postProcessMarkdown applies full-width headings and HRs.
// Called after mermaid splicing so the width scan sees all content.
func postProcessMarkdown(out string, minWidth int) string {
	fullWidth := minWidth
	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > fullWidth {
			fullWidth = w
		}
	}

	// Replace HR placeholders with full-width rules.
	hr := hrLineStyle.Render(strings.Repeat("─", fullWidth))
	out = strings.ReplaceAll(out, "HRPLACEHOLDER", hr)

	// Replace heading placeholders with full-width background bars.
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		for level := 1; level <= 6; level++ {
			tag := fmt.Sprintf("H%d", level)
			open := tag + "OPEN"
			close := tag + "CLOSE"
			if strings.Contains(line, open) && strings.Contains(line, close) {
				plain := xansi.Strip(line)
				plain = strings.Replace(plain, open, "", 1)
				plain = strings.Replace(plain, close, "", 1)
				plain = strings.TrimSpace(plain)
				pad := fullWidth - len(plain) - 2
				if pad < 0 {
					pad = 0
				}
				text := " " + plain + strings.Repeat(" ", pad+1)
				lines[i] = headingStyles[tag].Render(text)
				break
			}
		}
	}

	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func highlightContent(content, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" || strings.HasPrefix(filepath.Base(filename), ".") {
		return highlight(content, filename)
	}
	return highlight(content, filename)
}

func renderMermaid(content string) (string, error) {
	return mermaidCmd.RenderDiagram(content, diagram.DefaultConfig())
}

func renderMarkdownWithMermaid(content string, width int) string {
	// Render mermaid blocks and stash them behind placeholders so Glamour
	// doesn't word-wrap the ASCII art.
	rendered := map[string]string{}
	idx := 0
	stripped := mermaidBlockRe.ReplaceAllStringFunc(content, func(block string) string {
		matches := mermaidBlockRe.FindStringSubmatch(block)
		if len(matches) < 2 {
			return block
		}
		ascii, err := renderMermaid(matches[1])
		if err != nil {
			return block + "\n\n> Mermaid render error: " + err.Error()
		}
		key := fmt.Sprintf("MERMAIDPLACEHOLDER%d", idx)
		idx++
		rendered[key] = ascii
		return key
	})

	// Render the markdown (mermaid blocks are now just placeholder text).
	out := renderMarkdown(stripped, width)

	// Add table row separators BEFORE mermaid splicing — at this point
	// only Glamour table │ characters exist, no mermaid box-drawing.
	out = addTableRowSeparators(out)

	// Splice the raw ASCII diagrams back in.
	// Glamour wraps the placeholder in paragraph styling (leading spaces, etc.)
	// so we find the full line containing the placeholder and replace the whole line.
	for key, ascii := range rendered {
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			if strings.Contains(line, key) {
				lines[i] = ascii
				break
			}
		}
		out = strings.Join(lines, "\n")
	}

	// Post-process headings/HR after mermaid splicing so width calc sees everything.
	return postProcessMarkdown(out, width)
}
