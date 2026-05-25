package logic

import (
	"fmt"
	"strings"
)

func mapToInterface(m map[string]string) []interface{} {
	result := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

func buildFilterExpr(filter map[string]string) string {
	if len(filter) == 0 {
		return ""
	}
	exprs := make([]string, 0, len(filter))
	for k, v := range filter {
		exprs = append(exprs, fmt.Sprintf(`metadata["%s"] == "%s"`, k, v))
	}
	var sb strings.Builder
	for i, e := range exprs {
		if i > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(e)
	}
	return sb.String()
}

func joinInt64s(ids []int64) string {
	var sb strings.Builder
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d", id))
	}
	return sb.String()
}
