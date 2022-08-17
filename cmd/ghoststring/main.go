package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rstudio/ghoststring"
)

func main() {
	keyFlag := flag.String("k", "", "key to use in ghostifying")
	decryptFlag := flag.Bool("d", false, "decrypt input")
	namespaceFlag := flag.String("n", "default", "namespace to use in ghostifying")

	flag.Parse()

	ghostifyer, err := ghoststring.SetGhostifyer(*namespaceFlag, *keyFlag)
	if err != nil {
		log.Fatal(err)
	}

	inStringBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if *decryptFlag {
		gs, err := ghostifyer.Unghostify(string(inStringBytes))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, gs.String)
		return
	}

	encString, err := ghostifyer.Ghostify(
		&ghoststring.GhostString{
			Namespace: *namespaceFlag,
			String:    string(inStringBytes),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, encString)
}
