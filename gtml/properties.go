package gtml

import (
	"reflect"
	"strings"
)

func getGtmlProperties(field reflect.StructField) []string {
	propString := field.Tag.Get("gtml")
	props := strings.Split(propString, ",")

	return props
}
