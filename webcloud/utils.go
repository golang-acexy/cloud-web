package webcloud

import (
	"time"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/golang-acexy/cloud-database/databasecloud/rds"
)

// normalizeTimeRanges fills the default field and validates allowed names and interval bounds.
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

// ConvertTimeRanges validates names and bounds and returns independent RDS time pointers.
// At most one range may use the default field; invalid input returns ErrBadRequestParameters.
// Empty input returns an empty non-nil slice; input values are not modified.
func ConvertTimeRanges(timeRanges []TimeRange, allowedFields []string, defaultTimeRangeField string) ([]rds.TimeRange, error) {
	normalized, err := normalizeTimeRanges(timeRanges, allowedFields, defaultTimeRangeField)
	if err != nil {
		return nil, err
	}
	result := make([]rds.TimeRange, 0, len(normalized))
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
		result = append(result, rds.TimeRange{Field: timeRange.Field, StartTime: startTime, EndTime: endTime})
	}
	return result, nil
}
