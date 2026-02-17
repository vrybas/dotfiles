package main

import (
	"fmt"
	"os"
	"strings"
	_ "strings"

	"github.com/bxcodec/faker/v3"
	_ "github.com/bxcodec/faker/v3"
)

// Generate UUID
//
// Usage:
// 		no arguments:
//      c1322c4bc8104965a1dd6f8afd0de10c
// 		-h:
//      23a0e19f-7d8a-4fa5-bb3c-186eb6ea6d8c
func main() {
	var str string

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "-h" {
		str = strings.ToLower(faker.UUIDHyphenated())
	} else {
		str = strings.ToLower(faker.UUIDDigit())
	}

	fmt.Println(str)
}
