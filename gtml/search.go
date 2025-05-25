package gtml

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
)

type SearchLink[T any] struct {
	value T
}

type searchLinkGetter interface {
	GetSearchLinkValue() any
}

var _ searchLinkGetter = SearchLink[string]{}

func NewSearchLink[T any](data T) SearchLink[T] {
	return SearchLink[T]{
		data,
	}
}

func (sl SearchLink[TLink]) GetSearchLinkValue() any {
	return sl.value
}

func Search[TCriteria, TResult any](router *Router, searcher func(link TCriteria) ([]TResult, error)) error {
	linkType := reflect.TypeOf((*TCriteria)(nil)).Elem()
	if linkType.Kind() != reflect.Struct {
		return fmt.Errorf("gtlm: Search links must be structs, got: %s", linkType)
	}

	if _, ok := router.searchMap[linkType]; ok {
		return fmt.Errorf("gtlm: search already registered for type %s", linkType)
	}

	router.searchMap[linkType] = func(data any) (any, error) {
		typedData, ok := data.(TCriteria)
		if !ok {
			return nil, fmt.Errorf("gtlm: could not cast data of type %s into %s", reflect.TypeOf(data), reflect.TypeFor[TCriteria]())
		}

		return searcher(typedData)
	}

	return nil
}

func (r *Router) searchHtml(fieldName string, searchLink any, req *http.Request) (template.HTML, error) {
	if fieldName == "" {
		return template.HTML(""), fmt.Errorf("search recieved empty field name for type %s", reflect.TypeOf(searchLink))
	}

	searchLink = updateSearchValues(searchLink, req)

	linkValue := reflect.ValueOf(searchLink)
	targetId := "gtlm-search-" + fieldName
	target := "#" + targetId

	html, err := searchInputs(linkValue, target)
	if err != nil {
		return template.HTML(""), err
	}

	dataHtml, err := r.searchHtmlData(searchLink, req)
	if err != nil {
		return template.HTML(""), err
	}

	html += template.HTML(fmt.Sprintf("<div id=\"%s\">%s</div>", targetId, dataHtml))

	return html, nil
}

func updateSearchValues(searchLink any, req *http.Request) any {
	searchLinkType := reflect.TypeOf(searchLink)
	oldLinkValue := reflect.ValueOf(searchLink)
	newLinkValue := reflect.New(searchLinkType).Elem()

	for _, field := range reflect.VisibleFields(searchLinkType) {
		newFieldValue := newLinkValue.FieldByIndex(field.Index)
		oldFieldValue := oldLinkValue.FieldByIndex(field.Index)

		paramName := searchLinkType.Name() + "." + field.Name

		if req.URL.Query().Has(paramName) {
			newFieldValue.SetString(req.URL.Query().Get(paramName))
		} else {
			newFieldValue.Set(oldFieldValue)
		}
	}

	return newLinkValue.Interface()
}

func (r *Router) searchHtmlData(searchLink any, req *http.Request) (template.HTML, error) {
	searchLinkType := reflect.TypeOf(searchLink)

	searchGetter, ok := r.searchMap[searchLinkType]
	if !ok {
		return template.HTML(""), fmt.Errorf("failed to find search getter for type %s", searchLinkType)
	}

	data, err := searchGetter(searchLink)
	if err != nil {
		return template.HTML(""), fmt.Errorf("failed to execute search getter: %w", err)
	}

	return r.html(data, "", req)
}

func searchInputs(link reflect.Value, target string) (template.HTML, error) {
	html := template.HTML("")

	for i := range reflect.VisibleFields(link.Type()) {
		field := link.Type().Field(i)
		fieldValue := link.Field(i)
		tag := field.Tag
		gtlmTag := tag.Get("gtlm")

		if gtlmTag == "search" {
			descriptor := fmt.Sprintf("%s.%s", link.Type().Name(), field.Name)

			fieldHtml, err := searchInput(descriptor, field.Name, fieldValue, target)
			if err != nil {
				return template.HTML(""), err
			}

			html += fieldHtml
		}
	}

	html = template.HTML(fmt.Sprintf("<form hx-get=\"/\" hx-target=\"%s\">%s</form>", target, html))

	return html, nil
}

func searchInput(descriptor string, fieldName string, value reflect.Value, target string) (template.HTML, error) {
	if value.Kind() == reflect.String {
		return executeTemplate("string_input.html", struct {
			Name        string
			Value       string
			Placeholder string
			Target      string
		}{
			descriptor,
			value.String(),
			fieldName,
			target,
		})
	}

	return template.HTML(""), fmt.Errorf("input html to implemented for kind %s", value)
}
