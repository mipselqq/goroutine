package httpschema

type contextKey int

const (
	ContextKeyUserID    contextKey = iota
	ContextKeyRequestID contextKey = iota
)
