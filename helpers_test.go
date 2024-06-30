// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		num := 1
		assert.Equal(t, num, *ptr(1))
	})

	t.Run("string", func(t *testing.T) {
		str := "test"
		assert.Equal(t, str, *ptr("test"))
	})

	t.Run("bool", func(t *testing.T) {
		b := true
		assert.Equal(t, b, *ptr(true))
	})
}

func TestMemoize(t *testing.T) {
	count := 0

	fn := func(in string) string {
		count++
		return in + "_" + strconv.Itoa(count)
	}
	mfn := memoize(fn)

	assert.Equal(t, "foo_1", fn("foo"))
	assert.Equal(t, "foo_2", fn("foo"))
	assert.Equal(t, "foo_3", mfn("foo"))
	assert.Equal(t, "foo_3", mfn("foo"))
	assert.Equal(t, "bar_4", mfn("bar"))
	assert.Equal(t, "bar_4", mfn("bar"))
}

func TestSliceToRawMessage(t *testing.T) {
	tests := []struct {
		name string
		in   []any
		out  []json.RawMessage
	}{
		{
			name: "empty",
			in:   []any{},
			out:  []json.RawMessage{},
		},
		{
			name: "single",
			in:   []any{"foo"},
			out:  []json.RawMessage{json.RawMessage(`"foo"`)},
		},
		{
			name: "multiple",
			in:   []any{"foo", "bar", "baz"},
			out:  []json.RawMessage{json.RawMessage(`"foo"`), json.RawMessage(`"bar"`), json.RawMessage(`"baz"`)},
		},
		{
			name: "numbers",
			in:   []any{1, 2, 3},
			out:  []json.RawMessage{json.RawMessage(`1`), json.RawMessage(`2`), json.RawMessage(`3`)},
		},
		{
			name: "bools",
			in:   []any{true, false},
			out:  []json.RawMessage{json.RawMessage(`true`), json.RawMessage(`false`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, sliceToRawMessage(tt.in))
		})
	}
}

func TestAppendCompactFunc(t *testing.T) {
	tests := []struct {
		name string
		orig []string
		newv []string
		fn   func(oldv, newv string) (matches bool)
		out  []string
	}{
		{
			name: "empty",
			orig: []string{},
			newv: []string{"foo"},
			fn:   func(oldv, newv string) bool { return oldv == newv },
			out:  []string{"foo"},
		},
		{
			name: "single",
			orig: []string{"foo"},
			newv: []string{"bar"},
			fn:   func(oldv, newv string) bool { return oldv == newv },
			out:  []string{"foo", "bar"},
		},
		{
			name: "multiple",
			orig: []string{"foo", "bar"},
			newv: []string{"baz", "qux"},
			fn:   func(oldv, newv string) bool { return oldv == newv },
			out:  []string{"foo", "bar", "baz", "qux"},
		},
		{
			name: "duplicate",
			orig: []string{"foo", "bar"},
			newv: []string{"foo", "bar"},
			fn:   func(oldv, newv string) bool { return oldv == newv },
			out:  []string{"foo", "bar"},
		},
		{
			name: "not duplicate",
			orig: []string{"foo", "bar"},
			newv: []string{"baz", "qux"},
			fn:   func(oldv, newv string) bool { return oldv == newv },
			out:  []string{"foo", "bar", "baz", "qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, appendCompactFunc(tt.orig, tt.newv, tt.fn))
		})
	}
}

func TestAppendCompact(t *testing.T) {
	tests := []struct {
		name string
		orig []any
		newv []any
		out  []any
	}{
		{
			name: "empty",
			orig: []any{},
			newv: []any{"foo"},
			out:  []any{"foo"},
		},
		{
			name: "single",
			orig: []any{"foo"},
			newv: []any{"bar"},
			out:  []any{"foo", "bar"},
		},
		{
			name: "multiple",
			orig: []any{"foo", "bar"},
			newv: []any{"baz", "qux"},
			out:  []any{"foo", "bar", "baz", "qux"},
		},
		{
			name: "duplicate",
			orig: []any{"foo", "bar"},
			newv: []any{"foo", "bar"},
			out:  []any{"foo", "bar"},
		},
		{
			name: "not duplicate",
			orig: []any{"foo", "bar"},
			newv: []any{"baz", "qux"},
			out:  []any{"foo", "bar", "baz", "qux"},
		},
		{
			name: "numbers",
			orig: []any{1},
			newv: []any{2, 3},
			out:  []any{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, appendCompact(tt.orig, tt.newv))
		})
	}
}

func TestMergeMap(t *testing.T) {
	tests := []struct {
		name    string
		orig    map[string]any
		toMerge map[string]any
		out     map[string]any
		overlap bool
		wantErr bool
	}{
		{
			name:    "empty",
			orig:    map[string]any{},
			toMerge: map[string]any{"foo": "bar"},
			out:     map[string]any{"foo": "bar"},
			overlap: false,
			wantErr: false,
		},
		{
			name:    "single-non-overlap",
			orig:    map[string]any{"foo": "bar"},
			toMerge: map[string]any{"foo": "baz"},
			overlap: false,
			wantErr: true,
		},
		{
			name:    "single-overlap",
			orig:    map[string]any{"foo": "bar"},
			toMerge: map[string]any{"foo": "baz"},
			out:     map[string]any{"foo": "baz"},
			overlap: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeMap(tt.overlap, tt.orig, tt.toMerge)
			assert.Equal(t, tt.wantErr, err != nil)

			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.out, tt.orig)
		})
	}
}

func TestSliceOr(t *testing.T) {
	t.Run("string-zero", func(t *testing.T) {
		assert.Equal(t, []string{"foo"}, sliceOr([]string{}, []string{"foo"}))
	})

	t.Run("string-zero-2", func(t *testing.T) {
		assert.Equal(t, []string{"foo"}, sliceOr([]string{}, []string{"foo"}, []string{}))
	})

	t.Run("string-single", func(t *testing.T) {
		assert.Equal(t, []string{"bar"}, sliceOr([]string{"bar"}, []string{"foo"}))
	})

	t.Run("string-multiple", func(t *testing.T) {
		assert.Equal(t, []string{"baz", "qux"}, sliceOr([]string{"baz", "qux"}, []string{"foo"}, []string{"bar"}))
	})
}

func TestMapKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		out  []string
	}{
		{
			name: "empty",
			m:    map[string]any{},
			out:  []string{},
		},
		{
			name: "single",
			m:    map[string]any{"foo": "bar"},
			out:  []string{"foo"},
		},
		{
			name: "multiple",
			m:    map[string]any{"foo": "bar", "baz": "qux"},
			out:  []string{"baz", "foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, mapKeys(tt.m))
		})
	}
}
