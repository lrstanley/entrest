// Code generated by ent, DO NOT EDIT.

package rest

import (
	"bytes"
	_ "embed"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	uuid "github.com/google/uuid"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/category"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/friendship"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/pet"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/post"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/privacy"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/settings"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/user"
)

//go:embed openapi.json
var OpenAPI []byte // OpenAPI contains the JSON schema of the API.

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

// ErrorResponse is the response structure for errors.
type ErrorResponse struct {
	Error     string `json:"error"`                // The underlying error, which may be masked when debugging is disabled.
	Type      string `json:"type"`                 // A summary of the error code based off the HTTP status code or application error code.
	Code      int    `json:"code"`                 // The HTTP status code or other internal application error code.
	RequestID string `json:"request_id,omitempty"` // The unique request ID for this error.
	Timestamp string `json:"timestamp,omitempty"`  // The timestamp of the error, in RFC3339 format.
}

type ErrBadRequest struct {
	Err error
}

func (e ErrBadRequest) Error() string {
	return fmt.Sprintf("bad request: %s", e.Err)
}

func (e ErrBadRequest) Unwrap() error {
	return e.Err
}

// IsBadRequest returns true if the unwrapped/underlying error is of type ErrBadRequest.
func IsBadRequest(err error) bool {
	var target *ErrBadRequest
	return errors.As(err, &target)
}

var ErrEndpointNotFound = errors.New("endpoint not found")

// IsEndpointNotFound returns true if the unwrapped/underlying error is of type ErrEndpointNotFound.
func IsEndpointNotFound(err error) bool {
	return errors.Is(err, ErrEndpointNotFound)
}

var ErrMethodNotAllowed = errors.New("method not allowed")

// IsMethodNotAllowed returns true if the unwrapped/underlying error is of type ErrMethodNotAllowed.
func IsMethodNotAllowed(err error) bool {
	return errors.Is(err, ErrMethodNotAllowed)
}

type ErrInvalidID struct {
	ID  string
	Err error
}

func (e ErrInvalidID) Error() string {
	return fmt.Sprintf("invalid ID provided: %q: %v", e.ID, e.Err)
}

func (e ErrInvalidID) Unwrap() error {
	return e.Err
}

// IsInvalidID returns true if the unwrapped/underlying error is of type ErrInvalidID.
func IsInvalidID(err error) bool {
	var target *ErrInvalidID
	return errors.As(err, &target)
}

// JSON marshals 'v' to JSON, and setting the Content-Type as application/json.
// Note that this does NOT auto-escape HTML. If 'v' cannot be marshalled to JSON,
// this will panic.
//
// JSON also supports prettification when the origin request has a query parameter
// of "pretty" set to true.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)

	if pretty, _ := strconv.ParseBool(r.FormValue("pretty")); pretty {
		enc.SetIndent("", "    ")
	}

	if err := enc.Encode(v); err != nil && err != io.EOF {
		panic(fmt.Sprintf("failed to marshal response: %v", err))
	}
}

// M is an alias for map[string]any, which makes it easier to respond with generic JSON data structures.
type M map[string]any

var (
	// DefaultDecoder is the default decoder used by Bind. You can either override
	// this, or provide your own. Make sure it is set before Bind is called.
	DefaultDecoder = form.NewDecoder()

	// DefaultDecodeMaxMemory is the maximum amount of memory in bytes that will be
	// used for decoding multipart/form-data requests.
	DefaultDecodeMaxMemory int64 = 8 << 20
)

// Bind decodes the request body to the given struct. At this time the only supported
// content-types are application/json, application/x-www-form-urlencoded, as well as
// GET parameters.
func Bind(r *http.Request, v any) error {
	err := r.ParseForm()
	if err != nil {
		return &ErrBadRequest{Err: fmt.Errorf("parsing form parameters: %w", err)}
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		err = DefaultDecoder.Decode(v, r.Form)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		switch {
		case strings.HasPrefix(r.Header.Get("Content-Type"), "application/json"):
			dec := json.NewDecoder(r.Body)
			dec.DisallowUnknownFields()
			defer r.Body.Close()
			err = dec.Decode(v)
		case strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"):
			err = r.ParseMultipartForm(DefaultDecodeMaxMemory)
			if err == nil {
				err = DefaultDecoder.Decode(v, r.MultipartForm.Value)
			}
		default:
			err = DefaultDecoder.Decode(v, r.PostForm)
		}
	default:
		return &ErrBadRequest{Err: fmt.Errorf("unsupported method %s", r.Method)}
	}

	if err != nil {
		return &ErrBadRequest{Err: fmt.Errorf("error decoding %s request into required format (%T): %w", r.Method, v, err)}
	}
	return nil
}

