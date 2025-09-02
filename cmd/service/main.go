package main

import (
	"fmt"
	"os"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/di"
)

func main() {
	if err := di.RunService(); err != nil {
		if _, writeErr := fmt.Fprintf(os.Stderr, "Application failed: %v\n", err); writeErr != nil {
			panic(fmt.Sprintf("Application failed: %v (stderr write error: %v)", err, writeErr))
		}
		os.Exit(1)
	}
}
