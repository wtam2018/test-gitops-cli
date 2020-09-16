package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended environment command name.
const RecommendedCommandName = "version"

var Version string

// NewCmd creates a new environment command
func NewCmd(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Get version",
		Long:  "Get command version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gitops version %s\n", Version)
		},
	}
}
