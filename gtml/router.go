package gtml

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

type Router struct {
	Mux       *http.ServeMux
	searchMap map[reflect.Type]func(data any) (any, error)
}

func NewRouter() *Router {
	return &Router{
		http.NewServeMux(),
		make(map[reflect.Type]func(data any) (any, error)),
	}
}

func buildRoute(linkType reflect.Type) string {
	if linkType.NumField() == 0 {
		return "/"
	}

	route := ""

	for i := 0; i < linkType.NumField(); i++ {
		field := linkType.Field(i)
		name := field.Name
		route += fmt.Sprintf("/%s/{%s}", url.PathEscape(name), name)
	}

	return route
}

func requestToLink[TLink any](r *http.Request) (TLink, error) {
	t := reflect.TypeOf((*TLink)(nil)).Elem()
	value := reflect.New(t).Elem()

	for field := range reflect.VisibleFields(value.Type()) {
		partValue := r.PathValue(value.Type().Field(field).Name)
		value.Field(field).Set(reflect.ValueOf(partValue))
	}

	return value.Interface().(TLink), nil
}

func linkToPath(link any) (string, error) {
	v := reflect.ValueOf(link)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("only structs can be made into paths")
	}

	path := ""

	for field := range reflect.VisibleFields(t) {
		name := t.Field(field).Name
		value := v.Field(field).String()
		path += fmt.Sprintf("/%s/%s", url.PathEscape(name), url.PathEscape(value))
	}

	return path, nil
}