// Req simplifies making an HTTP handler that returns a single result, and an error.
// The result, if not nil, must be JSON-marshalable. If result is nil, [http.StatusNoContent]
// will be returned.
func Req[Resp any](s *Server, op Operation, fn func(*http.Request) (*Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := fn(r)
		handleResponse(s, w, r, op, results, err)
	}
}

// resolveID resolves the ID from the request path, and unmarshals it into the provided type.
// Only supports string, int, and types that support UnmarshalText, UnmarshalJSON, or UnmarshalBinary
// (in that order).
func resolveID[T any](r *http.Request) (id T, err error) {
	value := r.PathValue("id")

	switch any(id).(type) {
	case string:
		id = any(value).(T)
	case int:
		rid, err := strconv.Atoi(value)
		if err == nil {
			id = any(rid).(T)
		}
	default:
		hasUnmarshal := false

		// Check if the underlying type supports UnmarshalText, UnmarshalJSON, or UnmarshalBinary.
		if u, ok := any(&id).(encoding.TextUnmarshaler); ok {
			hasUnmarshal = true
			err = u.UnmarshalText([]byte(value))
		} else if u, ok := any(&id).(json.Unmarshaler); ok {
			hasUnmarshal = true
			err = u.UnmarshalJSON([]byte(value))
		} else if u, ok := any(&id).(encoding.BinaryUnmarshaler); ok {
			hasUnmarshal = true
			err = u.UnmarshalBinary([]byte(value))
		}

		if !hasUnmarshal {
			panic(fmt.Sprintf("unsupported ID type (cannot unmarshal): %T", id))
		}
	}

	if err != nil {
		return id, &ErrInvalidID{ID: value, Err: err}
	}
	return id, nil
}

// ReqID is similar to Req, but also processes an "id" path parameter and provides it to the
// handler function.
func ReqID[Resp, I any](s *Server, op Operation, fn func(*http.Request, I) (*Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := resolveID[I](r)
		if err != nil {
			handleResponse[Resp](s, w, r, op, nil, err)
			return
		}
		results, err := fn(r, id)
		handleResponse(s, w, r, op, results, err)
	}
}

// ReqParam is similar to Req, but also processes a request body/query params and provides it
// to the handler function.
func ReqParam[Params, Resp any](s *Server, op Operation, fn func(*http.Request, *Params) (*Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := new(Params)
		if err := Bind(r, params); err != nil {
			handleResponse[Resp](s, w, r, op, nil, err)
			return
		}
		results, err := fn(r, params)
		handleResponse(s, w, r, op, results, err)
	}
}

// ReqIDParam is similar to ReqParam, but also processes an "id" path parameter and request
// body/query params, and provides it to the handler function.
func ReqIDParam[Params, Resp, I any](s *Server, op Operation, fn func(*http.Request, I, *Params) (*Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := resolveID[I](r)
		if err != nil {
			handleResponse[Resp](s, w, r, op, nil, err)
			return
		}
		params := new(Params)
		err = Bind(r, params)
		if err != nil {
			handleResponse[Resp](s, w, r, op, nil, err)
			return
		}
		results, err := fn(r, id, params)
		handleResponse(s, w, r, op, results, err)
	}
}

// Links represents a set of linkable-relationsips that can be represented through
// the "Link" header. Note that all urls must be url-encoded already.
type Links map[string]string

func (l Links) String() string {
	var links []string
	var keys []string
	for k := range l {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		links = append(links, fmt.Sprintf(`<%s>; rel=%q`, l[k], k))
	}
	return strings.Join(links, ", ")
}

type linkablePagedResource interface {
	GetPage() int
	GetIsLastPage() bool
}

// Spec returns the OpenAPI spec for the server implementation.
func (s *Server) Spec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !s.config.DisableSpecInjectServer && s.config.BaseURL != "" {
		spec := map[string]any{}
		err := json.Unmarshal(OpenAPI, &spec)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal spec: %v", err))
		}

		type Server struct {
			URL string `json:"url"`
		}

		if _, ok := spec["servers"]; !ok {
			spec["servers"] = []Server{{URL: s.config.BaseURL}}
			JSON(w, r, http.StatusOK, spec)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(OpenAPI)
}

