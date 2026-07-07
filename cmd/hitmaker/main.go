package main

import "github.com/kerns/hitmaker/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
