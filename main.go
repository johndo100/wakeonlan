package main

import (
	"fmt"

	wakeonlan "github.com/johndo100/wakeonlan/pkg/magic"
)

func main() {
	// example
	err := wakeonlan.SendMagic("12-34-56-78-9A-BC", "", "123.456.789.101", "9")
	if err != nil {
		fmt.Println(err)
	}
}
