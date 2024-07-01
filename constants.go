// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"fmt"
	"net/http"

	"entgo.io/ent/entc/gen"
	"github.com/go-openapi/inflect"
	"github.com/ogen-go/ogen"
	"github.com/stoewer/go-strcase"
)

const OpenAPIVersion = "3.0.3"

var (
	// Add all casing and word-massaging functions here so others can use them if they
	// want to customize the naming of their spec/endpoints/etc.
	Pluralize   = memoize(inflect.Pluralize) // nolint: errcheck,unused
	KebabCase   = memoize(strcase.KebabCase)
	Singularize = memoize(gen.Funcs["singular"].(func(string) string)) // nolint: errcheck,unused
	PascalCase  = memoize(gen.Funcs["pascal"].(func(string) string))   // nolint: errcheck,unused
	CamelCase   = memoize(gen.Funcs["camel"].(func(string) string))    // nolint: errcheck,unused
	SnakeCase   = memoize(gen.Funcs["snake"].(func(string) string))    // nolint: errcheck,unused
)

// Operation represents the CRUD operation(s).
type Operation string

const (
	// OperationCreate represents the create operation (method: POST).
	OperationCreate Operation = "create"
	// OperationRead represents the read operation (method: GET).
	OperationRead Operation = "read"
	// OperationUpdate represents the update operation (method: PATCH).
	OperationUpdate Operation = "update"
	// OperationDelete represents the delete operation (method: DELETE).
	OperationDelete Operation = "delete"
	// OperationList represents the list operation (method: GET).
	OperationList Operation = "list"
)

// AllOperations holds a list of all supported operations.
var AllOperations = []Operation{OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList}

// Predicate represents a filtering predicate provided by ent.
type Predicate int

// Mirrored from entgo.io/ent/entc/gen with special groupings and support for bitwise operations.
const (
	// FilterEdge is a special filter which is applied to the edge itself, indicating
	// that all of the edges fields should also be included in filtering options.
	FilterEdge Predicate = 1 << iota

	FilterEQ           // =
	FilterNEQ          // <>
	FilterGT           // >
	FilterGTE          // >=
	FilterLT           // <
	FilterLTE          // <=
	FilterIsNil        // IS NULL / has
	FilterNotNil       // IS NOT NULL / hasNot
	FilterIn           // within
	FilterNotIn        // without
	FilterEqualFold    // equals case-insensitive
	FilterContains     // containing
	FilterContainsFold // containing case-insensitive
	FilterHasPrefix    // startingWith
	FilterHasSuffix    // endingWith

	// FilterGroupEqualExact includes: eq, neq, equal fold, is nil.
	FilterGroupEqualExact = FilterEQ | FilterNEQ | FilterEqualFold | FilterGroupNil
	// FilterGroupEqual includes: eq, neq, equal fold, contains, contains case, prefix, suffix, nil.
	FilterGroupEqual = FilterGroupEqualExact | FilterGroupContains | FilterHasPrefix | FilterHasSuffix
	// FilterGroupContains includes: contains, contains case, is nil.
	FilterGroupContains = FilterContains | FilterContainsFold | FilterGroupNil
	// FilterGroupNil includes: is nil.
	FilterGroupNil = FilterIsNil
	// FilterGroupLength includes: gt, lt (often gte/lte isn't really needed).
	FilterGroupLength = FilterGT | FilterLT
	// FilterGroupArray includes: in, not in.
	FilterGroupArray = FilterIn | FilterNotIn
)

