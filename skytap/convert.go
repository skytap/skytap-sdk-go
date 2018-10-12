package skytap

// String returns a string value for the passed string pointer.
// It returns the empty string if the pointer is nil.
func String(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// StringPtr returns a pointer to the passed string.
func StringPtr(s string) *string {
	return &s
}
