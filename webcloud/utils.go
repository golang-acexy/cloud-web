package webcloud

import (
	"time"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/golang-acexy/starter-gorm/gormstarter"
)

// normalizeTimeRanges 补全默认时间字段，并统一校验时间字段白名单和查询范围。
func normalizeTimeRanges(timeRanges []TimeRange, allowedFields []string, defaultTimeRangeField string) ([]TimeRange, error) {
	if len(timeRanges) == 0 {
		return nil, nil
	}
	defaultField := str.CamelToSnake(defaultTimeRangeField)
	allowedFields = coll.SliceCollect(allowedFields, func(field string) string {
		return str.CamelToSnake(field)
	})
	result := make([]TimeRange, 0, len(timeRanges))
	defaultFieldUsed := false
	for _, timeRange := range timeRanges {
		field := str.CamelToSnake(timeRange.Field)
		if field == "" {
			if defaultFieldUsed {
				return nil, ErrBadRequestParameters
			}
			defaultFieldUsed = true
			field = defaultField
		}
		if field == "" || !coll.SliceContains(allowedFields, field) {
			return nil, ErrBadRequestParameters
		}
		if timeRange.Start == nil && timeRange.End == nil {
			return nil, ErrBadRequestParameters
		}
		if timeRange.Start != nil && timeRange.End != nil && !timeRange.Start.Before(timeRange.End.Time) {
			return nil, ErrBadRequestParameters
		}
		timeRange.Field = field
		result = append(result, timeRange)
	}
	return result, nil
}

// ConvertTimeRanges 校验时间字段白名单，并转换为 GORM 分页查询参数。
func ConvertTimeRanges(timeRanges []TimeRange, allowedFields []string, defaultTimeRangeField string) ([]gormstarter.TimeRange, error) {
	normalized, err := normalizeTimeRanges(timeRanges, allowedFields, defaultTimeRangeField)
	if err != nil {
		return nil, err
	}
	result := make([]gormstarter.TimeRange, 0, len(normalized))
	for _, timeRange := range normalized {
		var startTime, endTime *time.Time
		if timeRange.Start != nil {
			value := timeRange.Start.Time
			startTime = &value
		}
		if timeRange.End != nil {
			value := timeRange.End.Time
			endTime = &value
		}
		result = append(result, gormstarter.TimeRange{Field: timeRange.Field, StartTime: startTime, EndTime: endTime})
	}
	return result, nil
}
