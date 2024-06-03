// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/ogen-go/ogen"
)

// PatchOperations applies a callback to each operation in a path inside of the OpenAPI spec.
func PatchOperations(pathItem *ogen.PathItem, cb func(method string, op *ogen.Operation) *ogen.Operation) *ogen.PathItem {
	pathItem.Get = cb(http.MethodGet, pathItem.Get)
	pathItem.Post = cb(http.MethodPost, pathItem.Post)
	pathItem.Put = cb(http.MethodPut, pathItem.Put)
	pathItem.Patch = cb(http.MethodPatch, pathItem.Patch)
	pathItem.Delete = cb(http.MethodDelete, pathItem.Delete)
	pathItem.Options = cb(http.MethodOptions, pathItem.Options)
	pathItem.Head = cb(http.MethodHead, pathItem.Head)
	pathItem.Trace = cb(http.MethodTrace, pathItem.Trace)
	return pathItem
}

// PatchPathItem applies a callback to each response in a path inside of the OpenAPI spec.
func PatchPathItem(pathItem *ogen.PathItem, cb func(resp *ogen.Response) *ogen.Response) *ogen.PathItem {
	return PatchOperations(pathItem, func(_ string, op *ogen.Operation) *ogen.Operation {
		if op == nil {
			return nil
		}
		for k, v := range op.Responses {
			op.Responses[k] = cb(v)
		}
		return op
	})
}

// ptr returns a pointer to the given value. Should only be used for primitives.
func ptr[T any](v T) *T {
	return &v
}

// sliceToRawMessage returns a slice of json.RawMessage from a slice of T. Panics
// if any of the values cannot be marshaled to JSON.
func sliceToRawMessage[T any](v []T) []json.RawMessage {
	r := make([]json.RawMessage, len(v))
	var err error
	for i, v := range v {
		r[i], err = json.Marshal(v)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %v: %v", v, err))
		}
	}
	return r
}

// appendIfNotContainsFunc returns a copy of orig with newv appended to it, but only if
// newv does not already exist in orig. fn is used to determine if two values are equal.
func appendIfNotContainsFunc[T any](orig, newv []T, fn func(oldv, newv T) (matches bool)) []T {
	for _, v := range newv {
		var found bool
		for _, ov := range orig {
			if fn(ov, v) {
				found = true
				break
			}
		}
		if !found {
			orig = append(orig, v)
		}
	}
	return orig
}

// appendIfNotContains returns a copy of orig with newv appended to it, but only if
// newv does not already exist in orig. T must be comparable.
func appendIfNotContains[T comparable](orig, newv []T) []T {
	return appendIfNotContainsFunc(orig, newv, func(oldv, newv T) bool {
		return oldv == newv
	})
}

// mergeMap returns a copy of orig with newv merged into it, but only if
// newv does not already exist in orig. If orig is nil, this will panic, as we cannot
// merge into a nil map without returning a new map.
func mergeMap[K comparable, V any](overlap bool, orig, newv map[K]V) error {
	if orig == nil {
		panic("orig is nil")
	}
	if newv == nil {
		return nil
	}

	for k, v := range newv {
		_, ok := orig[k]
		if !overlap && ok {
			return fmt.Errorf("key %v already exists in original map", k)
		}

		if !ok || overlap {
			orig[k] = v
			continue
		}
	}
	return nil
}

// withDefault returns the provided default value if the given value is the zero value.
func withDefault[T comparable](v T, defaults ...T) T {
	var zero T
	if v == zero {
		for i := range defaults {
			if defaults[i] != zero {
				return defaults[i]
			}
		}
	}
	return v
}

// mapKeys returns the keys of the map m, sorted.
func mapKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	slices.Sort(r)
	return r
}
