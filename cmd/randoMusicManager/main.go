package main

import (
	"flag"
	"fmt"

	"github.com/frogssoldseparately/randoMusicManager/pkg/parser"
)

func main() {
	src := flag.String("src", "", "Source manifest file")
	dest := flag.String("dest", "", "Destination folder")
	flag.Parse()
	if len(*src) > 0 {
		if len(*dest) > 0 {
			parser.Setup()
			if err := parser.ParseAndExecute(*src, *dest); err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Expected dest path")
		}
	} else {
		fmt.Println("Expected src path")
	}
}
