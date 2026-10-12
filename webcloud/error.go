package webcloud

import "errors"

var (
	// ErrUnauthorized indicates a missing required identity.
	ErrUnauthorized = errors.New("unauthorized request")
	// ErrBadRequestParameters indicates invalid condition keys or time-range parameters.
	ErrBadRequestParameters = errors.New("bad request parameters")
	// ErrUnsupportedIDType is the fallback for unsupported reflected ID kinds; IDType normally excludes them.
	ErrUnsupportedIDType = errors.New("unsupported id type")
	// ErrAuthorityFetchRequired indicates a missing authority callback when lookup is requested.
	ErrAuthorityFetchRequired = errors.New("authority fetch required")
	// ErrAuthorityStructFieldRequired indicates a missing save DTO authority field name.
	ErrAuthorityStructFieldRequired = errors.New("authority struct field required")
	// ErrAuthorityDataColumnRequired indicates a missing persistence authority column.
	ErrAuthorityDataColumnRequired = errors.New("authority data column required")
	// ErrDTOTypeMustBeStruct indicates a query or modify DTO is not a non-pointer struct.
	ErrDTOTypeMustBeStruct = errors.New("DTO type must be a struct")
	// ErrEmbeddedFieldConflict indicates same-depth field names or normalized columns conflict.
	ErrEmbeddedFieldConflict = errors.New("embedded field conflict")
)
