package cmd

import (
	"fmt"
	"os"

	"github.com/pasc4le-ai-sandbox/pocket/pocket"
	"github.com/spf13/cobra"
)

var (
	release bool
	list    bool
	cut     bool
	keep    bool
	deleteI int
)

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "pocket [flags] <file-or-dir> ...",
	Short: "A dead-simple file clipboard for the terminal",
	Long: `pocket is a file-based clipboard for your terminal.

Add files or directories to the clipboard with:

  pocket foo.txt bar/ baz.pdf

Then release them (copy to current directory):

  pocket --release

Or move them instead of copying:

  pocket --release --cut

List the current clipboard:

  pocket --list

Remove an item (by its list number) from the clipboard:

  pocket --delete 2

Copyright (c) 2026  Giuseppe Pascale
Licensed under the European Union Public Licence v1.2 (EUPL-1.2)
<https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12>
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Mode selection: exactly one of: add, release, list, delete.
		var mode string
		switch {
		case release:
			mode = "release"
		case list:
			mode = "list"
		case deleteI >= 0:
			mode = "delete"
		case len(args) > 0:
			mode = "add"
		default:
			return fmt.Errorf("pocket requires at least one argument or a flag; see pocket --help")
		}

		switch mode {
		case "release":
			if err := pocket.Release(cut, keep); err != nil {
				return err
			}
		case "list":
			items, err := pocket.List()
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Println("(empty)")
				return nil
			}
			for i, item := range items {
				fmt.Printf("%d  %s\n", i+1, item)
			}
		case "delete":
			if err := pocket.Delete(deleteI); err != nil {
				return err
			}
		case "add":
			if err := pocket.Add(args...); err != nil {
				return err
			}
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&release, "release", "r", false, "release (copy) all clipboard items to current directory")
	rootCmd.Flags().BoolVarP(&list, "list", "l", false, "list all clipboard items with numbers")
	rootCmd.Flags().IntVarP(&deleteI, "delete", "d", -1, "remove item NUMBER from clipboard (1-indexed)")
	rootCmd.Flags().BoolVarP(&cut, "cut", "c", false, "move files instead of copying (only with --release)")
	rootCmd.Flags().BoolVarP(&keep, "keep", "k", false, "keep clipboard items after release (don't clear)")

	// Make -d explicit: require a value for the short flag too
	// cobra already does this by default for IntVarP
}
