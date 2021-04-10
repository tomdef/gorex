package cmd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	common "gorex/pkg/common"
	"gorex/pkg/utils"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate example scan configuration",

		RunE: func(cmd *cobra.Command, args []string) error {

			logger := utils.CreateLogger("gen", trace)
			o, err := cmd.Flags().GetString(fOutput)

			if err != nil {
				return err
			}

			if err := gen(o, logger); err != nil {
				logger.Fatal().Err(err)
				return err
			}
			return nil
		},
	}
)

const (
	fOutput = "output"
)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func writeScopeConfiguration(config common.ScanConfig, p string) error {

	isXML := strings.HasSuffix(p, ".xml")
	isJSON := strings.HasSuffix(p, ".json")

	if !isXML && !isJSON {
		return errors.New("invalid file extension. Should be json or xml")
	}

	var b []byte
	var e error

	if isJSON {
		b, e = json.MarshalIndent(config, "", "\t")
	} else {
		b, e = xml.MarshalIndent(config, "", "\t")
	}

	if e != nil {
		return e
	}

	return ioutil.WriteFile(p, b, os.ModePerm)
}

func gen(o string, logger zerolog.Logger) error {
	logger.Info().Msgf("Start generate example file. Output file path : %v", o)
	defer logger.Info().Msg("End")

	cfg := common.ScanConfig{
		Folder: ".\\example",
		Filter: "*.txt",
		Scopes: []common.ScopeConfig{
			{
				Name:      "example-1",
				Begin:     "^\\W*BEGIN$",
				End:       "^\\W*END$",
				AutoClose: true,
				Matches: []common.MatchConfig{
					{
						Name:       "match-1",
						Match:      "^\\s*COMMAND\\=.*$",
						Occurrence: "OnceOrMore",
					},
				},
			},
		},
	}

	if err := cfg.IsValid(); err != nil {
		logger.Error().Msgf("Config is not valid: %v", err)
		return err
	}

	return writeScopeConfiguration(cfg, o)
}

func init() {

	genCmd.Flags().StringP(fOutput, "o", ".\\example.json", "Output configuration.")
	rootCmd.AddCommand(genCmd)
}
