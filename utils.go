package main

import (
	"fmt"
	"os"
)

func printError(err error) {
	os.Stderr.WriteString(fmt.Sprintf("[ERROR] %s.\n", err))
}

func createDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0744); err != nil {
			return err
		}
	}
	return nil
}
