package main

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	style     = styles.Get("monokai")
	formatter = formatters.Get("terminal256")
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

func highlightContent(content, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" || strings.HasPrefix(filepath.Base(filename), ".") {
		return highlight(content, filename)
	}
	return highlight(content, filename)
}
