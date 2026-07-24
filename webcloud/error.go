package webcloud

import "errors"

var (
	ErrUnauthorized                 = errors.New("unauthorized request")
	ErrBadRequestParameters         = errors.New("bad request parameters")
	ErrUnsupportedIDType            = errors.New("unsupported id type")
	ErrAuthorityFetchRequired       = errors.New("authority fetch required")
	ErrAuthorityStructFieldRequired = errors.New("authority struct field required")
	ErrAuthorityDataColumnRequired  = errors.New("authority data column required")
	ErrDTOTypeMustBeStruct          = errors.New("DTO type must be a struct")
	ErrEmbeddedFieldConflict        = errors.New("embedded field conflict")
)
