package main

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/microcosm-cc/bluemonday"
)

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func setUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

// traverse through struct
// refer from this cool snippet https://gist.github.com/hvoecking/10772475
func sanitize(p *bluemonday.Policy, obj interface{}) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)
	copy := reflect.New(original.Type()).Elem()

	fmt.Printf("Original: %+v, type: %+v, can get address: %t\nCopy: %+v, can get address: %t\n",
		original, original.Type(), original.CanAddr(),
		copy, copy.CanAddr(),
	)

	sanitizeRecursive(p, copy, original)

	// Remove the reflection wrapper
	return copy.Interface()
}

func sanitizeRecursive(p *bluemonday.Policy, copy, original reflect.Value) {
	// fmt.Printf("Original kind: %+v\n",
	// 	original.Kind(),
	// )

	switch original.Kind() {
	// The first cases handle nested structures and sanitize them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		fmt.Printf("Original kind %v, type %v, value %+v\n",
			original.Kind(), original.Type(), originalValue,
		)

		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		sanitizeRecursive(p, copy.Elem(), originalValue)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		fmt.Printf("Original kind %v, type %s, value %+v\n",
			original.Kind(), original.Type(), reflect.ValueOf(original.Interface()),
		)

		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		sanitizeRecursive(p, copyValue, originalValue)
		copy.Set(copyValue)

	// If it is a struct we sanitize each field
	case reflect.Struct:
		// fmt.Printf("Original kind %v, type %s, has %d fields, value %+v\n",
		// 	original.Kind(), original.Type(), original.NumField(), reflect.ValueOf(original.Interface()),
		// )
		//---
		fmt.Printf("Original kind %v, type %s, has %d fields\n",
			original.Kind(), original.Type(), original.NumField(),
		)

		// This way can't access the private fields
		// for i := 0; i < original.NumField(); i += 1 {
		// 	sanitizeRecursive(p, copy.Field(i), original.Field(i))
		// }

		// Should convert original from not addressable to addressable
		// in order to use unsafe.Pointer to access the (private) fields (read & write)
		converted := reflect.New(original.Type()).Elem()
		converted.Set(original)
		for i := 0; i < converted.NumField(); i += 1 {
			if converted.Field(i).Kind() == reflect.Slice {
				// t := reflect.MakeSlice(converted.Field(i).Type(), 0 , 0).Type()
				// copy.Field(i) = reflect.New(t).Elem()
				// copy.Field(i).Set(v)
				// copy.Field(i).Set(reflect.MakeSlice(converted.Field(i).Type(), converted.Field(i).Len(), converted.Field(i).Cap()))

				// copy.Field(i).Set(reflect.MakeSlice(converted.Field(i).Type(), converted.Field(i).Len(), converted.Field(i).Cap()))
				fmt.Printf("struct field %d kind %v, type %s, value %+v\n", i, copy.Field(i).Kind(), copy.Field(i).Type(), copy.Field(i))
				sanitizeRecursive(p, copy.Field(i).Addr(), converted.Field(i))

				fmt.Printf(">>> struct field %d kind %v, type %s, value %+v\n", i, copy.Field(i).Kind(), copy.Field(i).Type(), copy.Field(i))
			} else {
				sanitizeRecursive(p, copy.Field(i), converted.Field(i))
			}
		}

	// If it is a slice we create a new slice and sanitize each element
	case reflect.Slice:

		if copy.Kind() == reflect.Ptr {
			// copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
			copy = copy.Elem()
		}

		// fmt.Printf("Original kind %v, type %s, length: %d, capacity: %d\nCopy kind %v, type %s, length: %d, capacity: %d\n",
		// 	original.Kind(), original.Type(), original.Len(), original.Cap(),
		// 	copy.Kind(), copy.Type(), copy.Len(), copy.Cap())

		fmt.Printf("Original kind %v, type %s, length: %d, capacity: %d\nCopy kind %v, type %s\n",
			original.Kind(), original.Type(), original.Len(), original.Cap(),
			copy.Kind(), copy.Type())

		if !original.IsValid() {
			return
		}

		if original.CanInterface() {
			copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		} else {
			copy = reflect.MakeSlice(original.Type(), original.Len(), original.Cap())
			// copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
			fmt.Printf("Copy kind %v, type %s, length: %d, capacity: %d\n",
			copy.Kind(), copy.Type(), copy.Len(), copy.Cap(),)

			// 	copy = reflect.MakeSlice(original.Type(), 0,0)
			// 	// ---
			// 	// copy = reflect.New(original.Type()).Elem()
			// 	// copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
			// 	// ---
			// 	// copy = reflect.New(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()).Type()).Elem()
			// 	// copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
			// 	// fmt.Printf("copy kind %v, type %s\n", copy.Kind(), copy.Type())
			// 	fmt.Printf("copy kind %v, type %s, length: %d, capacity: %d\n", copy.Kind(), copy.Type(), copy.Len(), copy.Cap())
		}
		// copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))

		for i := 0; i < original.Len(); i += 1 {
			// sanitizeRecursive(p, copy.Index(i), original.Index(i))
			// ---
			val := reflect.New(original.Index(i).Type()).Elem()
			sanitizeRecursive(p, val, original.Index(i))
			fmt.Printf("val kind %v, type %s, value: %+v\n", 
			val.Kind(), val.Type(), val,
		)
			// copy = reflect.Append(copy, val)
			copy.Index(i).Set(val)
			// ---
		}

		fmt.Printf("copy value: %+v\n", copy)

	// If it is a map we create a new map and sanitize each value
	case reflect.Map:
		fmt.Printf("Original kind %v, type %s\n", original.Kind(), original.Type())
		if !original.IsValid() {
			return
		}

		if original.CanInterface() {
			copy.Set(reflect.MakeMap(original.Type()))
		} else {
			copy = reflect.MakeMap(original.Type())
		}

		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			sanitizeRecursive(p, copyValue, originalValue)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string sanitize it (yay finally we're doing what we came for)
	case reflect.String:
		fmt.Printf("Original kind %v, type %s\n", original.Kind(), original.Type())

		// sanitizedString := p.Sanitize(original.Interface().(string))
		if original.IsValid() {
			if original.CanInterface() {
				fmt.Printf("Original value from public filed: %s\n", original.Interface().(string))

				sanitizedString := original.Interface().(string) + "_updated"
				copy.SetString(sanitizedString)
			} else {
				// Get unexported field
				val := getUnexportedField(original)
				fmt.Printf("Original value from private filed: %s\n", val.(string))

				// Set unexported field
				setUnexportedField(copy, val)
			}
		}

	// And everything else will simply be taken from the original
	default:
		fmt.Printf("Original kind %v, type %s\n", original.Kind(), original.Type())
		if original.IsValid() {
			if original.CanInterface() {
				copy.Set(original)
			} else {
				// Get unexported field
				val := getUnexportedField(original)

				// Set unexported field
				setUnexportedField(copy, val)
			}
		}
	}

}

func main() {

}