var scalarTemplate = template.Must(template.New("docs").Parse(`<!DOCTYPE html>
<html>
  <head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="darkreader-lock">
    <link rel="icon" type="image/svg+xml"
      href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 1024 1024'%3E%3Cpath fill='currentColor' d='m917.7 148.8l-42.4-42.4c-1.6-1.6-3.6-2.3-5.7-2.3s-4.1.8-5.7 2.3l-76.1 76.1a199.27 199.27 0 0 0-112.1-34.3c-51.2 0-102.4 19.5-141.5 58.6L432.3 308.7a8.03 8.03 0 0 0 0 11.3L704 591.7c1.6 1.6 3.6 2.3 5.7 2.3c2 0 4.1-.8 5.7-2.3l101.9-101.9c68.9-69 77-175.7 24.3-253.5l76.1-76.1c3.1-3.2 3.1-8.3 0-11.4M769.1 441.7l-59.4 59.4l-186.8-186.8l59.4-59.4c24.9-24.9 58.1-38.7 93.4-38.7s68.4 13.7 93.4 38.7c24.9 24.9 38.7 58.1 38.7 93.4s-13.8 68.4-38.7 93.4m-190.2 105a8.03 8.03 0 0 0-11.3 0L501 613.3L410.7 523l66.7-66.7c3.1-3.1 3.1-8.2 0-11.3L441 408.6a8.03 8.03 0 0 0-11.3 0L363 475.3l-43-43a7.85 7.85 0 0 0-5.7-2.3c-2 0-4.1.8-5.7 2.3L206.8 534.2c-68.9 69-77 175.7-24.3 253.5l-76.1 76.1a8.03 8.03 0 0 0 0 11.3l42.4 42.4c1.6 1.6 3.6 2.3 5.7 2.3s4.1-.8 5.7-2.3l76.1-76.1c33.7 22.9 72.9 34.3 112.1 34.3c51.2 0 102.4-19.5 141.5-58.6l101.9-101.9c3.1-3.1 3.1-8.2 0-11.3l-43-43l66.7-66.7c3.1-3.1 3.1-8.2 0-11.3zM441.7 769.1a131.32 131.32 0 0 1-93.4 38.7c-35.3 0-68.4-13.7-93.4-38.7a131.32 131.32 0 0 1-38.7-93.4c0-35.3 13.7-68.4 38.7-93.4l59.4-59.4l186.8 186.8z'/%3E%3C/svg%3E" />
  </head>
  <body>
    <script id="api-reference"></script>
    <script>
      document.getElementById("api-reference").dataset.configuration = JSON.stringify({
        spec: {
          url: "{{ $.SpecPath }}",
        },
        {{- if $.DisableSpecInjectServer }}
        servers: [
            {url: window.location.origin + window.location.pathname.replace(/\/docs$/g, "")}
        ],
        {{- end }}
        theme: "kepler",
        isEditable: false,
        hideDownloadButton: true,
        customCss: ".darklight-reference-promo, .darklight-reference { visibility: hidden !important; height: 0 !important; } .open-api-client-button { display: none !important; }",
      });
    </script>
    <script
      src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.28.22"
      integrity="sha256-jx9Wj1V0D2G3VistmRA8MvsjTD3W0d5AW3Mu9Y17gwI="
      crossorigin="anonymous"
    ></script>
  </body>
</html>`))

