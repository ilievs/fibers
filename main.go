package main

import (
	"github.com/ilievs/fibers/system"
)

func main() {

	RunApplication()

	system.WaitForOsSignal()
	// Cleanup
}
