package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Println("subspace-assignment (scaffold)\n\nThis repository will implement an educational Rod-based automation PoC that ONLY runs against the included local Mock Social Network.\n\nNext: run `make lint` and `make test`, then follow README Quickstart.")
		return
	}

	fmt.Println("subspace-assignment scaffold. Try: `go run ./cmd/subspace-assignment --help`")
}
