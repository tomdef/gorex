package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	common "gorex/pkg/common"
	utils "gorex/pkg/utils"

	"github.com/spf13/cobra"
)

const (
	notMatchedMark    = "-"
	startScopeMark    = ">"
	finishScopeMark   = "<"
	matchedMark       = "*"
	formatContentHTML = "[%05d|%v][%v]"
	fInput            = "input"
	fOutputHTML       = "outputhtml"
	fOutputJSON       = "outputdata"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "A scan folder with advanced regex configurations",

		RunE: func(cmd *cobra.Command, args []string) error {

			if err := scan(input, outpouHTML, outputJSON); err != nil {
				return err
			}
			return nil
		},
	}
	wgFile    sync.WaitGroup
	wgScope   sync.WaitGroup
	wgSummary sync.WaitGroup
	cFile     = make(channelFile)
	cScope    = make(channelScope)
	cSummary  = make(channelSummary)

	mutex = &sync.Mutex{}

	// Commands represents path to command file
	input      string
	outpouHTML string
	outputJSON string

	logger = utils.CreateLogger("scan")
)

type channelFile chan (string)
type channelScope chan (common.ScopeSummaryWithConfig)
type channelSummary chan (common.ScopeSummary)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func checkIfBeginScope(line string, rx *regexp.Regexp, scopeIsOpen bool) bool {
	return (scopeIsOpen == false) && (rx.MatchString(line) == true)
}

func checkIfEndScope(line string, rx *regexp.Regexp, scopeIsOpen bool) bool {
	return (scopeIsOpen == true) && (rx.MatchString(line) == true)
}

func checkScopeMatch(line string, rx *regexp.Regexp, scopeIsOpen bool) bool {
	return (scopeIsOpen == true) && (rx.MatchString(line) == true)
}

func findMatchesInScope(scope common.ScopeSummaryWithConfig) (common.ScopeSummary, error) {

	var rx []*regexp.Regexp
	var result common.ScopeSummary = scope.ScopeSummary

	for _, v := range scope.ScopeConfig.SearchQuery {
		r, err := regexp.Compile(v)
		if err != nil {
			return scope.ScopeSummary, err
		}

		rx = append(rx, r)
	}

	logger.Info().Msgf("PROCESS SCOPE [name=%v] in [%v][%06d..%06d]",
		scope.ScopeSummary.Name, scope.ScopeSummary.FileName, scope.ScopeSummary.Started, scope.ScopeSummary.Finished)

	requiredMatchCount := len(rx)
	var matchLines []common.MatchLine
	matchesOfRxCounter := make([]int, requiredMatchCount)

	for i, line := range scope.ScopeSummary.Content {
		for j, r := range rx {
			isMatch := checkScopeMatch(line, r, true)

			if isMatch == true {
				matchesOfRxCounter[j] = matchesOfRxCounter[j] + 1
				matchLines = append(matchLines, common.MatchLine{
					Index: i + scope.ScopeSummary.Started,
					Line:  line,
				})
			}
		}
	}

	foundMatchesOfRx := 0
	for _, k := range matchesOfRxCounter {
		if k > 0 {
			foundMatchesOfRx++
		}
	}

	if (scope.ScopeConfig.SearchQueryMode == common.SearchQueryOperatorAny && len(matchLines) > 0) || (foundMatchesOfRx >= requiredMatchCount) {
		result.Matches = matchLines
		logger.Trace().Msgf("\t\tMATCHES FOUND IN SCOPE [%06d..%06d], lines [%v] of [%v] query", result.Started, result.Finished, len(matchLines), foundMatchesOfRx)
	}

	return result, nil
}

