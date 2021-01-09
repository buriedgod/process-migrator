package main

import (
	"fmt"

	criu "github.com/checkpoint-restore/go-criu"
)

func main() {
	c := criu.MakeCriu()
	version, err := c.GetCriuVersion()
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println(version)
}
