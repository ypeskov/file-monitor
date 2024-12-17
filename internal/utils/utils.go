package utils

import (
	"os"
)

/*
	parseArgs returns the list of directories to monitor.

If no directories are provided, it logs an error and exits.
*/
func ParseArgs() []string {
	if len(os.Args) < 2 {
		return []string{"."}
	}
	return os.Args[1:]
}
