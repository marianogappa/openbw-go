package main

import (
	"fmt"
	"os"

	"github.com/marianogappa/openbw-go/downgrade"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: downgrade in.rep out.rep")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "downgrade:", err)
		os.Exit(1)
	}
}

func run(inPath, outPath string) error {
	in, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	out, report, err := downgrade.Downgrade(in)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return err
	}
	printReport(report)
	return nil
}

func printReport(report downgrade.Report) {
	fmt.Fprintf(os.Stderr, "source format: %s\n", report.SourceFormat)
	for _, change := range report.Changes {
		fmt.Fprintf(os.Stderr, "changed: %s\n", change)
	}
	for _, d := range report.Dropped {
		fmt.Fprintf(os.Stderr, "dropped: id %#02x, frame %d, player %d, count %d (%s)\n", d.ID, d.Frame, d.PlayerID, d.Count, d.Reason)
	}
}
