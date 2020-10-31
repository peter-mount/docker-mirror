package mirror

import (
	"flag"
	"fmt"
	"syscall"
)

var verbose *bool

func Main() error {
	fmt.Println("Docker Mirror 0.01a")

	verbose = flag.Bool("v", false, "Verbosity")
	dest := flag.String("d", "", "Destination repository")

	flag.Parse()

	if *dest == "" {
		fmt.Println("-d is required")
		syscall.Exit(1)
	}

	fmt.Printf("Using mirror %v\n", *dest)

	for _, image := range flag.Args() {
		err := MirrorImage(image, *dest)
		if err != nil {
			return err
		}
	}

	return nil
}
