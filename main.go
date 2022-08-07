package main

import (
	"maintenance/maintenance"
)

func main() {
	command := maintenance.ReadCommand()
	command.Run()
}
