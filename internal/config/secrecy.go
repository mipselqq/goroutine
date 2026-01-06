package config

import "fmt"

func hideStringContents(str string) string {
	return fmt.Sprintf("(%d chars)", len(str))
}
