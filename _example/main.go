package main

import (
	"encoding/json"
	"fmt"

	"github.com/wythers/algalon"
)

type Storage struct{}

func (s Storage) Write(r algalon.Record) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func main() {
	al := algalon.Default(Storage{})

	al.Run(":8080")
}
