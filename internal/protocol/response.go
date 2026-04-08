package protocol

// ErrorResponse formats a protocol error using the Redis-like ERR prefix.
func ErrorResponse(message string) string {
	return "ERR " + message
}
