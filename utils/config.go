package utils

import (
	"fmt"
	"os"
)

// setOption takes 'option' (opt) and 'default' (def) values, and returns the
// option to use (either the provided option, or the default)
func SetOption(opt, def string) string {

	if opt == "" {
		if def == "" {
			fmt.Printf("WARNING: No option provided and missing default, unable to proceed aborting...")
			os.Exit(1)
		}

		return def
	}

	return opt
}
