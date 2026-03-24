package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, build date, and other information about gxpdf.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("gxpdf %s\n", Version)
		fmt.Printf("  Go:         %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		if GitCommit != unknownValue && GitCommit != "" {
			fmt.Printf("  Commit:     %s\n", GitCommit)
		}
		if BuildDate != unknownValue && BuildDate != "" {
			fmt.Printf("  Built:      %s\n", BuildDate)
		}
	},
}
