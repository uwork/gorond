package notify

import (
	"fmt"
)

func SendStdout(subject string, body string) {
	fmt.Printf("%s: %s\n", subject, body)
}
