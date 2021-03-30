package common

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"
)

//go:embed htmlPattern.html
var htmlPattern string

// -----------------------------------------------------------------------------
// types
// -----------------------------------------------------------------------------

// SearchQueryOperator describe types of operator ...
type SearchQueryOperator int

const (
	//SearchQueryOperatorAll is AND operator for queries
	SearchQueryOperatorAll SearchQueryOperator = iota
	//SearchQueryOperatorAny is OR operator for queries
	SearchQueryOperatorAny
	//SearchQueryOperatorStrictOrder ...
	SearchQueryOperatorStrictOrder
)

// Scan config structs :

// ScopeConfig provides configuration of scan
type ScopeConfig struct {
	Name                 string              `json:"name" xml:"name,attr"`
	StartQuery           string              `json:"startQuery" xml:"startQuery"`
	FinishQuery          string              `json:"finishQuery" xml:"finishQuery"`
	StartQueryCloseScope bool                `json:"startQueryCloseScope" xml:"startQueryCloseScope"`
	SearchQuery          []string            `json:"searchQuery" xml:"searchQuery"`
	SearchQueryMode      SearchQueryOperator `json:"searchQueryMode" xml:"searchQueryMode"`
}

// ScanConfig provides scan configuration
type ScanConfig struct {
	Folder string        `json:"folder" xml:"folder,attr"`
	Filter string        `json:"filter" xml:"filder,attr"`
	Scopes []ScopeConfig `json:"scopes" xml:"scopes"`
}

// Scan summary structs :

// FileScopeSummary provides...
type FileScopeSummary struct {
	FileName   string         `json:"fileName" xml:"fileName,attr"`
	Scopes     []ScopeSummary `json:"scopes" xml:"scopes"`
	AllMatches int
}

// MatchLine provides...
type MatchLine struct {
	Line  string `json:"line" xml:"line,attr"`
	Index int    `json:"index" xml:"index,attr"`
}

// ScopeSummary provides...
type ScopeSummary struct {
	Name          string   `json:"name" xml:"name,attr"`
	FileName      string   `json:"fileName" xml:"fileName,attr"`
	Started       int      `json:"started" xml:"started,attr"`
	Finished      int      `json:"finished" xml:"finished,attr"`
	Content       []string `json:"content" xml:"content"`
	ContentAsHTML []string
	Matches       []MatchLine `json:"matches" xml:"matches"`
}

// ScanSummary provides...
type ScanSummary struct {
	Folder       string
	Filter       string
	CreationTime time.Time
	Summary      []FileScopeSummary
	ScanFiles    int
}

// ScopeSummaryWithConfig provides...
type ScopeSummaryWithConfig struct {
	ScopeSummary ScopeSummary
	ScopeConfig  ScopeConfig
}

// ReadScopeConfiguration read ScanConfig from file
func ReadScopeConfiguration(configPath string) (ScanConfig, error) {

	jsonFile, err := os.Open(configPath)
	if err != nil {
		return ScanConfig{}, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return ScanConfig{}, err
	}

	var scanConfig ScanConfig

	if json.Unmarshal(byteValue, &scanConfig) != nil {
		return ScanConfig{}, err
	}
	return scanConfig, nil
}

// -----------------------------------------------------------------------------
// extensions
// -----------------------------------------------------------------------------

// IsValid check if ScanConfig contains every required fields
func (cfg ScanConfig) IsValid() error {

	if cfg.Folder == "" {
		return errors.New("Empty folder")
	}

	if cfg.Filter == "" {
		return errors.New("Empty filter")
	}

	if len(cfg.Scopes) == 0 {
		return errors.New("Empty scopes")
	}

	for i, v := range cfg.Scopes {
		if v.Name == "" {
			return fmt.Errorf("Empty name of scope [%v]", i)
		}
		if len(v.SearchQuery) == 0 {
			return fmt.Errorf("Empty search queries of scope [%v]", i)
		}

		for j, q := range v.SearchQuery {
			if q == "" {
				return fmt.Errorf("Empty #%v search query of scope [%v]", j, i)
			}
		}

		if (v.StartQuery != "") && (v.FinishQuery == "") {
			return fmt.Errorf("Empty finish query of scope [%v]", i)
		}
		if (v.StartQuery == "") && (v.FinishQuery != "") {
			return fmt.Errorf("Empty start query of scope [%v]", i)
		}
	}

	return nil
}

// LogToFile writes summary to json file
func (s ScanSummary) LogToFile(p string) error {
	file, _ := json.MarshalIndent(s, "", " ")

	return ioutil.WriteFile(p, file, 0644)
}

// LogToHTML generate html log file
func (s ScanSummary) LogToHTML(p string) error {
	if s.Summary == nil {
		return errors.New("scan summary does not contains any summaries")
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New("template").Parse(htmlPattern)
	if err != nil {
		return err
	}
	err = t.Execute(f, s)
	if err != nil {
		return err
	}
	return nil
}
