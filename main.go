package main

import (
	"os"

	"fmt"

	"github.com/jakecoffman/stldevs/web"
)

func main() {
	// check prereqs
	if os.Getenv("GITHUB_API_KEY") == "" {
		fmt.Println("please set GITHUB_API_KEY")
		return
	}

	web.Run()
}
