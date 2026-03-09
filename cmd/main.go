package main

import (
	"backend/internal/config"
	"fmt"
)

// main loads runtime configuration and prints it for local sanity-check runs.
func main() {
	cnf, err := config.Load()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", cnf)
}
