// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
)

func ptr[T any](v T) *T {
	return &v
}

func sliceToRawMessage[T any](v []T) []json.RawMessage {
	r := make([]json.RawMessage, len(v))
	var err error
	for i, v := range v {
		r[i], err = json.Marshal(v)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %v: %w", v, err))
		}
	}
	return r
}

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

func appendIfNotContains[T comparable](orig, newv []T) []T {
	return appendIfNotContainsFunc(orig, newv, func(oldv, newv T) bool {
		return oldv == newv
	})
}

func mergeMap[K comparable, V any](overlap bool, orig, newv map[K]V) error {
	if orig == nil {
		orig = make(map[K]V)
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