func (s *Server) Docs(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	err := scalarTemplate.Execute(&buf, map[string]any{
		"SpecPath":                s.config.BasePath + "/openapi.json",
		"DisableSpecInjectServer": s.config.DisableSpecInjectServer,
	})
	if err != nil {
		handleResponse[struct{}](s, w, r, "", nil, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Security-Policy", "default-src 'self' cdn.jsdelivr.net fonts.scalar.com 'unsafe-inline' 'unsafe-eval' data: blob:")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
	w.Header().Set("Permissions-Policy", "clipboard-write=(self)")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

type ServerConfig struct {
	// BaseURL is similar to [ServerConfig.BasePath], however, only the path of the URL is used
	// to prefill BasePath. This is not required if BasePath is provided.
	BaseURL string

	// BasePath if provided, and the /openapi.json endpoint is enabled, will allow annotating
	// API responses with "Link" headers. See [ServerConfig.EnableLinks] for more information.
	BasePath string

	// DisableSpecHandler if set to true, will disable the /openapi.json endpoint. This will also
	// disable the embedded API reference documentation, see [ServerConfig.DisableDocs] for more
	// information.
	DisableSpecHandler bool

	// DisableSpecInjectServer if set to true, will disable the automatic injection of the
	// server URL into the spec. This only applies if [ServerConfig.BaseURL] is provided.
	DisableSpecInjectServer bool

	// DisableDocsHandler if set to true, will disable the embedded API reference documentation
	// endpoint at /docs. Use this if you want to provide your own documentation functionality.
	// This is disabled by default if [ServerConfig.DisableSpecHandler] is true.
	DisableDocsHandler bool

	// EnableLinks if set to true, will enable the "Link" response header, which can be used to hint
	// to clients about the location of the OpenAPI spec, API documentation, how to auto-paginate
	// through results, and more.
	EnableLinks bool

	// MaskErrors if set to true, will mask the error message returned to the client,
	// returning a generic error message based on the HTTP status code.
	MaskErrors bool

	// ErrorHandler is invoked when an error occurs. If not provided, the default
	// error handling logic will be used. If you want to run logic on errors, but
	// not actually handle the error yourself, you can still call [Server.DefaultErrorHandler]
	// after your logic.
	ErrorHandler func(w http.ResponseWriter, r *http.Request, op Operation, err error)

	// GetReqID returns the request ID for the given request. If not provided, the
	// default implementation will use the X-Request-Id header, otherwise an empty
	// string will be returned. If using go-chi, middleware.GetReqID will be used.
	GetReqID func(r *http.Request) string
}

type Server struct {
	db     *ent.Client
	config *ServerConfig
}

// NewServer returns a new auto-generated server implementation for your ent schema.
// [Server.Handler] returns a ready-to-use http.Handler that mounts all of the
// necessary endpoints.
func NewServer(db *ent.Client, config *ServerConfig) (*Server, error) {
	s := &Server{
		db:     db,
		config: config,
	}
	if s.config == nil {
		s.config = &ServerConfig{}
	}
	if s.config.BaseURL != "" && s.config.BasePath == "" {
		uri, err := url.Parse(s.config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse BaseURL: %w", err)
		}
		s.config.BasePath = uri.Path
	}
	if s.config.BaseURL == "" {
		s.config.DisableSpecInjectServer = true
	}
	if s.config.BasePath != "" {
		if !strings.HasPrefix(s.config.BasePath, "/") {
			s.config.BasePath = "/" + s.config.BasePath
		}
		s.config.BasePath = strings.TrimRight(s.config.BasePath, "/")
	}
	return s, nil
}

// DefaultErrorHandler is the default error handler for the Server.
func (s *Server) DefaultErrorHandler(w http.ResponseWriter, r *http.Request, op Operation, err error) {
	ts := time.Now().UTC().Format(time.RFC3339)

	resp := ErrorResponse{
		Error:     err.Error(),
		Timestamp: ts,
	}

	var numErr *strconv.NumError

	switch {
	case IsEndpointNotFound(err):
		resp.Code = http.StatusNotFound
	case IsMethodNotAllowed(err):
		resp.Code = http.StatusMethodNotAllowed
	case IsBadRequest(err):
		resp.Code = http.StatusBadRequest
	case IsInvalidID(err):
		resp.Code = http.StatusBadRequest
	case errors.Is(err, privacy.Deny):
		resp.Code = http.StatusForbidden
	case ent.IsNotFound(err):
		resp.Code = http.StatusNotFound
	case ent.IsConstraintError(err), ent.IsNotSingular(err):
		resp.Code = http.StatusConflict
	case ent.IsValidationError(err):
		resp.Code = http.StatusBadRequest
	case errors.As(err, &numErr):
		resp.Code = http.StatusBadRequest
		resp.Error = fmt.Sprintf("invalid ID provided: %v", err)
	default:
		resp.Code = http.StatusInternalServerError
	}

	if resp.Type == "" {
		resp.Type = http.StatusText(resp.Code)
	}
	if s.config.MaskErrors {
		resp.Error = http.StatusText(resp.Code)
	}
	if s.config.GetReqID != nil {
		resp.RequestID = s.config.GetReqID(r)
	} else {
		resp.RequestID = r.Header.Get("X-Request-Id")
	}
	JSON(w, r, resp.Code, resp)
}

func handleResponse[Resp any](s *Server, w http.ResponseWriter, r *http.Request, op Operation, resp *Resp, err error) {
	if s.config.EnableLinks {
		links := Links{}
		if !s.config.DisableSpecHandler {
			links["service-desc"] = s.config.BasePath + "/openapi.json"
			links["describedby"] = s.config.BasePath + "/openapi.json"
		}

		if err == nil && resp != nil && op == OperationList {
			if lr, ok := any(resp).(linkablePagedResource); ok {
				query := r.URL.Query()
				if page := lr.GetPage(); page > 1 {
					query.Set("page", strconv.Itoa(page-1))
					r.URL.RawQuery = query.Encode()
					links["prev"] = r.URL.String()
					if !strings.HasPrefix(links["prev"], s.config.BasePath) {
						links["prev"] = s.config.BasePath + links["prev"]
					}
				}
				if !lr.GetIsLastPage() {
					query.Set("page", strconv.Itoa(lr.GetPage()+1))
					r.URL.RawQuery = query.Encode()
					links["next"] = r.URL.String()
					if !strings.HasPrefix(links["next"], s.config.BasePath) {
						links["next"] = s.config.BasePath + links["next"]
					}
				}
			}
		}

		if v := links.String(); v != "" {
			w.Header().Set("Link", v)
		}
	}
	if err != nil {
		if s.config.ErrorHandler != nil {
			s.config.ErrorHandler(w, r, op, err)
			return
		}
		s.DefaultErrorHandler(w, r, op, err)
		return
	}
	if resp != nil {
		type pagedResp interface {
			GetTotalCount() int
		}
		if v, ok := any(resp).(pagedResp); ok && v.GetTotalCount() == 0 && r.Method == http.MethodGet {
			JSON(w, r, http.StatusNotFound, resp)
			return
		}
		if r.Method == http.MethodPost {
			JSON(w, r, http.StatusCreated, resp)
			return
		}
		JSON(w, r, http.StatusOK, resp)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UseEntContext can be used to inject an [ent.Client] into the context for use
// by other middleware, or ent privacy layers. Note that the server will do this
// by default, so you don't need to do this manually, unless it's a context that's
// not being passed to the server and is being consumed elsewhere.
func UseEntContext(db *ent.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ent.NewContext(r.Context(), db)))
		})
	}
}