func scan(input string, outputhtml string, outputjson string) error {
	logger.Info().Msgf("START SCAN. Command(s) file path : %v", input)

	cfg, err := common.ReadScopeConfiguration(input)
	if err != nil {
		logger.Err(err)
		return err
	}

	if err = cfg.IsValid(); err != nil {
		logger.Err(err)
		return err
	}

	var folder string = cfg.Folder
	var filter string = cfg.Filter

	abs, err := filepath.Abs(folder)
	if err == nil {
		folder = abs
		logger.Trace().Msgf("Folder resolved to: %v", folder)
	} else {
		logger.Err(err)
	}

	var scanSummary common.ScanSummary = common.ScanSummary{
		Folder:       folder,
		Filter:       filter,
		CreationTime: time.Now(),
		Summary:      nil,
		ScanFiles:    0,
	}

	// -----------------------------------------------------------------------------
	// read files and find scope(s)
	// -----------------------------------------------------------------------------
	go func(channel *channelFile, wgFile *sync.WaitGroup, wgScope *sync.WaitGroup, sc common.ScanConfig) {
		for {
			path := <-(*channel)

			logger.Info().Msgf("\t-> Process file [%v]", path)

			fileScopeSummary := common.FileScopeSummary{
				FileName:   path,
				Scopes:     []common.ScopeSummary{},
				AllMatches: 0,
			}

			for _, s := range sc.Scopes {

				rxStart, err := regexp.Compile(s.StartQuery)
				if err != nil {
					logger.Err(err).Send()
				}

				rxStop, err := regexp.Compile(s.FinishQuery)
				if err != nil {
					logger.Err(err).Send()
				}

				// read file line by line -->
				file, err := os.Open(path)
				if err != nil {
					logger.Err(err).Send()
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				index := 1
				scopeIsOpen := false
				var scopeSummary common.ScopeSummary

				for scanner.Scan() {

					line := scanner.Text()

					if s.IsOneLineSearch() == true {
						logger.Err(errors.New("One line search is not supported yet")).Send()
					} else {
						if checkIfBeginScope(line, rxStart, scopeIsOpen) == true {
							logger.Info().Msgf("BEGIN SCOPE [%v] in line [%v]", s.Name, index)
							scopeIsOpen = true
							scopeSummary = common.ScopeSummary{
								Name:     s.Name,
								FileName: path,
								Started:  index,
								Finished: 0,
								Matches:  nil,
								Content:  nil,
							}
							scopeSummary.Content = append(scopeSummary.Content, line)
							tmp := fmt.Sprintf(formatContentHTML, index, startScopeMark, line)
							scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
						} else {
							if checkIfEndScope(line, rxStop, scopeIsOpen) == true {
								logger.Info().Msgf("END SCOPE [%v] in line [%v]", s.Name, index)
								scopeIsOpen = false

								scopeSummary.Finished = index
								scopeSummary.Content = append(scopeSummary.Content, line)
								tmp := fmt.Sprintf(formatContentHTML, index, finishScopeMark, line)
								scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))

								scopeSummaryWithConfig := common.ScopeSummaryWithConfig{
									ScopeSummary: scopeSummary,
									ScopeConfig:  s,
								}

								s, err := findMatchesInScope(scopeSummaryWithConfig)
								if err == nil {
									mutex.Lock()

									scopeSummary.Matches = append(scopeSummary.Matches, s.Matches...)

									if len(scopeSummary.Matches) > 0 {
										logger.Info().Msg("UPDATE FILE SUMMARY")
										fileScopeSummary.Scopes = append(fileScopeSummary.Scopes, scopeSummary)
										fileScopeSummary.AllMatches = len(fileScopeSummary.Scopes)
									}
									mutex.Unlock()
								}
							} else {
								if scopeIsOpen == true {
									scopeSummary.Content = append(scopeSummary.Content, line)
									tmp := fmt.Sprintf(formatContentHTML, index, notMatchedMark, line)
									scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
								}
							}
						}
					}
					index++
				}

				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
				// <--
			}

			mutex.Lock()
			scanSummary.ScanFiles++
			if (fileScopeSummary.Scopes != nil) && (len(fileScopeSummary.Scopes) > 0) {
				logger.Info().Msgf("ADD FILE [%v] MATCHES TO SUMMARY", fileScopeSummary.FileName)
				scanSummary.Summary = append(scanSummary.Summary, fileScopeSummary)
			}
			mutex.Unlock()

			wgFile.Done()
		}

	}(&cFile, &wgFile, &wgScope, cfg)

	// -----------------------------------------------------------------------------

	logger.Info().Msgf("SCAN FOLDER [%v]...", folder)

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() == true {
			return nil
		}

		if info.IsDir() == false {

			matched, merr := filepath.Match(filter, filepath.Base(path))
			if merr != nil {
				logger.Trace().Msgf("Filter match warning:%v", merr)
			} else {

				if matched == true {
					wgFile.Add(1)
					cFile <- path
				}
			}
		}

		return nil
	})

	wgFile.Wait()

	if (outputhtml != "") || (outputjson != "") {
		logger.Info().Msg("SAVE...")
		if outputhtml != "" {
			logger.Info().Msgf("\tSave to [%v]", outputhtml)
			scanSummary.LogToHTML(outputhtml)
		}
		if outputjson != "" {
			logger.Info().Msgf("\tSave to [%v]", outputjson)
			scanSummary.LogToFile(outputjson)
		}
	} else {
		logger.Info().Msg("SAVE skipped")
	}

	logger.Info().Msg("*** END ***")

	return err
}

func init() {

	scanCmd.Flags().StringVarP(&input, "input", "i", ".", "Input file path (*.json) with scan commands.")
	scanCmd.Flags().StringVarP(&outpouHTML, fOutputHTML, "o", "", "Output html report.")
	scanCmd.Flags().StringVarP(&outputJSON, fOutputJSON, "d", "", "Output raw data in json format.")

	rootCmd.AddCommand(scanCmd)
}