// filterMap maps a predicate to the entgo.io/ent/entc/gen.Op (to get string representation).
var filterMap = map[Predicate]gen.Op{
	FilterEQ:           gen.EQ,
	FilterNEQ:          gen.NEQ,
	FilterGT:           gen.GT,
	FilterGTE:          gen.GTE,
	FilterLT:           gen.LT,
	FilterLTE:          gen.LTE,
	FilterIsNil:        gen.IsNil,
	FilterNotNil:       gen.NotNil,
	FilterIn:           gen.In,
	FilterNotIn:        gen.NotIn,
	FilterEqualFold:    gen.EqualFold,
	FilterContains:     gen.Contains,
	FilterContainsFold: gen.ContainsFold,
	FilterHasPrefix:    gen.HasPrefix,
	FilterHasSuffix:    gen.HasSuffix,
}

// String returns the gen.Op string representation of a predicate.
func (p Predicate) String() string {
	if _, ok := filterMap[p]; ok {
		return filterMap[p].Name()
	}
	panic("predicate.String() called with grouped predicate, use Explode() first")
}

// Has returns if the predicate has the provided predicate.
func (p Predicate) Has(v Predicate) bool {
	return p&v != 0
}

// Add adds the provided predicate to the current predicate.
func (p Predicate) Add(v Predicate) Predicate {
	p |= v
	return p
}

// Remove removes the provided predicate from the current predicate.
func (p Predicate) Remove(v Predicate) Predicate {
	p &^= v
	return p
}

// Explode returns all individual predicates as []gen.Op.
func (p Predicate) Explode() (ops []gen.Op) {
	for pred, op := range filterMap {
		if p.Has(pred) {
			ops = append(ops, op)
		}
	}
	return ops
}

// predicateFormat returns the query string representation of a filter predicate.
func predicateFormat(op gen.Op) string {
	switch op {
	case gen.Contains:
		return "has"
	case gen.ContainsFold:
		return "ihas"
	case gen.EqualFold:
		return "ieq"
	case gen.HasPrefix:
		return "prefix"
	case gen.HasSuffix:
		return "suffix"
	case gen.IsNil:
		return "null"
	default:
		return CamelCase(SnakeCase(op.Name()))
	}
}

// predicateDescription returns the description of a filter predicate.
func predicateDescription(f *gen.Field, op gen.Op) string {
	switch op {
	case gen.EQ:
		return fmt.Sprintf("Filters field %q to be equal to the provided value.", f.Name)
	case gen.NEQ:
		return fmt.Sprintf("Filters field %q to be not equal to the provided value.", f.Name)
	case gen.GT:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be longer than the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be greater than the provided value.", f.Name)
	case gen.GTE:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be longer than or equal in length to the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be greater than or equal to the provided value.", f.Name)
	case gen.LT:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be shorter than the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be less than the provided value.", f.Name)
	case gen.LTE:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be shorter than or equal in length to the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be less than or equal to the provided value.", f.Name)
	case gen.IsNil:
		return fmt.Sprintf("Filters field %q to be null/nil.", f.Name)
	case gen.NotNil:
		return fmt.Sprintf("Filters field %q to be not null/nil.", f.Name)
	case gen.In:
		return fmt.Sprintf("Filters field %q to be within the provided values.", f.Name)
	case gen.NotIn:
		return fmt.Sprintf("Filters field %q to be not within the provided values.", f.Name)
	case gen.EqualFold:
		return fmt.Sprintf("Filters field %q to be equal to the provided value, case-insensitive.", f.Name)
	case gen.Contains:
		return fmt.Sprintf("Filters field %q to contain the provided value.", f.Name)
	case gen.ContainsFold:
		return fmt.Sprintf("Filters field %q to contain the provided value, case-insensitive.", f.Name)
	case gen.HasPrefix:
		return fmt.Sprintf("Filters field %q to start with the provided value.", f.Name)
	case gen.HasSuffix:
		return fmt.Sprintf("Filters field %q to end with the provided value.", f.Name)
	default:
		panic("unknown predicate")
	}
}

const (
	defaultMinItemsPerPage = 1
	defaultMaxItemsPerPage = 100
	defaultItemsPerPage    = 10
)

// HTTPHandler represents the HTTP handler to use for the HTTP server implementation.
type HTTPHandler string