// Handler returns a ready-to-use http.Handler that mounts all of the necessary endpoints.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", ReqParam(s, OperationList, s.ListCategories))
	mux.HandleFunc("GET /categories/{id}", ReqID(s, OperationRead, s.GetCategory))
	mux.HandleFunc("GET /categories/{id}/pets", ReqIDParam(s, OperationList, s.ListCategoryPets))
	mux.HandleFunc("POST /categories", ReqParam(s, OperationCreate, s.CreateCategory))
	mux.HandleFunc("PATCH /categories/{id}", ReqIDParam(s, OperationUpdate, s.UpdateCategory))
	mux.HandleFunc("DELETE /categories/{id}", ReqID(s, OperationDelete, s.DeleteCategory))
	mux.HandleFunc("GET /follows", ReqParam(s, OperationList, s.ListFollows))
	mux.HandleFunc("POST /follows", ReqParam(s, OperationCreate, s.CreateFollow))
	mux.HandleFunc("GET /friendships", ReqParam(s, OperationList, s.ListFriendships))
	mux.HandleFunc("GET /friendships/{id}", ReqID(s, OperationRead, s.GetFriendship))
	mux.HandleFunc("GET /friendships/{id}/user", ReqID(s, OperationRead, s.GetFriendshipUser))
	mux.HandleFunc("GET /friendships/{id}/friend", ReqID(s, OperationRead, s.GetFriendshipFriend))
	mux.HandleFunc("POST /friendships", ReqParam(s, OperationCreate, s.CreateFriendship))
	mux.HandleFunc("PATCH /friendships/{id}", ReqIDParam(s, OperationUpdate, s.UpdateFriendship))
	mux.HandleFunc("DELETE /friendships/{id}", ReqID(s, OperationDelete, s.DeleteFriendship))
	mux.HandleFunc("GET /pets", ReqParam(s, OperationList, s.ListPets))
	mux.HandleFunc("GET /pets/{id}", ReqID(s, OperationRead, s.GetPet))
	mux.HandleFunc("GET /pets/{id}/categories", ReqIDParam(s, OperationList, s.ListPetCategories))
	mux.HandleFunc("GET /pets/{id}/owner", ReqID(s, OperationRead, s.GetPetOwner))
	mux.HandleFunc("GET /pets/{id}/friends", ReqIDParam(s, OperationList, s.ListPetFriends))
	mux.HandleFunc("GET /pets/{id}/followed-by", ReqIDParam(s, OperationList, s.ListPetFollowedBys))
	mux.HandleFunc("POST /pets", ReqParam(s, OperationCreate, s.CreatePet))
	mux.HandleFunc("PATCH /pets/{id}", ReqIDParam(s, OperationUpdate, s.UpdatePet))
	mux.HandleFunc("DELETE /pets/{id}", ReqID(s, OperationDelete, s.DeletePet))
	mux.HandleFunc("GET /posts", ReqParam(s, OperationList, s.ListPosts))
	mux.HandleFunc("GET /posts/{id}", ReqID(s, OperationRead, s.GetPost))
	mux.HandleFunc("GET /posts/{id}/author", ReqID(s, OperationRead, s.GetPostAuthor))
	mux.HandleFunc("POST /posts", ReqParam(s, OperationCreate, s.CreatePost))
	mux.HandleFunc("PATCH /posts/{id}", ReqIDParam(s, OperationUpdate, s.UpdatePost))
	mux.HandleFunc("DELETE /posts/{id}", ReqID(s, OperationDelete, s.DeletePost))
	mux.HandleFunc("GET /settings", ReqParam(s, OperationList, s.ListSettings))
	mux.HandleFunc("GET /settings/{id}", ReqID(s, OperationRead, s.GetSetting))
	mux.HandleFunc("GET /settings/{id}/admins", ReqIDParam(s, OperationList, s.ListSettingAdmins))
	mux.HandleFunc("PATCH /settings/{id}", ReqIDParam(s, OperationUpdate, s.UpdateSetting))
	mux.HandleFunc("GET /users", ReqParam(s, OperationList, s.ListUsers))
	mux.HandleFunc("GET /users/{id}", ReqID(s, OperationRead, s.GetUser))
	mux.HandleFunc("GET /users/{id}/pets", ReqIDParam(s, OperationList, s.ListUserPets))
	mux.HandleFunc("GET /users/{id}/followed-pets", ReqIDParam(s, OperationList, s.ListUserFollowedPets))
	mux.HandleFunc("GET /users/{id}/friends", ReqIDParam(s, OperationList, s.ListUserFriends))
	mux.HandleFunc("GET /users/{id}/posts", ReqIDParam(s, OperationList, s.ListUserPosts))
	mux.HandleFunc("GET /users/{id}/friendships", ReqIDParam(s, OperationList, s.ListUserFriendships))
	mux.HandleFunc("POST /users", ReqParam(s, OperationCreate, s.CreateUser))
	mux.HandleFunc("PATCH /users/{id}", ReqIDParam(s, OperationUpdate, s.UpdateUser))
	mux.HandleFunc("DELETE /users/{id}", ReqID(s, OperationDelete, s.DeleteUser))

	if !s.config.DisableSpecHandler {
		mux.HandleFunc("GET /openapi.json", s.Spec)
	}

	if !s.config.DisableSpecHandler && !s.config.DisableDocsHandler {
		mux.HandleFunc("GET /docs", s.Docs)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !s.config.DisableSpecHandler && !s.config.DisableDocsHandler && r.URL.Path == "/" && r.Method == http.MethodGet {
			// If specs are enabled, it's safe to provide documentation, and if they don't override the
			// root endpoint, we can redirect to the docs.
			http.Redirect(w, r, s.config.BasePath+"/docs", http.StatusTemporaryRedirect)
			return
		}
		if r.Method != http.MethodGet {
			handleResponse[struct{}](s, w, r, "", nil, ErrMethodNotAllowed)
			return
		}
		handleResponse[struct{}](s, w, r, "", nil, ErrEndpointNotFound)
	})
	return http.StripPrefix(s.config.BasePath, UseEntContext(s.db)(mux))
}

