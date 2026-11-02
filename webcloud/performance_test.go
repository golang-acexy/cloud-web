package webcloud

import (
	"fmt"
	"testing"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/str"
)

var auditFieldResult map[string]any

// BenchmarkFieldValidation compares the previous algorithm with the production fused path.
func BenchmarkFieldValidation(b *testing.B) {
	for _, count := range []int{8, 64, 256} {
		allowed := make([]string, count)
		for i := range allowed {
			allowed[i] = fmt.Sprintf("field_%d", i)
		}
		input := make(map[string]any, 8)
		for i := 0; i < 8; i++ {
			input[fmt.Sprintf("field_%d", i)] = i
		}

		set := coll.SliceFilterToMap(allowed, func(s string) (string, struct{}, bool) { return s, struct{}{}, true })
		b.Run(fmt.Sprintf("previous/%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				keys := coll.MapFilterToSlice(input, func(k string, v any) (string, bool) { return str.CamelToSnake(k), true })
				if !coll.SliceIsSubset(keys, allowed) {
					b.Fatal("rejected")
				}
				auditFieldResult = coll.MapCollect(input, func(k string, v any) (string, any) { return str.CamelToSnake(k), v })
				if len(auditFieldResult) != 8 {
					b.Fatal("missing fields")
				}
			}
		})
		router := &BaseRouter[int, struct{}, struct{}, struct{}, struct{}]{queryAllowedColumns: set}
		b.Run(fmt.Sprintf("optimized/%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				output, err := router.convertFieldsToColumns(input, queryMode)
				if err != nil {
					b.Fatal(err)
				}
				auditFieldResult = output
				if len(auditFieldResult) != 8 {
					b.Fatal("missing fields")
				}
			}
		})
	}
}
