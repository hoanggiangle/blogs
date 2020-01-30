// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: search/search.proto

package search

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogo/protobuf/types"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = types.DynamicAny{}
)

func errorField(fieldName, msg string) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldName,
		Description: msg,
	}
}

// Validate checks the field values on DeleteReq with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *DeleteReq) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Id

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// DeleteReqValidationError is the validation error returned by
// DeleteReq.Validate if the designated constraints aren't met.
type DeleteReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteReqValidationError) ErrorName() string { return "DeleteReqValidationError" }

// Error satisfies the builtin error interface
func (e DeleteReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteReqValidationError{}

// Validate checks the field values on DeleteRes with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *DeleteRes) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Success

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// DeleteResValidationError is the validation error returned by
// DeleteRes.Validate if the designated constraints aren't met.
type DeleteResValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteResValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteResValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteResValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteResValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteResValidationError) ErrorName() string { return "DeleteResValidationError" }

// Error satisfies the builtin error interface
func (e DeleteResValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteRes.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteResValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteResValidationError{}

// Validate checks the field values on Note with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *Note) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Id

	// no validation rules for Text

	if v, ok := interface{}(m.GetCreated()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errorFields = append(errorFields, errorField("Created", "embedded message failed validation"))
		}
	}

	if v, ok := interface{}(m.GetModified()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errorFields = append(errorFields, errorField("Modified", "embedded message failed validation"))
		}
	}

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NoteValidationError is the validation error returned by Note.Validate if the
// designated constraints aren't met.
type NoteValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NoteValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NoteValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NoteValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NoteValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NoteValidationError) ErrorName() string { return "NoteValidationError" }

// Error satisfies the builtin error interface
func (e NoteValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNote.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NoteValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NoteValidationError{}

// Validate checks the field values on Notes with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Notes) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Total

	for idx, item := range m.GetNotes() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				errorFields = append(errorFields, errorField(fmt.Sprintf("Notes[%v]", idx), "embedded message failed validation"))
			}
		}

	}

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NotesValidationError is the validation error returned by Notes.Validate if
// the designated constraints aren't met.
type NotesValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NotesValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NotesValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NotesValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NotesValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NotesValidationError) ErrorName() string { return "NotesValidationError" }

// Error satisfies the builtin error interface
func (e NotesValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNotes.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NotesValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NotesValidationError{}

// Validate checks the field values on NoteAddReq with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *NoteAddReq) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Text

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NoteAddReqValidationError is the validation error returned by
// NoteAddReq.Validate if the designated constraints aren't met.
type NoteAddReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NoteAddReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NoteAddReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NoteAddReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NoteAddReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NoteAddReqValidationError) ErrorName() string { return "NoteAddReqValidationError" }

// Error satisfies the builtin error interface
func (e NoteAddReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNoteAddReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NoteAddReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NoteAddReqValidationError{}

// Validate checks the field values on NoteListReq with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *NoteListReq) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	if v, ok := interface{}(m.GetPagination()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errorFields = append(errorFields, errorField("Pagination", "embedded message failed validation"))
		}
	}

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NoteListReqValidationError is the validation error returned by
// NoteListReq.Validate if the designated constraints aren't met.
type NoteListReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NoteListReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NoteListReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NoteListReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NoteListReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NoteListReqValidationError) ErrorName() string { return "NoteListReqValidationError" }

// Error satisfies the builtin error interface
func (e NoteListReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNoteListReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NoteListReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NoteListReqValidationError{}

// Validate checks the field values on NoteFilter with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *NoteFilter) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	// no validation rules for Type

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NoteFilterValidationError is the validation error returned by
// NoteFilter.Validate if the designated constraints aren't met.
type NoteFilterValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NoteFilterValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NoteFilterValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NoteFilterValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NoteFilterValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NoteFilterValidationError) ErrorName() string { return "NoteFilterValidationError" }

// Error satisfies the builtin error interface
func (e NoteFilterValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNoteFilter.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NoteFilterValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NoteFilterValidationError{}

// Validate checks the field values on NoteChangedEvent with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *NoteChangedEvent) Validate() error {
	if m == nil {
		return nil
	}

	errorFields := []*errdetails.BadRequest_FieldViolation{}

	if v, ok := interface{}(m.GetNote()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errorFields = append(errorFields, errorField("Note", "embedded message failed validation"))
		}
	}

	// no validation rules for Type

	if len(errorFields) > 0 {
		st := status.New(codes.InvalidArgument, "Invalid data")
		br := &errdetails.BadRequest{}
		br.FieldViolations = errorFields
		st, _ = st.WithDetails(br)
		return st.Err()
	}

	return nil
}

// NoteChangedEventValidationError is the validation error returned by
// NoteChangedEvent.Validate if the designated constraints aren't met.
type NoteChangedEventValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NoteChangedEventValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NoteChangedEventValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NoteChangedEventValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NoteChangedEventValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NoteChangedEventValidationError) ErrorName() string { return "NoteChangedEventValidationError" }

// Error satisfies the builtin error interface
func (e NoteChangedEventValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNoteChangedEvent.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NoteChangedEventValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NoteChangedEventValidationError{}