// ListCategories maps to "GET /categories".
func (s *Server) ListCategories(r *http.Request, p *ListCategoryParams) (*PagedResponse[ent.Category], error) {
	return p.Exec(r.Context(), s.db.Category.Query())
}

// GetCategory maps to "GET /categories/{id}".
func (s *Server) GetCategory(r *http.Request, categoryID int) (*ent.Category, error) {
	return EagerLoadCategory(s.db.Category.Query().Where(category.ID(categoryID))).Only(r.Context())
}

// ListCategoryPets maps to "GET /categories/{id}/pets".
func (s *Server) ListCategoryPets(r *http.Request, categoryID int, p *ListPetParams) (*PagedResponse[ent.Pet], error) {
	return p.Exec(r.Context(), s.db.Category.Query().Where(category.ID(categoryID)).QueryPets())
}

// CreateCategory maps to "POST /categories".
func (s *Server) CreateCategory(r *http.Request, p *CreateCategoryParams) (*ent.Category, error) {
	return p.Exec(r.Context(), s.db.Category.Create(), s.db.Category.Query())
}

// UpdateCategory maps to "PATCH /categories/{id}".
func (s *Server) UpdateCategory(r *http.Request, categoryID int, p *UpdateCategoryParams) (*ent.Category, error) {
	return p.Exec(r.Context(), s.db.Category.UpdateOneID(categoryID), s.db.Category.Query())
}

// DeleteCategory maps to "DELETE /categories/{id}".
func (s *Server) DeleteCategory(r *http.Request, categoryID int) (*struct{}, error) {
	return nil, s.db.Category.DeleteOneID(categoryID).Exec(r.Context())
}

// ListFollows maps to "GET /follows".
func (s *Server) ListFollows(r *http.Request, p *ListFollowParams) (*PagedResponse[ent.Follows], error) {
	return p.Exec(r.Context(), s.db.Follows.Query())
}

// CreateFollow maps to "POST /follows".
func (s *Server) CreateFollow(r *http.Request, p *CreateFollowParams) (*ent.Follows, error) {
	return p.Exec(r.Context(), s.db.Follows.Create(), s.db.Follows.Query())
}

// ListFriendships maps to "GET /friendships".
func (s *Server) ListFriendships(r *http.Request, p *ListFriendshipParams) (*PagedResponse[ent.Friendship], error) {
	return p.Exec(r.Context(), s.db.Friendship.Query())
}

// GetFriendship maps to "GET /friendships/{id}".
func (s *Server) GetFriendship(r *http.Request, friendshipID int) (*ent.Friendship, error) {
	return EagerLoadFriendship(s.db.Friendship.Query().Where(friendship.ID(friendshipID))).Only(r.Context())
}

