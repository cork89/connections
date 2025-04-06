package main

import (
	"os"
	"testing"
)

func setup() {
	// main()
	badwords = []string{"badword"}
}

func TestMain(m *testing.M) {
	// Setup code here
	setup()

	// Run the tests
	exitCode := m.Run()
	// Teardown code here
	// teardown()
	os.Exit(exitCode)
}
