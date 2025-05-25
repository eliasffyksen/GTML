package gtml

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates
var templateFS embed.FS
var templates *template.Template
var templateFuncMap template.FuncMap = template.FuncMap{
	"tableEntry": tableEntry,
}

func tableEntry(data any) any {
	if linker, ok := data.(Linker); ok {
		return linkToHtml(linker.Link())
	}

	return fmt.Sprint(data)
}

func executeTemplate(name string, data any) (template.HTML, error) {
	if templates == nil {
		var err error

		templates, err = template.New("").
			Funcs(templateFuncMap).
			ParseFS(templateFS, "templates/*.html")
		if err != nil {
			return "", err
		}
	}

	buffer := bytes.NewBufferString("")
	if err := templates.ExecuteTemplate(buffer, name, data); err != nil {
		return "", err
	}

	return template.HTML(buffer.String()), nil
}
