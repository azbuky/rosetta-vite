package cmd

import (
	"github.com/azbuky/rosetta-vite/vite"

	"github.com/spf13/cobra"
)

var (
	utilsSignCmd = &cobra.Command{
		Use:   "utils:sign",
		Short: "Sign message with private key",
		Long: `Sign message with private key.

When calling this command, you must provide 2 arguments:
[1] the private key to sign with, in hex format
[2] the message to sign, in hex format`,
		RunE: runUtilsSignCmd,
		Args: cobra.ExactArgs(2), //nolint:gomnd
	}
)

func runUtilsSignCmd(cmd *cobra.Command, args []string) error {
	return vite.SignData(args[0], args[1])
}
