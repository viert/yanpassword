package main

import (
	"github.com/viert/yanpassword/manager"
)

func main() {
	m, err := manager.NewManager()
	if err != nil {
		panic(err)
	}
	err = m.Start()
	if err != nil {
		panic(err)
	}
}
