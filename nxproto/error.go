package nxproto

import (
	"errors"

	"connectrpc.com/connect"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
	"google.golang.org/protobuf/proto"
)

// Error is a business error carrying a Connect status code and an
// ErrorDetail payload (error_code + error_name + metadata).
// Service layers return *Error; handler layers convert with ToConnect.
type Error struct {
	ConnectCode connect.Code
	Code        int32
	Name        string
	Metadata    map[string]string
}

func (e *Error) Error() string { return e.Name }

// NewError creates a business error.
func NewError(connectCode connect.Code, code int32, name string) *Error {
	return &Error{ConnectCode: connectCode, Code: code, Name: name}
}

// WithMeta returns a shallow copy of the error with the given metadata
// key-value pair added. Safe to call on package-level sentinel errors.
func (e *Error) WithMeta(key, value string) *Error {
	cp := *e
	cp.Metadata = make(map[string]string, len(e.Metadata)+1)
	for k, v := range e.Metadata {
		cp.Metadata[k] = v
	}
	cp.Metadata[key] = value
	return &cp
}

// --- Server side: construct Connect errors ---

// ToConnect converts any error to a *connect.Error.
//   - *Error → connect.Error with ErrorDetail in details.
//   - other  → connect.Error with CodeInternal, no details.
func ToConnect(err error) *connect.Error {
	if err == nil {
		return nil
	}
	var xe *Error
	if !errors.As(err, &xe) {
		return connect.NewError(connect.CodeInternal, err)
	}
	ce := connect.NewError(xe.ConnectCode, err)
	detail, detailErr := connect.NewErrorDetail(errorToProto(xe))
	if detailErr == nil {
		ce.AddDetail(detail)
	}
	return ce
}

func errorToProto(e *Error) proto.Message {
	return &sharedv1.ErrorDetail{
		ErrorCode: e.Code,
		ErrorName: e.Name,
		Metadata:  e.Metadata,
	}
}

// --- Client side: parse Connect errors ---

// ErrorDetail holds the parsed business error from a Connect error response.
type ErrorDetail struct {
	ConnectCode connect.Code
	Code        int32
	Name        string
	Metadata    map[string]string
}

// ParseError extracts the business ErrorDetail from a Connect error.
// Returns nil if the error is not a Connect error or has no ErrorDetail.
func ParseError(err error) *ErrorDetail {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if !errors.As(err, &ce) {
		return nil
	}
	for _, d := range ce.Details() {
		msg, unmarshalErr := d.Value()
		if unmarshalErr != nil {
			continue
		}
		if ed, ok := msg.(*sharedv1.ErrorDetail); ok {
			return &ErrorDetail{
				ConnectCode: ce.Code(),
				Code:        ed.GetErrorCode(),
				Name:        ed.GetErrorName(),
				Metadata:    ed.GetMetadata(),
			}
		}
	}
	return nil
}

// IsErrorCode checks if an error contains a specific business error code.
func IsErrorCode(err error, code int32) bool {
	ed := ParseError(err)
	return ed != nil && ed.Code == code
}