// GetFriendshipUser maps to "GET /friendships/{id}/user".
func (s *Server) GetFriendshipUser(r *http.Request, friendshipID int) (*ent.User, error) {
	return EagerLoadUser(s.db.Friendship.Query().Where(friendship.ID(friendshipID)).QueryUser()).Only(r.Context())
}

// GetFriendshipFriend maps to "GET /friendships/{id}/friend".
func (s *Server) GetFriendshipFriend(r *http.Request, friendshipID int) (*ent.User, error) {
	return EagerLoadUser(s.db.Friendship.Query().Where(friendship.ID(friendshipID)).QueryFriend()).Only(r.Context())
}

// CreateFriendship maps to "POST /friendships".
func (s *Server) CreateFriendship(r *http.Request, p *CreateFriendshipParams) (*ent.Friendship, error) {
	return p.Exec(r.Context(), s.db.Friendship.Create(), s.db.Friendship.Query())
}

// UpdateFriendship maps to "PATCH /friendships/{id}".
func (s *Server) UpdateFriendship(r *http.Request, friendshipID int, p *UpdateFriendshipParams) (*ent.Friendship, error) {
	return p.Exec(r.Context(), s.db.Friendship.UpdateOneID(friendshipID), s.db.Friendship.Query())
}

// DeleteFriendship maps to "DELETE /friendships/{id}".
func (s *Server) DeleteFriendship(r *http.Request, friendshipID int) (*struct{}, error) {
	return nil, s.db.Friendship.DeleteOneID(friendshipID).Exec(r.Context())
}

// ListPets maps to "GET /pets".
func (s *Server) ListPets(r *http.Request, p *ListPetParams) (*PagedResponse[ent.Pet], error) {
	return p.Exec(r.Context(), s.db.Pet.Query())
}

// GetPet maps to "GET /pets/{id}".
func (s *Server) GetPet(r *http.Request, petID int) (*ent.Pet, error) {
	return EagerLoadPet(s.db.Pet.Query().Where(pet.ID(petID))).Only(r.Context())
}

// ListPetCategories maps to "GET /pets/{id}/categories".
func (s *Server) ListPetCategories(r *http.Request, petID int, p *ListCategoryParams) (*PagedResponse[ent.Category], error) {
	return p.Exec(r.Context(), s.db.Pet.Query().Where(pet.ID(petID)).QueryCategories())
}

// GetPetOwner maps to "GET /pets/{id}/owner".
func (s *Server) GetPetOwner(r *http.Request, petID int) (*ent.User, error) {
	return EagerLoadUser(s.db.Pet.Query().Where(pet.ID(petID)).QueryOwner()).Only(r.Context())
}

// ListPetFriends maps to "GET /pets/{id}/friends".
func (s *Server) ListPetFriends(r *http.Request, petID int, p *ListPetParams) (*PagedResponse[ent.Pet], error) {
	return p.Exec(r.Context(), s.db.Pet.Query().Where(pet.ID(petID)).QueryFriends())
}

// ListPetFollowedBys maps to "GET /pets/{id}/followed-by".
func (s *Server) ListPetFollowedBys(r *http.Request, petID int, p *ListUserParams) (*PagedResponse[ent.User], error) {
	return p.Exec(r.Context(), s.db.Pet.Query().Where(pet.ID(petID)).QueryFollowedBy())
}

// CreatePet maps to "POST /pets".
func (s *Server) CreatePet(r *http.Request, p *CreatePetParams) (*ent.Pet, error) {
	return p.Exec(r.Context(), s.db.Pet.Create(), s.db.Pet.Query())
}

// UpdatePet maps to "PATCH /pets/{id}".
func (s *Server) UpdatePet(r *http.Request, petID int, p *UpdatePetParams) (*ent.Pet, error) {
	return p.Exec(r.Context(), s.db.Pet.UpdateOneID(petID), s.db.Pet.Query())
}

// DeletePet maps to "DELETE /pets/{id}".
func (s *Server) DeletePet(r *http.Request, petID int) (*struct{}, error) {
	return nil, s.db.Pet.DeleteOneID(petID).Exec(r.Context())
}

// ListPosts maps to "GET /posts".
func (s *Server) ListPosts(r *http.Request, p *ListPostParams) (*PagedResponse[ent.Post], error) {
	return p.Exec(r.Context(), s.db.Post.Query())
}

// GetPost maps to "GET /posts/{id}".
func (s *Server) GetPost(r *http.Request, postID int) (*ent.Post, error) {
	return EagerLoadPost(s.db.Post.Query().Where(post.ID(postID))).Only(r.Context())
}

// GetPostAuthor maps to "GET /posts/{id}/author".
func (s *Server) GetPostAuthor(r *http.Request, postID int) (*ent.User, error) {
	return EagerLoadUser(s.db.Post.Query().Where(post.ID(postID)).QueryAuthor()).Only(r.Context())
}

