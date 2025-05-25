package gtml

import (
	"fmt"
	"html/template"
)

type TableValue struct {
	Value string
	Href  *string
}

type TableRow struct {
	Values []TableValue
	Href   *string
}

type TableData struct {
	Headers []string
	Rows    []TableRow
}

func Table(data TableData) (template.HTML, error) {
	html, err := executeTemplate("table.html", data)
	if err != nil {
		return "", fmt.Errorf("gtml: failed to build table: %w", err)
	}

	return template.HTML(html), nil
}
