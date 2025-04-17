// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"maps"
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
	maps.Copy(out, r)
	for _, m := range toMerge {
		maps.Copy(out, m)
	}
	return out
}

type ResponseHeaders map[string]*ogen.Header

// Append merges the provided response headers into the current response headers, returning
// a new response headers map.
func (r ResponseHeaders) Append(toMerge ...ResponseHeaders) ResponseHeaders {
	out := ResponseHeaders{}
	maps.Copy(out, r)
	for _, m := range toMerge {
		maps.Copy(out, m)
	}
	return out
}

type ErrorResponses map[int]*ogen.Schema

// Append merges the provided error responses into the current error responses, returning
// a new error responses map.
func (r ErrorResponses) Append(toMerge ...ErrorResponses) ErrorResponses {
	out := ErrorResponses{}
	maps.Copy(out, r)
	for _, m := range toMerge {
		maps.Copy(out, m)
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

// SchemaObjectAny can be used to define an object which may contain any properties.
var SchemaObjectAny = &ogen.Schema{
	Type: "object",
	AdditionalProperties: &ogen.AdditionalProperties{
		Bool: ptr(true), // https://github.com/ogen-go/ogen/issues/1221
	},
}
