package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/manno/lsnet/internal"
)

const version = "0.1.0"

var (
	showAll       = flag.Bool("a", false, "show all interfaces including DOWN")
	allNamespaces = flag.Bool("N", false, "show interfaces from all network namespaces")
	outputCols    = flag.String("o", "", "output columns (comma-separated)")
	listColumns   = flag.Bool("list-columns", false, "list all available columns")
	directionDown = flag.Bool("d", false, "tree direction: logical devices down")
	directionUp   = flag.Bool("u", false, "tree direction: physical devices up (default)")
	listFormat    = flag.Bool("l", false, "list format (no tree)")
	jsonOutput    = flag.Bool("J", false, "JSON output")
	noHeadings    = flag.Bool("n", false, "don't print column headers")
	excludeTypes  = flag.String("x", "", "exclude interface types (comma-separated)")
	filterTypes   = flag.String("t", "", "show only specified types (comma-separated)")
	showVersion   = flag.Bool("v", false, "show version")
)

// preprocessArgs handles -o+COLUMNS syntax by converting it to -o +COLUMNS
func preprocessArgs() {
	for i, arg := range os.Args {
		// Handle -o+COLUMNS (no space)
		if strings.HasPrefix(arg, "-o+") {
			cols := strings.TrimPrefix(arg, "-o+")
			os.Args[i] = "-o"
			// Insert the +COLUMNS as next argument
			os.Args = append(os.Args[:i+1], append([]string{"+" + cols}, os.Args[i+1:]...)...)
			return
		}
		// Handle --output+COLUMNS (no space)
		if strings.HasPrefix(arg, "--output+") {
			cols := strings.TrimPrefix(arg, "--output+")
			os.Args[i] = "-o"
			os.Args = append(os.Args[:i+1], append([]string{"+" + cols}, os.Args[i+1:]...)...)
			return
		}
	}
}

func main() {
	// Preprocess args to support -o+COLUMNS syntax (like lsblk)
	preprocessArgs()

	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("lsnet version %s\n", version)
		os.Exit(0)
	}

	if *listColumns {
		internal.PrintAvailableColumns()
		os.Exit(0)
	}

	// Determine columns to display
	columns := internal.DefaultColumns
	if *outputCols != "" {
		if strings.HasPrefix(*outputCols, "+") {
			// Append to defaults
			appendCols := strings.TrimPrefix(*outputCols, "+")
			columns = append(columns, strings.Split(appendCols, ",")...)
		} else {
			// Replace defaults
			columns = strings.Split(*outputCols, ",")
		}
	}

	// Validate columns
	if err := internal.ValidateColumns(columns); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := &internal.Options{
		ShowAll:       *showAll,
		AllNamespaces: *allNamespaces,
		DirectionDown: *directionDown,
		Columns:       columns,
		ListFormat:    *listFormat,
		JSONOutput:    *jsonOutput,
		NoHeadings:    *noHeadings,
		ExcludeTypes:  parseTypeList(*excludeTypes),
		FilterTypes:   parseTypeList(*filterTypes),
	}

	if err := internal.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `lsnet - Network Device Tree Viewer

Usage: lsnet [options]

Options:
  -a, --all              show all interfaces including DOWN
  -N, --all-namespaces   show interfaces from all network namespaces
  -o, --output <list>    output columns (comma-separated, or +COL to append)
  --list-columns         list all available columns
  -d                     tree direction: logical devices down
  -u                     tree direction: physical devices up (default)
  -l, --list             list format (no tree)
  -J, --json             JSON output
  -n, --noheadings       don't print column headers
  -t, --type <types>     show only specified types (comma-separated)
  -x, --exclude <types>  exclude specified types (comma-separated)
  -v, --version          show version

Examples:
  lsnet                      # show active interfaces in current namespace
  lsnet -a                   # show all interfaces (including DOWN)
  lsnet -o NAME,TYPE,DRIVER  # custom columns
  lsnet -o+MAC,MTU           # append columns to defaults
  lsnet -N                   # show all namespaces
  lsnet -t bridge,veth       # show only bridges and veth interfaces

Default columns: NAME,TYPE,STATE
Note: IP addresses are shown as child nodes in the tree
`)
}

func parseTypeList(s string) []string {
	if s == "" {
		return nil
	}
	types := strings.Split(s, ",")
	for i := range types {
		types[i] = strings.TrimSpace(types[i])
	}
	return types
}
