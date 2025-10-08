package internal

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// StructToQueryParams converts a struct or map to URL query parameters
func StructToQueryParams(v interface{}) url.Values {
	params := url.Values{}
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// Handle pointer types
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return params
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	// Handle maps
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			value := val.MapIndex(key)

			// Handle nil values for pointers, maps, slices, and interfaces
			switch value.Kind() {
			case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
				if value.IsNil() {
					continue
				}
			}

			valueStr := fieldValueToString(value)
			if valueStr != "" {
				params.Add(keyStr, valueStr)
			}
		}
		return params
	}

	if val.Kind() != reflect.Struct {
		return params
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the json tag or use field name
		tag := fieldType.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag (remove omitempty, etc.)
		name := tag
		if idx := len(tag); idx > 0 {
			for j, r := range tag {
				if r == ',' {
					idx = j
					break
				}
			}
			name = tag[:idx]
		}

		// Dereference pointers for type inspection
		fieldValue := field
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		// Handle slices using comma-separated values
		if fieldValue.Kind() == reflect.Slice {
			length := fieldValue.Len()
			if length == 0 {
				continue
			}

			var values []string
			for idx := 0; idx < length; idx++ {
				elem := fieldValue.Index(idx)
				elemStr := fieldValueToString(elem)
				if elemStr == "" {
					continue
				}
				values = append(values, elemStr)
			}

			if len(values) > 0 {
				params.Add(name, strings.Join(values, ","))
			}
			continue
		}

		// Handle maps by flattening key-value pairs using dot notation
		if fieldValue.Kind() == reflect.Map {
			iter := fieldValue.MapRange()
			for iter.Next() {
				subKey := fmt.Sprintf("%v", iter.Key().Interface())
				subValue := iter.Value()

				switch subValue.Kind() {
				case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
					if subValue.IsNil() {
						continue
					}
				}

				subValueStr := fieldValueToString(subValue)
				if subValueStr == "" {
					continue
				}

				params.Add(fmt.Sprintf("%s.%s", name, subKey), subValueStr)
			}
			continue
		}

		// Convert field value to string
		value := fieldValueToString(field)
		if value != "" {
			params.Add(name, value)
		}
	}

	return params
}

// fieldValueToString converts a reflect.Value to its string representation
func fieldValueToString(v reflect.Value) string {
	// Handle nil pointers
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return ""
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Slice:
		// Handle slices (convert to comma-separated values)
		if v.Len() == 0 {
			return ""
		}
		var result string
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				result += ","
			}
			result += fieldValueToString(v.Index(i))
		}
		return result
	default:
		// Try to convert to string using String() method if available
		if v.CanInterface() {
			if stringer, ok := v.Interface().(interface{ String() string }); ok {
				return stringer.String()
			}
			// Handle time.Time specially
			if t, ok := v.Interface().(time.Time); ok {
				return t.Format(time.RFC3339)
			}
		}
		return ""
	}
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}
