package cmd

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	common "gorex/pkg/common"
	"gorex/pkg/utils"

	"github.com/dlclark/regexp2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	notMatchedMark    = " "
	startScopeMark    = ">"
	finishScopeMark   = "<"
	matchedMark       = "*"
	formatContentHTML = "[%05d|%v][%v]"
	fInput            = "input"
	fOutputHTML       = "outputhtml"
	fOutputJSON       = "outputdata"
	fTrace            = "trace"
	fShow             = "show"
	eofLine           = "[EOF]"
	regexOpt          = regexp2.Singleline
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "A scan folder with advanced regex configurations",

		RunE: func(cmd *cobra.Command, args []string) error {

			if err := scan(input, outputHTML, outputJSON, trace); err != nil {
				return err
			}
			return nil
		},
	}

	wgFile sync.WaitGroup
	cFile  = make(channelFile)
	mutex  = &sync.Mutex{}

	// Commands represents path to command file
	input      string
	outputHTML string
	outputJSON string
	trace      bool
	show       bool
)

type channelFile chan (string)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func checkIfBeginScope(line string, rx *regexp2.Regexp, scopeIsOpen bool) bool {
	m, e := rx.MatchString(line)
	return (scopeIsOpen == false) && (m == true) && (e == nil)
}

func checkIfEndScope(line string, rx *regexp2.Regexp, scopeIsOpen bool) bool {
	m, e := rx.MatchString(line)
	return (scopeIsOpen == true) && (m == true) && (e == nil)
}

func checkScopeMatch(line string, rx *regexp2.Regexp, scopeIsOpen bool) bool {
	m, e := rx.MatchString(line)
	return (scopeIsOpen == true) && (m == true) && (e == nil)
}

