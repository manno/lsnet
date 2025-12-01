package internal

import (
	"fmt"
)

// Run is the main entry point for the netree application
func Run(opts *Options) error {
	// Normalize columns
	opts.Columns = NormalizeColumns(opts.Columns)

	// Discover interfaces
	interfaces, err := DiscoverInterfaces(opts)
	if err != nil {
		return fmt.Errorf("failed to discover interfaces: %w", err)
	}

	// Apply filters
	interfaces = FilterInterfaces(interfaces, opts)

	if len(interfaces) == 0 {
		if !opts.NoHeadings {
			fmt.Println("No interfaces found")
		}
		return nil
	}

	// Output format
	if opts.ListFormat {
		// Flat list output
		return OutputList(interfaces, opts)
	}

	// Build tree
	roots := BuildTree(interfaces, opts)

	// Output tree
	return OutputTree(roots, opts)
}
