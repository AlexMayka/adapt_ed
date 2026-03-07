package main

import (
	"backend/internal/config"
	"fmt"
)

func main() {
	cnf, err := config.Load()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", cnf)
}
