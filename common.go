package config_merger

import (
	"reflect"
	"strings"
)

func maskString(s string, showLastSymbols int) string {
	if len(s) <= showLastSymbols {
		return s
	}
	return strings.Repeat("*", len(s)-showLastSymbols) + s[len(s)-showLastSymbols:]
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetTagContents(s Source, tag string, field reflect.StructField) string {
	column := ""
	tagId := field.Tag.Get("tagId")
	if (tagId != "") && s.GetTagIds()[tagId] != "" {
		column = s.GetTagIds()[tagId]
	} else {
		column = field.Tag.Get(tag)
	}

	return column
}
