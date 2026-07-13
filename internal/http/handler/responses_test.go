package handler_test

import (
	"goroutine/internal/testutil"
)

func columnNotFoundError(field string) map[string]any {
	return map[string]any{
		"code":      "COLUMN_NOT_FOUND",
		"message":   "Column not found",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": field, "issues": []string{"Column not found"}},
		},
	}
}

func taskNotFoundError(field string) map[string]any {
	return map[string]any{
		"code":      "TASK_NOT_FOUND",
		"message":   "Task not found",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": field, "issues": []string{"Task not found"}},
		},
	}
}

func payloadTooLargeError() map[string]any {
	return map[string]any{
		"code":      "PAYLOAD_TOO_LARGE",
		"message":   "Request body too large",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": "body", "issues": []string{"Please stop spamming >_<"}},
		},
	}
}

func invalidJSONError() map[string]any {
	return map[string]any{
		"code":      "VALIDATION_ERROR",
		"message":   "Some fields are invalid",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
		},
	}
}

func internalError() map[string]any {
	return map[string]any{
		"code":      "INTERNAL_SERVER_ERROR",
		"message":   "Internal server error",
		"timestamp": testutil.FixedNowStr(),
	}
}

func boardNotFoundError() map[string]any {
	return map[string]any{
		"code":      "BOARD_NOT_FOUND",
		"message":   "Board not found",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": "boardId", "issues": []string{"Board not found"}},
		},
	}
}

func userAlreadyExistsError() map[string]any {
	return map[string]any{
		"code":      "USER_ALREADY_EXISTS",
		"message":   "User already exists",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": "email", "issues": []string{"Email already registered"}},
		},
	}
}

func unauthorizedTokenError() map[string]any {
	return map[string]any{
		"code":      "INVALID_TOKEN",
		"message":   "Invalid token",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
		},
	}
}

func validationError(field string, issues []string) map[string]any {
	return map[string]any{
		"code":      "VALIDATION_ERROR",
		"message":   "Some fields are invalid",
		"timestamp": testutil.FixedNowStr(),
		"details": []any{
			map[string]any{"field": field, "issues": issues},
		},
	}
}
