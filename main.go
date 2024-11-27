package main

import (
	"github.com/hyunwoododev/golang-for-blockchain/cli"
	"github.com/hyunwoododev/golang-for-blockchain/db"
)

func main() {
	defer db.Close()
	cli.Start()
}
