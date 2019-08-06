package main

import (
	"manager"
)

func main() {
	m, err := manager.NewManager()
	if err != nil {
		panic(err)
	}
	m.Start()
}
