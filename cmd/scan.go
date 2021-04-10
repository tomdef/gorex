package cmd

import (
	"gorex/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan with advanced regex configurations",

		RunE: func(cmd *cobra.Command, args []string) error {

			if err := scan(input, outputHTML, outputJSON, trace, show); err != nil {
				return err
			}
			return nil
		},
	}

	// Commands represents path to command file
	input      string
	outputHTML string
	outputJSON string
	show       bool
	trace      bool
)

const (
	fInput      = "input"
	fOutputHTML = "outputhtml"
	fShow       = "show"
	fTrace      = "trace"
)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func init() {

	scanCmd.Flags().StringVarP(&input, "input", "i", ".", "Input file path (*.json) with scan commands.")
	scanCmd.Flags().StringVarP(&outputHTML, fOutputHTML, "o", "", "Output html report.")
	scanCmd.Flags().BoolVarP(&trace, fTrace, "t", false, "Set trace mode.")
	scanCmd.Flags().BoolVarP(&show, fShow, "s", false, "Show result after scan.")

	rootCmd.AddCommand(scanCmd)
}

func scan(input string, outputhtml string, outputjson string, trace bool, show bool) error {

	durationStart := time.Now()

	logger := utils.CreateLogger("scan", trace)
	logger.Info().Msgf("START SCAN. Command(s) file path : %v", input)
	// -->

	// <--
	elapsed := time.Since(durationStart)
	logger.Info().Msgf("FINISH SCAN after %s", elapsed)

	return nil
}
