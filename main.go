package main

import (
	"github.com/hyunwoo.do/go-coin/cli"
	"github.com/hyunwoo.do/go-coin/db"
)

func main() {
	defer db.Close()
	cli.Start()
}
