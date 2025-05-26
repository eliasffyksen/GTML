package gtml

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
)

type HTMLer interface {
	HTML() (template.HTML, error)
}

func linkToHtml(link any) template.HTML {
	href := ""

	t := reflect.TypeOf(link).Elem()
	v := reflect.ValueOf(link).Elem()
	for i := 0; i < v.NumField(); i++ {
		name := t.Field(i).Name
		value := v.Field(i).Interface()

		href += fmt.Sprintf("/%s/%s",
			name, url.PathEscape(fmt.Sprint(value)))
	}

	return template.HTML(fmt.Sprintf("<a href=\"%s\">%s</a>", href, template.HTMLEscaper(link)))
}

func (r *Router) html(data any, fieldName string, req *http.Request) (template.HTML, error) {
	if data == nil {
		return "N/A", nil
	}

	switch v := data.(type) {
	case int:
	case float64:
	case bool:
		return template.HTML(fmt.Sprint(v)), nil
	case string:
		return template.HTML(fmt.Sprintf("<p>%s</p>", template.HTMLEscapeString(v))), nil
	case searchLinkGetter:
		return r.searchHtml(fieldName, v.GetSearchLinkValue(), req)
	case HTMLer:
		return v.HTML()
	}

	value := reflect.ValueOf(data)
	switch value.Kind() {
	case reflect.Struct:
		return r.structHtml(value, req)
	case reflect.Slice:
		return sliceHtml(value)
	}

	return template.HTML(template.HTMLEscapeString(fmt.Sprint(data))), nil
}

func (r *Router) structHtml(value reflect.Value, req *http.Request) (template.HTML, error) {
	if value.Type().Kind() != reflect.Struct {
		return "", fmt.Errorf("gtml: structHtml called on non-struct value: %v", value)
	}

	if strings.Contains(value.Type().Name(), "Search") {
	}

	title := value.Type().Name()
	output := template.HTML(fmt.Sprintf("<h1>%s</h1>", template.HTMLEscapeString(title)))

	for field := range reflect.VisibleFields(value.Type()) {
		if !value.Type().Field(field).IsExported() {
			continue
		}

		name := value.Type().Field(field).Name
		value := value.Field(field).Interface()
		output += template.HTML(fmt.Sprintf("<h2>%s</h2>", template.HTMLEscapeString(name)))

		elemHtml, err := r.html(value, name, req)
		if err != nil {
			return "", err
		}

		output += elemHtml
	}

	return output, nil
}

func sliceHtml(value reflect.Value) (template.HTML, error) {
	if value.Kind() != reflect.Slice {
		return "", fmt.Errorf("sliceHtml called with non-slice kind %s", value.Kind())
	}

	rowType := value.Type().Elem()
	tableData := TableData{
		Headers: getSliceHeaders(rowType),
	}

	for i := 0; i < value.Len(); i++ {
		tableRow, err := getSliceTableRow(value.Index(i))
		if err != nil {
			return "", err
		}

		tableData.Rows = append(tableData.Rows, tableRow)
	}

	html, err := Table(tableData)
	if err != nil {
		return "", fmt.Errorf("failed to build table from slice: %w", err)
	}

	return html, nil
}

func getSliceTableRow(v reflect.Value) (TableRow, error) {
	t := v.Type()
	href, err := getHref(v)
	if err != nil {
		return TableRow{}, err
	}

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return TableRow{}, fmt.Errorf("table row must be either ptr or struct")
	}

	tableRow := TableRow{
		Href: href,
	}

	for _, field := range reflect.VisibleFields(t) {
		fieldV := v.FieldByIndex(field.Index)
		props := getGtmlProperties(field)
		if slices.Contains(props, "table-hide") {
			continue
		}

		fieldHref, err := getHref(fieldV)
		if err != nil {
			return TableRow{}, err
		}

		tableRow.Values = append(tableRow.Values, TableValue{
			fmt.Sprint(fieldV.Interface()),
			fieldHref,
		})
	}

	return tableRow, nil
}

func getHref(v reflect.Value) (*string, error) {
	linker, ok := v.Interface().(Linker)
	if !ok {
		return nil, nil
	}

	value, err := linkToPath(linker.Link())
	if err != nil {
		return nil, err
	}

	return &value, nil
}

func getSliceHeaders(t reflect.Type) []string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	headers := make([]string, 0)
	for _, field := range reflect.VisibleFields(t) {
		props := getGtmlProperties(field)
		if slices.Contains(props, "table-hide") {
			continue
		}

		headers = append(headers, field.Name)
	}

	return headers
}
