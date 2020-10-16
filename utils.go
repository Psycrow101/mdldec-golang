package main

import (
	"fmt"
	"os"
)

func printError(err error) {
	os.Stderr.WriteString(fmt.Sprintf("[ERROR] %s.\n", err))
}