// CreatePost maps to "POST /posts".
func (s *Server) CreatePost(r *http.Request, p *CreatePostParams) (*ent.Post, error) {
	return p.Exec(r.Context(), s.db.Post.Create(), s.db.Post.Query())
}

// UpdatePost maps to "PATCH /posts/{id}".
func (s *Server) UpdatePost(r *http.Request, postID int, p *UpdatePostParams) (*ent.Post, error) {
	return p.Exec(r.Context(), s.db.Post.UpdateOneID(postID), s.db.Post.Query())
}

// DeletePost maps to "DELETE /posts/{id}".
func (s *Server) DeletePost(r *http.Request, postID int) (*struct{}, error) {
	return nil, s.db.Post.DeleteOneID(postID).Exec(r.Context())
}

// ListSettings maps to "GET /settings".
func (s *Server) ListSettings(r *http.Request, p *ListSettingParams) (*PagedResponse[ent.Settings], error) {
	return p.Exec(r.Context(), s.db.Settings.Query())
}

// GetSetting maps to "GET /settings/{id}".
func (s *Server) GetSetting(r *http.Request, settingID int) (*ent.Settings, error) {
	return EagerLoadSetting(s.db.Settings.Query().Where(settings.ID(settingID))).Only(r.Context())
}

// ListSettingAdmins maps to "GET /settings/{id}/admins".
func (s *Server) ListSettingAdmins(r *http.Request, settingID int, p *ListUserParams) (*PagedResponse[ent.User], error) {
	return p.Exec(r.Context(), s.db.Settings.Query().Where(settings.ID(settingID)).QueryAdmins())
}

// UpdateSetting maps to "PATCH /settings/{id}".
func (s *Server) UpdateSetting(r *http.Request, settingID int, p *UpdateSettingParams) (*ent.Settings, error) {
	return p.Exec(r.Context(), s.db.Settings.UpdateOneID(settingID), s.db.Settings.Query())
}

// ListUsers maps to "GET /users".
func (s *Server) ListUsers(r *http.Request, p *ListUserParams) (*PagedResponse[ent.User], error) {
	return p.Exec(r.Context(), s.db.User.Query())
}

// GetUser maps to "GET /users/{id}".
func (s *Server) GetUser(r *http.Request, userID uuid.UUID) (*ent.User, error) {
	return EagerLoadUser(s.db.User.Query().Where(user.ID(userID))).Only(r.Context())
}

// ListUserPets maps to "GET /users/{id}/pets".
func (s *Server) ListUserPets(r *http.Request, userID uuid.UUID, p *ListPetParams) (*PagedResponse[ent.Pet], error) {
	return p.Exec(r.Context(), s.db.User.Query().Where(user.ID(userID)).QueryPets())
}

// ListUserFollowedPets maps to "GET /users/{id}/followed-pets".
func (s *Server) ListUserFollowedPets(r *http.Request, userID uuid.UUID, p *ListPetParams) (*PagedResponse[ent.Pet], error) {
	return p.Exec(r.Context(), s.db.User.Query().Where(user.ID(userID)).QueryFollowedPets())
}

// ListUserFriends maps to "GET /users/{id}/friends".
func (s *Server) ListUserFriends(r *http.Request, userID uuid.UUID, p *ListUserParams) (*PagedResponse[ent.User], error) {
	return p.Exec(r.Context(), s.db.User.Query().Where(user.ID(userID)).QueryFriends())
}

// ListUserPosts maps to "GET /users/{id}/posts".
func (s *Server) ListUserPosts(r *http.Request, userID uuid.UUID, p *ListPostParams) (*PagedResponse[ent.Post], error) {
	return p.Exec(r.Context(), s.db.User.Query().Where(user.ID(userID)).QueryPosts())
}

// ListUserFriendships maps to "GET /users/{id}/friendships".
func (s *Server) ListUserFriendships(r *http.Request, userID uuid.UUID, p *ListFriendshipParams) (*PagedResponse[ent.Friendship], error) {
	return p.Exec(r.Context(), s.db.User.Query().Where(user.ID(userID)).QueryFriendships())
}

// CreateUser maps to "POST /users".
func (s *Server) CreateUser(r *http.Request, p *CreateUserParams) (*ent.User, error) {
	return p.Exec(r.Context(), s.db.User.Create(), s.db.User.Query())
}

// UpdateUser maps to "PATCH /users/{id}".
func (s *Server) UpdateUser(r *http.Request, userID uuid.UUID, p *UpdateUserParams) (*ent.User, error) {
	return p.Exec(r.Context(), s.db.User.UpdateOneID(userID), s.db.User.Query())
}

// DeleteUser maps to "DELETE /users/{id}".
func (s *Server) DeleteUser(r *http.Request, userID uuid.UUID) (*struct{}, error) {
	return nil, s.db.User.DeleteOneID(userID).Exec(r.Context())
}
