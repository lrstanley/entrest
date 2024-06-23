// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
)

// ptr returns a pointer to the given value. Should only be used for primitives.
func ptr[T any](v T) *T {
	return &v
}

// memoize memoizes the provided function, so that it is only called once for each
// input.
func memoize[K comparable, V any](fn func(K) V) func(K) V {
	var mu sync.RWMutex
	cache := map[K]V{}

	return func(in K) V {
		mu.RLock()
		if cached, ok := cache[in]; ok {
			mu.RUnlock()
			return cached
		}
		mu.RUnlock()

		mu.Lock()
		defer mu.Unlock()

		cache[in] = fn(in)
		return cache[in]
	}
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

// withDefaultSlice returns the provided default value(s) if the given value is nil.
func withDefaultSlice[T any](v []T, defaults ...[]T) []T {
	if len(v) == 0 {
		for i := range defaults {
			if len(defaults[i]) > 0 {
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