func findMatchesInScope(scope common.ScopeSummaryWithConfig, logger zerolog.Logger) (common.ScopeSummary, error) {

	var rx []*regexp2.Regexp
	var result common.ScopeSummary = scope.ScopeSummary

	for _, v := range scope.ScopeConfig.SearchQuery {
		r := regexp2.MustCompile(v, regexOpt)
		// if err != nil {
		// 	return scope.ScopeSummary, err
		// }

		rx = append(rx, r)
	}

	logger.Trace().Msgf("Process scope [name=%v] in [%v][%06d..%06d]",
		scope.ScopeSummary.Name, scope.ScopeSummary.FileName, scope.ScopeSummary.Started, scope.ScopeSummary.Finished)

	requiredMatchCount := len(rx)
	var matchLines []common.MatchLine
	matchesOfRxCounter := make([]int, requiredMatchCount)

	for i, line := range scope.ScopeSummary.Content {
		for j, r := range rx {

			isMatch := checkScopeMatch(line, r, true)

			if isMatch == true {
				findAndMarkAsMatches(&logger, &scope.ScopeSummary.ContentAsHTML, line)
				if (scope.ScopeConfig.SearchQueryMode == common.SearchQueryOperatorAny) || (scope.ScopeConfig.SearchQueryMode == common.SearchQueryOperatorAll) || (j == 0) || (matchesOfRxCounter[j-1] > 0) {
					matchesOfRxCounter[j] = matchesOfRxCounter[j] + 1
					matchLines = append(matchLines, common.MatchLine{
						Index: i + scope.ScopeSummary.Started,
						Line:  line,
					})
				}
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

func findAndMarkAsMatches(logger *zerolog.Logger, l *[]string, x string) {

	fx := "[" + x + "]"
	logger.Trace().Msgf("*** Start search [%v]", x)

	for i := 0; i < len(*l); i++ {
		v := (*l)[i]

		if strings.HasSuffix(v, fx) == true {
			logger.Trace().Msgf("*** Found and modify line %v", v)
			v = strings.Replace(v, "| ][", "|*][", 1)
			(*l)[i] = v
		}
	}
}

func beginScope(logger *zerolog.Logger, fileName string, line string, index int, scopeName string, scopeIsOpen *bool, scopeSummary *common.ScopeSummary) {

	logger.Trace().Msgf("Begin scope [%v] in line [%v]", scopeName, index)
	*scopeIsOpen = true
	*scopeSummary = common.ScopeSummary{
		Name:     scopeName,
		FileName: fileName,
		Started:  index,
		Finished: 0,
		Matches:  nil,
		Content:  nil,
	}
	scopeSummary.Content = append(scopeSummary.Content, line)
	tmp := fmt.Sprintf(formatContentHTML, index, startScopeMark, line)
	scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
}

func endScope(logger *zerolog.Logger, scan bool, line string, index int, scopeName string, scopeIsOpen *bool, scopeSummary *common.ScopeSummary, scopeConfig *common.ScopeConfig, fileScopeSummary *common.FileScopeSummary) {

	logger.Trace().Msgf("End scope [%v] in line [%v]", scopeName, index)

	*scopeIsOpen = false
	(*scopeSummary).Finished = index

	if scan == false {
		logger.Trace().Msg("End of file")

		scopeSummary.Content = append(scopeSummary.Content, eofLine)
		tmp := html.EscapeString(eofLine)
		scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, tmp)

	} else {
		if line != "" {
			scopeSummary.Content = append(scopeSummary.Content, line)
			tmp := fmt.Sprintf(formatContentHTML, index, finishScopeMark, line)
			scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
		}
	}

	scopeSummaryWithConfig := common.ScopeSummaryWithConfig{
		ScopeSummary: *scopeSummary,
		ScopeConfig:  *scopeConfig,
	}

	s, err := findMatchesInScope(scopeSummaryWithConfig, *logger)
	if err == nil {
		mutex.Lock()

		scopeSummary.Matches = append(scopeSummary.Matches, s.Matches...)

		if len(scopeSummary.Matches) > 0 {
			logger.Trace().Msg("Update summary")
			fileScopeSummary.Scopes = append(fileScopeSummary.Scopes, *scopeSummary)
			fileScopeSummary.AllMatches = len(fileScopeSummary.Scopes)
		}
		mutex.Unlock()
	}
}

func scan(input string, outputhtml string, outputjson string, trace bool) error {

	logger := utils.CreateLogger("scan", trace)

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
	go func(channel *channelFile, wgFile *sync.WaitGroup, sc common.ScanConfig) {
		for {
			path := <-(*channel)

			logger.Info().Msgf("\t-> Process file [%v]", path)

			fileScopeSummary := common.FileScopeSummary{
				FileName:   path,
				Scopes:     []common.ScopeSummary{},
				AllMatches: 0,
			}

			for _, s := range sc.Scopes {

				rxStart := regexp2.MustCompile(s.StartQuery, regexOpt)
				// if err != nil {
				// 	logger.Err(err).Send()
				// }

				rxStop := regexp2.MustCompile(s.FinishQuery, regexOpt)
				// if err != nil {
				// 	logger.Err(err).Send()
				// }

				// read file line by line -->
				file, err := os.Open(path)
				if err != nil {
					logger.Err(err).Send()
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				index := 0
				scopeIsOpen := false
				var scopeSummary common.ScopeSummary
				var scan bool = true
				var line string
				for scan {
					scan = scanner.Scan()

					if scan == true {
						line = scanner.Text()
						index++
					}

					if checkIfBeginScope(line, rxStart, scopeIsOpen) {

						beginScope(&logger, path, line, index, s.Name, &scopeIsOpen, &scopeSummary)

					} else if (checkIfBeginScope(line, rxStart, false)) && (s.StartQueryCloseScope == true) && (scopeIsOpen == true) {

						logger.Info().Msg("End scope because of StartQueryCloseScope flag.")

						endScope(&logger, scan, line, index, s.Name, &scopeIsOpen, &scopeSummary, &s, &fileScopeSummary)
						beginScope(&logger, path, line, index, s.Name, &scopeIsOpen, &scopeSummary)

					} else {
						if (checkIfEndScope(line, rxStop, scopeIsOpen) == true) || ((scopeIsOpen == true) && (scan == false)) {

							endScope(&logger, scan, line, index, s.Name, &scopeIsOpen, &scopeSummary, &s, &fileScopeSummary)

						} else {
							if scopeIsOpen == true {

								scopeSummary.Content = append(scopeSummary.Content, html.EscapeString(line))

								logger.Info().Msgf("|%v|", line)

								tmp := fmt.Sprintf(formatContentHTML, index, notMatchedMark, line)
								scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))

							}
						}
					}

					//index++
				}

				if err := scanner.Err(); err != nil {
					logger.Fatal().Err(err)
				}
				// <--
			}

			mutex.Lock()
			scanSummary.ScanFiles++
			if (fileScopeSummary.Scopes != nil) && (len(fileScopeSummary.Scopes) > 0) {
				logger.Trace().Msgf("ADD FILE [%v] MATCHES TO SUMMARY", fileScopeSummary.FileName)
				scanSummary.Summary = append(scanSummary.Summary, fileScopeSummary)
			}
			mutex.Unlock()

			wgFile.Done()
		}

	}(&cFile, &wgFile, cfg)

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

			info, err := os.Stat(outputhtml)
			if !os.IsNotExist(err) {
				newName := outputhtml + ".backup"
				logger.Info().Msgf("\tRename previous file [%v] to [%v]", info.Name(), newName)
				os.Rename(outputhtml, newName)
			}

			logger.Info().Msgf("\tSave to [%v]", outputhtml)
			e := scanSummary.LogToHTML(outputhtml)
			if e != nil {
				logger.Err(err)
			}
		}
		if outputjson != "" {
			logger.Info().Msgf("\tSave to [%v]", outputjson)
			e := scanSummary.LogToFile(outputjson)
			if e != nil {
				logger.Err(err)
			}
		}

		if (outputjson != "") || (outputhtml != "") && (show == true) {

			logger.Info().Msg("SHOW result...")
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", outputhtml).Start()
			if err != nil {
				logger.Error().Msg(err.Error())
			}
		}
	} else {
		logger.Info().Msg("SAVE skipped")
	}

	logger.Info().Msg("*** END ***")

	return err
}

func init() {

	scanCmd.Flags().StringVarP(&input, "input", "i", ".", "Input file path (*.json) with scan commands.")
	scanCmd.Flags().StringVarP(&outputHTML, fOutputHTML, "o", "", "Output html report.")
	scanCmd.Flags().StringVarP(&outputJSON, fOutputJSON, "d", "", "Output raw data in json format.")
	scanCmd.Flags().BoolVarP(&trace, fTrace, "t", false, "Set trace mode.")
	scanCmd.Flags().BoolVarP(&show, fShow, "s", false, "Show result after scan.")

	rootCmd.AddCommand(scanCmd)
}
