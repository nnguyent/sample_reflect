package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/require"
)

func TestSanitize(t *testing.T) {
	type SimpleStructPrivateFields struct {
		str     string
		integer int
		decimal float32
	}
	type SimpleStructPublicFields struct {
		Str     string
		Integer int
		Decimal float32
	}

	tests := []struct {
		name       string
		beforeTest func() interface{}
		expectedFn func(interface{})
	}{
		{
			name: "simple_structure_private_fields",
			beforeTest: func() (obj interface{}) {
				return SimpleStructPrivateFields{
					str:     "this is string",
					integer: 100,
					decimal: 10.01,
				}
			},
			expectedFn: func(got interface{}) {
				want := SimpleStructPrivateFields{
					str:     "this is string",
					integer: 100,
					decimal: 10.01,
				}

				require.True(t, reflect.DeepEqual(got, want), "got: %+v, want: %+v", got, want)
			},
		},
		{
			name: "simple_structure_public_fields",
			beforeTest: func() (obj interface{}) {
				return SimpleStructPublicFields{
					Str:     "this is string",
					Integer: 100,
					Decimal: 10.01,
				}
			},
			expectedFn: func(got interface{}) {
				want := SimpleStructPublicFields{
					Str:     "this is string_updated",
					Integer: 100,
					Decimal: 10.01,
				}

				require.True(t, reflect.DeepEqual(got, want), "got: %+v, want: %+v", got, want)
			},
		},
		{
			name: "pointer_of_simple_structure_private_fields",
			beforeTest: func() (obj interface{}) {
				return &SimpleStructPrivateFields{
					str:     "this is string",
					integer: 100,
					decimal: 10.01,
				}
			},
			expectedFn: func(got interface{}) {
				want := &SimpleStructPrivateFields{
					str:     "this is string",
					integer: 100,
					decimal: 10.01,
				}

				require.True(t, reflect.DeepEqual(got, want), "got: %+v, want: %+v", got, want)
			},
		},
		{
			name: "pointer_of_simple_structure_public_fields",
			beforeTest: func() (obj interface{}) {
				return &SimpleStructPublicFields{
					Str:     "this is string",
					Integer: 100,
					Decimal: 10.01,
				}
			},
			expectedFn: func(got interface{}) {
				want := &SimpleStructPublicFields{
					Str:     "this is string_updated",
					Integer: 100,
					Decimal: 10.01,
				}

				require.True(t, reflect.DeepEqual(got, want), "got: %+v, want: %+v", got, want)
			},
		},
		// ----
		// {
		// 	name: "structure_private_fields",
		// 	beforeTest: func() (obj interface{}) {
		// 		return struct {
		// 			sliceStr []string
		// 			// sliceObjPrivField []SimpleStructPrivateFields
		// 			// sliceObjPubField  []SimpleStructPublicFields
		// 			// mapStr            map[string]string
		// 			// mapObjPrivField   map[string]SimpleStructPrivateFields
		// 			// mapObjPubField    map[string]SimpleStructPublicFields
		// 		}{
		// 			sliceStr: []string{"string 1", "string 2"},
		// 			// sliceObjPrivField: []SimpleStructPrivateFields{{str: "this is string", integer: 100, decimal: 10.01}},
		// 			// sliceObjPubField:  []SimpleStructPublicFields{{Str: "this is string", Integer: 100, Decimal: 10.01}},
		// 			// mapStr:            map[string]string{"1": "this is string"},
		// 			// mapObjPrivField:   map[string]SimpleStructPrivateFields{"1": {str: "this is string", integer: 100, decimal: 10.01}},
		// 			// mapObjPubField:    map[string]SimpleStructPublicFields{"1": {Str: "this is string", Integer: 100, Decimal: 10.01}},
		// 		}
		// 	},
		// 	expectedFn: func(got interface{}) {
		// 		want := struct {
		// 			sliceStr []string
		// 			// sliceObjPrivField []SimpleStructPrivateFields
		// 			// sliceObjPubField  []SimpleStructPublicFields
		// 			// mapStr            map[string]string
		// 			// mapObjPrivField   map[string]SimpleStructPrivateFields
		// 			// mapObjPubField    map[string]SimpleStructPublicFields
		// 		}{
		// 			sliceStr: []string{"string 1", "string 2"},
		// 			// sliceObjPrivField: []SimpleStructPrivateFields{{str: "this is string", integer: 100, decimal: 10.01}},
		// 			// sliceObjPubField:  []SimpleStructPublicFields{{Str: "this is string", Integer: 100, Decimal: 10.01}},
		// 			// mapStr:            map[string]string{"1": "this is string"},
		// 			// mapObjPrivField:   map[string]SimpleStructPrivateFields{"1": {str: "this is string", integer: 100, decimal: 10.01}},
		// 			// mapObjPubField:    map[string]SimpleStructPublicFields{"1": {Str: "this is string", Integer: 100, Decimal: 10.01}},
		// 		}

		// 		require.True(t, reflect.DeepEqual(got, want), "got: %+v, want: %+v", got, want)
		// 	},
		// },
	}

	p := bluemonday.UGCPolicy()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			obj := tc.beforeTest()
			out := sanitize(p, obj)
			tc.expectedFn(out)
		})
	}
}

func Foo(s *[]string) {
	fmt.Printf("s: %p, %v\n", s, s)
	// s = []string{"a", "b"}
	*s = append(*s, "a")
	*s = append(*s, "b")
	fmt.Printf("s: %p, %v\n", s, s)
}

func TestSlice(t *testing.T) {
	slice := []string{}
	Foo(&slice)
	fmt.Printf("s: %p, %v\n", slice, slice)
}
