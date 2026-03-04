package httpschema

func MapCodeToDescription(code string) string {
	codeMap := map[string]string{
		"INVALID_CREDENTIALS":   "Invalid login or password",
		"VALIDATION_ERROR":      "Some fields are invalid",
		"INTERNAL_SERVER_ERROR": "Internal server error",
		"NOT_FOUND":             "Resource not found",
		"INVALID_AUTH_HEADER":   "Invalid authorization header",
		"USER_NOT_FOUND":        "User not found",
		"INVALID_TOKEN":         "Invalid token",
	}

	description, ok := codeMap[code]
	if !ok {
		return "Unknown error"
	}

	return description
}
