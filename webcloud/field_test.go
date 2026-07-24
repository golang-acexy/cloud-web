package webcloud

import (
	"errors"
	"reflect"
	"testing"
)

type EmbeddedBase struct {
	CreatedAt int64
	Name      string
}

type EmbeddedMiddle struct {
	*EmbeddedBase
	Age uint
}

type embeddedDTO struct {
	EmbeddedMiddle
	Name   string
	UserID uint64
}

func TestCollectStructFieldNames(t *testing.T) {
	fields, err := collectStructFieldNames(embeddedDTO{})
	if err != nil {
		t.Fatalf("解析匿名字段失败: %v", err)
	}
	expected := []string{"Name", "UserID", "Age", "CreatedAt"}
	if !reflect.DeepEqual(fields, expected) {
		t.Fatalf("匿名字段展开结果不正确: actual=%v expected=%v", fields, expected)
	}
	if _, err = collectStructFieldNames(&embeddedDTO{}); !errors.Is(err, ErrDTOTypeMustBeStruct) {
		t.Fatalf("顶层指针 DTO 应返回 ErrDTOTypeMustBeStruct: %v", err)
	}
}

func TestCollectStructFieldNamesConflict(t *testing.T) {
	type First struct{ Name string }
	type Second struct{ Name string }
	type fieldConflict struct {
		First
		Second
	}
	if _, err := collectStructFieldNames(fieldConflict{}); !errors.Is(err, ErrEmbeddedFieldConflict) {
		t.Fatalf("同层字段冲突应返回 ErrEmbeddedFieldConflict: %v", err)
	}

	type columnConflict struct {
		UserID uint64
		UserId uint64
	}
	if _, err := collectStructFieldNames(columnConflict{}); !errors.Is(err, ErrEmbeddedFieldConflict) {
		t.Fatalf("数据库列冲突应返回 ErrEmbeddedFieldConflict: %v", err)
	}
}

func TestConvertStringToNamedID(t *testing.T) {
	type userID int64
	type tenantID string

	id, err := ConvertStringToID[userID]("123")
	if err != nil || id != 123 {
		t.Fatalf("命名整数 ID 转换失败: id=%v err=%v", id, err)
	}
	tenant, err := ConvertStringToID[tenantID]("tenant-1")
	if err != nil || tenant != "tenant-1" {
		t.Fatalf("命名字符串 ID 转换失败: id=%v err=%v", tenant, err)
	}
	if _, err = ConvertStringToID[userID]("invalid"); err == nil {
		t.Fatal("非法整数 ID 应返回错误")
	}
}
