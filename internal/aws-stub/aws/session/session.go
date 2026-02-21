// Package session provides stub types for the AWS SDK session.
package session

// Session holds AWS session configuration.
type Session struct{}

// New creates a new AWS session.
func New(cfgs ...interface{}) *Session {
	return &Session{}
}
