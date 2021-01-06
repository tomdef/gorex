package cmd

import (
	"bufio"
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

	for i, line := range scope.ScopeSummary.Content {
		logger.Trace().Msgf("\tProcess line [%06d][%v]", i, line)
		for j, r := range rx {
			isMatch := checkScopeMatch(line, r, true)
			logger.Trace().Msgf("\t\tProcess match [%02d][%v][match=%v]", j, r, isMatch)

			if isMatch == true {
				matchLines = append(matchLines, common.MatchLine{
					Index: i + scope.ScopeSummary.Started,
					Line:  line,
				})
			}
		}
	}

	if (scope.ScopeConfig.SearchQueryMode == common.SearchQueryOperatorAny && len(matchLines) > 0) || (len(matchLines) >= requiredMatchCount) {
		result.Matches = matchLines
		logger.Trace().Msgf("\t\tMATCHES FOUND IN SCOPE [%06d..%06d]:[%v]", result.Started, result.Finished, len(matchLines))
	}

	return result, nil
}

func scan(input string, outputhtml string, outputjson string) error {
	log.Printf("*** Start scan. Commands file path : %v", input)

	cfg, err := common.ReadScopeConfiguration(input)
	if err != nil {
		log.Printf("Read config error: %v\n", err)
		return err
	}

	if err = cfg.IsValid(); err != nil {
		log.Printf("Config is not valid: %v\n", err)
		return err
	}

	var folder string = cfg.Folder
	var filter string = cfg.Filter

	abs, err := filepath.Abs(folder)
	if err == nil {
		folder = abs
		log.Printf("\tFolder resolve to: %v", folder)
	} else {
		log.Printf("[!] Folder path resolve error: %v", err)
	}

	var scanSummary common.ScanSummary = common.ScanSummary{
		Folder:       folder,
		Filter:       filter,
		CreationTime: time.Now(),
		Summary:      nil,
	}

	// -----------------------------------------------------------------------------
	// consume scope summary
	// -----------------------------------------------------------------------------
	// go func(channel *channelSummary, w *sync.WaitGroup) {
	// 	for {
	// 		s := <-(*channel)

	// 		logger.Trace().Msg("Add scope to summary...")
	// 		found := false
	// 		for _, value := range scanSummary.Summary {
	// 			if value.FileName == s.FileName {
	// 				logger.Trace().Msg("Found and modify FileScopeSummary")
	// 				(&value).Scopes = append((&value).Scopes, s)
	// 				(&value).AllMatches = len((&value).Scopes)
	// 				found = true
	// 				break
	// 			}
	// 		}

	// 		if found == false {
	// 			logger.Trace().Msg("Not found FileScopeSummary, add new")
	// 			fileScopeSummary := common.FileScopeSummary{
	// 				FileName:   s.FileName,
	// 				Scopes:     []common.ScopeSummary{s},
	// 				AllMatches: 1,
	// 			}

	// 			scanSummary.Summary = append(scanSummary.Summary, fileScopeSummary)
	// 		}
	// 		w.Done()
	// 	}
	// }(&cSummary, &wgSummary)

	// -----------------------------------------------------------------------------
	// read scopes and find matches
	// -----------------------------------------------------------------------------
	// go func(channel *channelScope, w *sync.WaitGroup) {
	// 	for {
	// 		scope := <-(*channel)

	// 		var rx []*regexp.Regexp

	// 		for _, v := range scope.ScopeConfig.SearchQuery {
	// 			r, err := regexp.Compile(v)
	// 			if err == nil {
	// 				rx = append(rx, r)
	// 			}
	// 		}

	// 		logger.Trace().Msgf("Process scope [name=%v] in [%v][%06d..%06d]",
	// 			scope.ScopeSummary.Name, scope.ScopeSummary.FileName, scope.ScopeSummary.Started, scope.ScopeSummary.Finished)

	// 		requiredMatchCount := len(rx)
	// 		var matchLines []common.MatchLine

	// 		for i, line := range scope.ScopeSummary.Content {
	// 			logger.Trace().Msgf("\tProcess line [%06d][%v]", i, line)
	// 			for j, r := range rx {
	// 				isMatch := checkScopeMatch(line, r, true)
	// 				logger.Trace().Msgf("\t\tProcess match [%02d][%v][match=%v]", j, r, isMatch)

	// 				if isMatch == true {
	// 					matchLines = append(matchLines, common.MatchLine{
	// 						Index: i + scope.ScopeSummary.Started,
	// 						Line:  line,
	// 					})
	// 				}
	// 			}
	// 		}

	// 		if (scope.ScopeConfig.SearchQueryMode == common.SearchQueryOperatorAny && len(matchLines) > 0) || (len(matchLines) >= requiredMatchCount) {
	// 			logger.Trace().Msgf("\t\tRequired match(es) found in scope [%06d..%06d]",
	// 				scope.ScopeSummary.Started, scope.ScopeSummary.Finished)
	// 			scope.ScopeSummary.Matches = matchLines
	// 			wgSummary.Add(1)
	// 			cSummary <- scope.ScopeSummary
	// 		}
	// 		wgSummary.Wait()
	// 		w.Done()
	// 	}
	// }(&cScope, &wgScope)

	// -----------------------------------------------------------------------------
	// read files and find scope(s)
	// -----------------------------------------------------------------------------
	go func(channel *channelFile, wgFile *sync.WaitGroup, wgScope *sync.WaitGroup, sc common.ScanConfig) {
		for {
			path := <-(*channel)

			log.Printf("\t\t-> Process file [%v]\n", path)
			//var scopes []common.ScopeSummary

			fileScopeSummary := common.FileScopeSummary{
				FileName:   path,
				Scopes:     []common.ScopeSummary{},
				AllMatches: 0,
			}

			for _, s := range sc.Scopes {

				rxStart, err := regexp.Compile(s.StartQuery)
				if err != nil {
					log.Printf("[!] regex start (%v) is empty or invalid.", s.StartQuery)
				}

				rxStop, err := regexp.Compile(s.FinishQuery)
				if err != nil {
					log.Printf("[!] regex finish (%v) is empty or invalid.", s.FinishQuery)
				}

				// read file line by line -->
				file, err := os.Open(path)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				index := 1
				scopeIsOpen := false
				var scopeSummary common.ScopeSummary

				for scanner.Scan() {

					line := scanner.Text()

					if s.IsOneLineSearch() == true {
						log.Print("[!] One line search is not supported yet\n")
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
									logger.Info().Msgf("FIND Matches:%v", s)
									scopeSummary.Matches = append(scopeSummary.Matches, s.Matches...)

									if len(scopeSummary.Matches) > 0 {
										logger.Info().Msg("UPDATE FILE SUMMARY")
										fileScopeSummary.Scopes = append(fileScopeSummary.Scopes, scopeSummary)
										fileScopeSummary.AllMatches = len(fileScopeSummary.Scopes)
									}
								}
							} else {
								scopeSummary.Content = append(scopeSummary.Content, line)
								tmp := fmt.Sprintf(formatContentHTML, index, notMatchedMark, line)
								scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
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

			if fileScopeSummary.Scopes != nil {
				logger.Info().Msgf("ADD FILE [%v] MATCHES TO SUMMARY", fileScopeSummary.FileName)
				mutex.Lock()
				scanSummary.Summary = append(scanSummary.Summary, fileScopeSummary)
				mutex.Unlock()
			}

			wgFile.Done()
		}

	}(&cFile, &wgFile, &wgScope, cfg)

	// -----------------------------------------------------------------------------

	log.Printf("\tScan folder [%v]...\n", folder)

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
				log.Printf("\tFilter match error:%v", merr)
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

	logger.Info().Msg("SAVE")
	if outputhtml != "" {
		scanSummary.LogToHTML(outputhtml)
	}
	if outputjson != "" {
		scanSummary.LogToFile(outputjson)
	}

	log.Println("*** End ***")

	return err
}

func init() {

	scanCmd.Flags().StringVarP(&input, "input", "i", ".", "Input file path (*.json) with scan commands.")
	scanCmd.Flags().StringVarP(&outpouHTML, fOutputHTML, "o", "", "Output html report.")
	scanCmd.Flags().StringVarP(&outputJSON, fOutputJSON, "d", "", "Output raw data in json format.")

	rootCmd.AddCommand(scanCmd)
}
