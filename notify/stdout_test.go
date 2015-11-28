package notify

func ExampleSendStdout() {
	SendStdout("subject", "body")

	// Output:
	// subject: body
}
