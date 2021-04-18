package common

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/dlclark/regexp2"
)

const (
	MustMatchToAll       = "*"
	MustMatchToOneOrMore = "+"
	MustNotMatchToAny    = "!"
)

//go:embed htmlPattern.html
var htmlPattern string

// -----------------------------------------------------------------------------
// Types
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// Scan config structs :

// MatchConfig provides configuration of single match
type MatchConfig struct {
	Name           string `json:"name" xml:"name,attr"`
	Match          string `json:"match" xml:"match,attr"`
	IgnoreInResult bool   `json:"ignoreInResult" xml:"ignoreInResult,attr"`
	Occurrence     string `json:"occurrence" xml:"occurrence,attr"`
}

// IsValid verified MatchConfig settings
func (cfg MatchConfig) IsValid() error {

	if cfg.Occurrence == "*" || cfg.Occurrence == "!" || cfg.Occurrence == "+" {
		return nil
	} else {
		if _, err := strconv.Atoi(cfg.Occurrence); err != nil {
			return fmt.Errorf("match %v has not valid occurrence [%v]", cfg.Name, cfg.Occurrence)
		}
	}

	return nil
}

// IsMatchOccurrence check if string collection is enought
func (cfg MatchConfig) IsMatch(matchLines []MatchLine) []int {

	allLinesCount := len(matchLines)
	var allMatchedLines []string
	var result []int

	rx := regexp2.MustCompile(cfg.Match, regexp2.Singleline)

	for _, l := range matchLines {
		if m, e := rx.MatchString(l.Line); m && e == nil {
			allMatchedLines = append(allMatchedLines, l.Line)
			result = append(result, l.Index)
		}
	}

	if cfg.IsMatchOccurrence(allMatchedLines, allLinesCount) {
		return result
	} else {
		return []int{}
	}
}

// IsMatchOccurrence check if string collection is enought
func (cfg MatchConfig) IsMatchOccurrence(lines []string, allLinesCount int) bool {

	if cfg.Occurrence == "*" && len(lines) == allLinesCount {
		return true
	}

	if cfg.Occurrence == "+" && len(lines) >= 1 {
		return true
	}

	if cfg.Occurrence == "!" && len(lines) == 0 {
		return true
	}

	if v, err := strconv.Atoi(cfg.Occurrence); err == nil && v == len(lines) {
		return true
	}

	return false
}

// ScopeConfig provides configuration of scan
type ScopeConfig struct {
	Name      string        `json:"name" xml:"name,attr"`
	Begin     string        `json:"begin" xml:"begin"`
	End       string        `json:"end" xml:"end"`
	AutoClose bool          `json:"autoClose" xml:"autoClose"`
	Matches   []MatchConfig `json:"matches" xml:"matches"`
}

// IsValid verified ScopeConfig settings
func (cfg ScopeConfig) IsValid() error {

	if cfg.Name == "" {
		return fmt.Errorf("scope %v name is empty", cfg)
	}

	if cfg.Begin == "" {
		return fmt.Errorf("scope %v begin expression is empty", cfg)
	}

	if cfg.End == "" {
		return fmt.Errorf("scope %v end expression is empty", cfg)
	}

	if len(cfg.Matches) == 0 {
		return fmt.Errorf("scope %v matches list is empty", cfg)
	}

	for i, v := range cfg.Matches {
		if e := v.IsValid(); e != nil {
			return fmt.Errorf("scope %v matche #%v in not valid: %v", cfg, i, e.Error())
		}
	}

	return nil
}

// ScanConfig provides scan configuration
type ScanConfig struct {
	Folder string        `json:"folder" xml:"folder,attr"`
	Filter string        `json:"filter" xml:"filder,attr"`
	Scopes []ScopeConfig `json:"scopes" xml:"scopes"`
}

// -----------------------------------------------------------------------------
// Scan summary structs :

// MatchLine provides...
type MatchLine struct {
	Line       string   `json:"line" xml:"line,attr"`
	Index      int      `json:"index" xml:"index,attr"`
	MatchNames []string `json:"matchNames" xml:"matchNames,attr"`
}

// ScopeSummary provides...
type ScopeSummary struct {
	Id            string   `json:"id" xml:"id,attr"`
	Name          string   `json:"name" xml:"name,attr"`
	FileName      string   `json:"fileName" xml:"fileName,attr"`
	Started       int      `json:"started" xml:"started,attr"`
	Finished      int      `json:"finished" xml:"finished,attr"`
	Content       []string `json:"content" xml:"content"`
	ContentAsHTML []string
	Matches       []MatchLine `json:"matches" xml:"matches"`
}

