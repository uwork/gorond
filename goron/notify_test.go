package goron

import (
	"gorond/config"
)

func ExampleNotifyStdout() {
	conf := createTestConfig()
	job := &config.Job{Command: "echo output notify"}

	Notify(conf, "output notify", 0, nil, job)

	// Output:
	// notify successful: command: echo output notify
	// output: output notify
	// error: <nil>
}