const (
	// HandlerNone disables all code generation for the HTTP server implementation.
	HandlerNone HTTPHandler = ""
	// HandlerStdlib uses a net/http servemux, along with the Go 1.22 advanced
	// path matching functionality to map methods and URL parameters (e.g. {id}).
	// Technically, this can be used with many other stdlib-compatible routers,
	// as long as they support [http.Request.PathValue] for path parameters.
	HandlerStdlib HTTPHandler = "stdlib"
	// HandlerChi uses the chi router, mounting each endpoint individually using
	// the provided router. Note that you must use at least chi v5.0.12 or newer,
	// which supports populating the requests path values, and accessing them via
	// [http.Request.PathValue].
	HandlerChi HTTPHandler = "chi"
)

// AllSupportedHTTPHandlers is a list of all supported HTTP handlers.
var AllSupportedHTTPHandlers = []HTTPHandler{
	HandlerNone,
	HandlerStdlib,
	HandlerChi,
}

type RequestHeaders map[string]*ogen.Parameter

// Append merges the provided request headers into the current request headers, returning
// a new request headers map.
func (r RequestHeaders) Append(toMerge ...RequestHeaders) RequestHeaders {
	out := RequestHeaders{}
	for k, v := range r {
		out[k] = v
	}
	for _, m := range toMerge {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

type ResponseHeaders map[string]*ogen.Header

// Append merges the provided response headers into the current response headers, returning
// a new response headers map.
func (r ResponseHeaders) Append(toMerge ...ResponseHeaders) ResponseHeaders {
	out := ResponseHeaders{}
	for k, v := range r {
		out[k] = v
	}
	for _, m := range toMerge {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

type ErrorResponses map[int]*ogen.Schema

// Append merges the provided error responses into the current error responses, returning
// a new error responses map.
func (r ErrorResponses) Append(toMerge ...ErrorResponses) ErrorResponses {
	out := ErrorResponses{}
	for k, v := range r {
		out[k] = v
	}
	for _, m := range toMerge {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

var (
	// RateLimitHeaders are standardized rate limit response headers.
	RateLimitHeaders = ResponseHeaders{
		"X-Ratelimit-Limit": {
			Description: "The maximum number of requests that the consumer is permitted to make in a given period.",
			Required:    true,
			Schema:      &ogen.Schema{Type: "integer"},
		},
		"X-Ratelimit-Remaining": {
			Description: "The number of requests remaining in the current rate limit window.",
			Required:    true,
			Schema:      &ogen.Schema{Type: "integer"},
		},
		"X-Ratelimit-Reset": {
			Description: "The time at which the current rate limit window resets in UTC epoch seconds.",
			Required:    true,
			Schema:      &ogen.Schema{Type: "integer"},
		},
	}

	// RequestIDHeader is a standardized request ID request header.
	RequestIDHeader = RequestHeaders{
		"X-Request-Id": {
			Description: "A unique identifier for the request.",
			Required:    false,
			Schema:      &ogen.Schema{Type: "string"},
		},
	}

	// DefaultErrorResponses are the default error responses for the HTTP status codes,
	// which includes 400, 401, 403, 404, 409, 429, and 500.
	DefaultErrorResponses = ErrorResponses{
		http.StatusBadRequest:          ErrorResponseObject(http.StatusBadRequest),
		http.StatusUnauthorized:        ErrorResponseObject(http.StatusUnauthorized),
		http.StatusForbidden:           ErrorResponseObject(http.StatusForbidden),
		http.StatusNotFound:            ErrorResponseObject(http.StatusNotFound),
		http.StatusConflict:            ErrorResponseObject(http.StatusConflict),
		http.StatusTooManyRequests:     ErrorResponseObject(http.StatusTooManyRequests),
		http.StatusInternalServerError: ErrorResponseObject(http.StatusInternalServerError),
	}
)
