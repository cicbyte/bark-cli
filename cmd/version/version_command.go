package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("bark-cli %s\n", Version)
			fmt.Printf("  commit: %s\n", GitCommit)
			fmt.Printf("  built:  %s\n", BuildTime)
			return nil
		},
	}
}
