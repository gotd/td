package message

import "fmt"

func formatMessage(msg string, args ...interface{}) string {
	return fmt.Sprintf(msg, args...)
}
