package beeperdesktop

import (
	"github.com/beeper/desktop-api-go/internal"
	"net/url"
)

// Utility functions for pointer conversion and query building

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

// Int64Ptr returns a pointer to the given int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the given float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// BuildQuery builds URL query parameters from a struct
func BuildQuery(params interface{}) url.Values {
	return internal.StructToQueryParams(params)
}
