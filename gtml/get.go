package gtml

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
)

func Get[TResult, TLink any](router *Router, getter func(TLink) (TResult, error)) error {
	getLinkType := reflect.TypeOf((*TLink)(nil)).Elem()
	if getLinkType.Kind() != reflect.Struct {
		return fmt.Errorf("gtlm: Route links must be structs, got: %s", getLinkType)
	}

	route := buildRoute(getLinkType)
	router.Mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		link, err := requestToLink[TLink](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("501 - ERROR: %s", err.Error())))

			return
		}

		getObj, err := getter(link)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("501 - ERROR: %s", err.Error())))

			return
		}

		if target := r.Header.Get("HX-Target"); strings.HasPrefix(target, "gtlm-search-") {
			html, err := getSearchDataOnly(router, getObj, target, r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("501 - ERROR: %s", err.Error())))

				return
			}

			w.Header().Set("HX-Push-Url", r.URL.String())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))

			return
		}

		body, err := router.html(getObj, "", r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("501 - ERROR: %s", err.Error())))

			return
		}

		html, err := executeTemplate("body.html", struct {
			Main template.HTML
		}{template.HTML(body)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("501 - ERROR: %s", err.Error())))

			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	})

	return nil
}

func getSearchDataOnly(router *Router, getObj any, target string, req *http.Request) (template.HTML, error) {
	targetFieldName := strings.Split(target, "-")[2]
	getValue := reflect.ValueOf(getObj)

	for i := range reflect.VisibleFields(getValue.Type()) {
		field := getValue.Type().Field(i).Name
		fieldValue := getValue.Field(i)

		if field == targetFieldName {
			searchLinkGetter, ok := fieldValue.Interface().(searchLinkGetter)
			if !ok {
				return template.HTML(""), fmt.Errorf("search link does not implement search link getter (type %s)", fieldValue.Type())
			}

			searchLink := updateSearchValues(searchLinkGetter.GetSearchLinkValue(), req)

			html, err := router.searchHtmlData(searchLink, req)
			if err != nil {
				return template.HTML(""), err
			}

			return html, nil
		}
	}

	return template.HTML(""), fmt.Errorf("failed to find search field %s in type %s", targetFieldName, getValue.Type())
}