// FileScopeSummary provides...
type FileScopeSummary struct {
	FileName   string         `json:"fileName" xml:"fileName,attr"`
	Scopes     []ScopeSummary `json:"scopes" xml:"scopes"`
	AllMatches int
}

// ScanSummary provides...
type ScanSummary struct {
	Folder       string
	Filter       string
	Scopes       []ScopeConfig
	CreationTime time.Time
	DurationTime time.Duration
	Summary      []FileScopeSummary
	ScanFiles    int
}

// ScopeSummaryWithConfig provides...
type ScopeSummaryWithConfig struct {
	ScopeSummary ScopeSummary
	ScopeConfig  ScopeConfig
}

// -----------------------------------------------------------------------------
// Extensions
// -----------------------------------------------------------------------------

func (s *ScopeSummary) ResolveId() error {
	guid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, guid)
	if n != len(guid) || err != nil {
		return err
	}
	guid[8] = guid[8]&^0xc0 | 0x80
	guid[6] = guid[6]&^0xf0 | 0x40

	s.Id = fmt.Sprintf("%x", guid[0:4])
	return nil
}

//Read ScopeConfiguration read ScanConfig from content
func ReadScanConfiguration(content []byte) (ScanConfig, error) {

	var scanConfig ScanConfig

	if e := json.Unmarshal(content, &scanConfig); e != nil {
		return ScanConfig{}, e
	}

	if e := scanConfig.IsValid(); e != nil {
		return ScanConfig{}, e
	}

	return scanConfig, nil
}

//ReadScanConfiguration read ScanConfig from file path
func ReadScanConfigurationFromFile(name string) (ScanConfig, error) {

	jsonConfigFile, err := os.Open(name)
	if err != nil {
		return ScanConfig{}, err
	}
	defer jsonConfigFile.Close()

	byteValue, err := ioutil.ReadAll(jsonConfigFile)

	if err != nil {
		return ScanConfig{}, err
	}

	return ReadScanConfiguration(byteValue)
}

//ReadScopeConfiguration read ScanConfig from content
func (cfg ScanConfig) WriteScanConfiguration() ([]byte, error) {

	if e := cfg.IsValid(); e != nil {
		return nil, e
	}

	return json.Marshal(cfg)
}

// -----------------------------------------------------------------------------
// extensions
// -----------------------------------------------------------------------------

// IsValid check if ScanConfig contains every required fields
func (cfg ScanConfig) IsValid() error {

	if cfg.Folder == "" {
		return fmt.Errorf("scan %v folder is empty", cfg)
	}

	if cfg.Filter == "" {
		return fmt.Errorf("scan %v filter is empty", cfg)
	}

	if len(cfg.Scopes) == 0 {
		return fmt.Errorf("scan %v does not have scopes", cfg)
	}

	for _, v := range cfg.Scopes {

		if e := v.IsValid(); e != nil {
			return e
		}
	}

	return nil
}

// LogToHTML generate html summary file
func (s ScanSummary) WriteAsHTML(wr io.Writer) error {

	if s.Summary == nil {
		return fmt.Errorf("scan summary does not contains any summaries")
	}

	t, err := template.New("template").Parse(htmlPattern)
	if err != nil {
		return fmt.Errorf("cannot read html template. %v", err.Error())
	}

	err = t.Execute(wr, s)
	if err != nil {
		return fmt.Errorf("cannot create html content. %v", err.Error())
	}

	return nil
}

// LogToHTML generate html summary file
func (s ScanSummary) AsHTML() ([]byte, error) {

	if s.Summary == nil {
		return nil, fmt.Errorf("scan summary does not contains any summaries")
	}

	t, err := template.New("template").Parse(htmlPattern)
	if err != nil {
		return nil, fmt.Errorf("cannot read html template. %v", err.Error())
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, s)
	if err != nil {
		return nil, fmt.Errorf("cannot create html content. %v", err.Error())
	}

	return buffer.Bytes(), nil
}

// AsJSON generate json summary file
func (s ScanSummary) AsJSON() ([]byte, error) {

	if s.Summary == nil {
		return []byte{}, fmt.Errorf("scan summary does not contains any summaries")
	}

	return json.MarshalIndent(s, "", " ")
}

// LogToFile generate json summary file
// func (s ScanSummary) LogToFile(p string) error {
// 	file, _ := json.MarshalIndent(s, "", " ")

// 	return ioutil.WriteFile(p, file, 0644)
// }

// // LogToHTML generate html summary file
// func (s ScanSummary) LogToHTML(p string) error {
// 	if s.Summary == nil {
// 		return errors.New("scan summary does not contains any summaries")
// 	}

// 	f, err := os.Create(p)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	t, err := template.New("template").Parse(htmlPattern)
// 	if err != nil {
// 		return err
// 	}
// 	err = t.Execute(f, s)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
