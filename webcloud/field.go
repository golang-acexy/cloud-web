package webcloud

import (
	"fmt"
	"reflect"

	"github.com/acexy/golang-toolkit/util/str"
)

type embeddedStruct struct {
	typ       reflect.Type
	ancestors map[reflect.Type]struct{}
}

type structField struct {
	name   string
	column string
}

func structFieldToColumn(field string) string {
	if field == "ID" || field == "Id" {
		return "id"
	}
	return str.CamelToSnake(str.LowFirstChar(field))
}

// collectStructFieldNames 按 Go 字段提升规则展开匿名结构体，并返回有效字段名。
func collectStructFieldNames(value any) ([]string, error) {
	typ := reflect.TypeOf(value)
	if typ == nil || typ.Kind() != reflect.Struct {
		return nil, ErrDTOTypeMustBeStruct
	}

	queue := []embeddedStruct{{
		typ:       typ,
		ancestors: map[reflect.Type]struct{}{typ: {}},
	}}
	resolvedNames := make(map[string]struct{})
	resolvedColumns := make(map[string]struct{})
	result := make([]string, 0, typ.NumField())

	for len(queue) > 0 {
		currentLevel := queue
		queue = nil
		levelFields := make([]structField, 0)
		levelNames := make(map[string]struct{})
		levelColumns := make(map[string]struct{})

		for _, current := range currentLevel {
			for i := 0; i < current.typ.NumField(); i++ {
				field := current.typ.Field(i)
				if field.PkgPath != "" {
					continue
				}

				fieldType := field.Type
				if field.Anonymous {
					if fieldType.Kind() == reflect.Ptr {
						fieldType = fieldType.Elem()
					}
					if fieldType.Kind() == reflect.Struct {
						if _, cyclic := current.ancestors[fieldType]; !cyclic {
							ancestors := make(map[reflect.Type]struct{}, len(current.ancestors)+1)
							for ancestor := range current.ancestors {
								ancestors[ancestor] = struct{}{}
							}
							ancestors[fieldType] = struct{}{}
							queue = append(queue, embeddedStruct{typ: fieldType, ancestors: ancestors})
						}
						continue
					}
				}

				column := structFieldToColumn(field.Name)
				// 更浅层字段已经生效时，忽略被遮蔽的嵌入字段。
				if _, exists := resolvedNames[field.Name]; exists {
					continue
				}
				if _, exists := resolvedColumns[column]; exists {
					continue
				}
				if _, exists := levelNames[field.Name]; exists {
					return nil, fmt.Errorf("%w: field %s", ErrEmbeddedFieldConflict, field.Name)
				}
				if _, exists := levelColumns[column]; exists {
					return nil, fmt.Errorf("%w: column %s", ErrEmbeddedFieldConflict, column)
				}
				levelNames[field.Name] = struct{}{}
				levelColumns[column] = struct{}{}
				levelFields = append(levelFields, structField{name: field.Name, column: column})
			}
		}

		for _, field := range levelFields {
			if _, exists := resolvedNames[field.name]; exists {
				continue
			}
			if _, exists := resolvedColumns[field.column]; exists {
				continue
			}
			resolvedNames[field.name] = struct{}{}
			resolvedColumns[field.column] = struct{}{}
			result = append(result, field.name)
		}
	}

	return result, nil
}
