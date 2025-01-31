Conditional Requests

- WithConditional(bool) && WithConditionalRequired(methods []string)
  - if required, WithConditional automatically is set to true
  - if required, requests MUST contain conditional headers for all updates (GETs would be implicitly supported)
  - Have an option to require for specific methods, and another annotation function for ALL writes?
- etag
  - shasum of all marked fields (crc32, don't forget to denote security)
  - if etag matches, return 304
- last-modified
  - if field is of type datetime, use If-Modified-Since and similar (but still include in etag)

Scanning logic:

```go
	scanned, err := s.db.User.Query().Where(user.ID(userID)).Select(
		user.FieldUsername,
		user.FieldEmail,
		user.FieldUpdatedAt,
	).Only(r.Context())
	if err != nil {
		return nil, err
	}

	fmt.Printf("values: %#v\n", scanned)
```

separate function to generate the etag and modified time?

when updating, we need to recalculate the etag and modified time AFTER we update, so the client knows the latest updated version.

428 precondition required

```go
type Context[I any, P any] struct {
	context.Context

	Request *http.Request
	Response http.ResponseWriter
	ID *I
	Params *P
}
```

TODO: how do we grab the previous version from the database?
we should also get the etag, and do the update, in the same transaction?
