package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	common "gorex/pkg/common"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate example scan configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		o, err := cmd.Flags().GetString(fOutput)

		if err != nil {
			return err
		}

		if err := gen(o); err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	},
}

const (
	fOutput = "output"
)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func writeScopeConfiguration(config common.ScanConfig, p string) error {

	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p, b, os.ModePerm)
}

func gen(o string) error {
	log.Printf("*** Start generate example file \n*** Output file path : %v", o)
	defer log.Println("*** End ***")

	var scopes []common.ScopeConfig

	squery := []string{"^\\s*COMMAND\\=.*$"}
	scopes = append(scopes, common.ScopeConfig{
		Name:            "example-find-any-command-in-scope",
		StartQuery:      "^\\W*BEGIN$",
		FinishQuery:     "^\\W*END$",
		SearchQuery:     squery,
		SearchQueryMode: common.SearchQueryOperatorAny,
	})
	squery = []string{
		"^\\s*COMMAND\\=AAA$",
		"^\\s*COMMAND\\=BBB$"}
	scopes = append(scopes, common.ScopeConfig{
		Name:            "example-find-two-commands-in-scope",
		StartQuery:      "^\\W*BEGIN$",
		FinishQuery:     "^\\W*END$",
		SearchQuery:     squery,
		SearchQueryMode: common.SearchQueryOperatorAll,
	})

	cfg := common.ScanConfig{
		Folder: ".\\example",
		Filter: "*.txt",
		Scopes: scopes,
	}

	if err := cfg.IsValid(); err != nil {
		log.Printf("Config is not valid: %v\n", err)
		return err
	}

	return writeScopeConfiguration(cfg, o)
}

func init() {

	genCmd.Flags().StringP(fOutput, "o", ".\\example.json", "Output configuration.")
	rootCmd.AddCommand(genCmd)
}
