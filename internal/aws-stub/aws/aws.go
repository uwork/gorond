// Package aws provides stub types for the AWS SDK.
// This is a minimal stub used for building without network access.
package aws

// Config holds AWS configuration.
type Config struct {
	Region *string
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
