package main

import (
	"fmt"
	"go-tg-transcriber/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
