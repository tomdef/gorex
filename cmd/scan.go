package cmd

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gorex/pkg/common"
	"gorex/pkg/utils"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type channelFile chan (string)

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

	wgFile sync.WaitGroup
	cFile  = make(channelFile)
	mutex  = &sync.Mutex{}
)

const (
	fInput      = "input"
	fOutputHTML = "outputhtml"
	fOutputJSON = "outputjson"
	fShow       = "show"
	fTrace      = "trace"
	regexOpt    = regexp2.Singleline
	eofLine     = "[EOF]"

	notMatchedMark    = " "
	startScopeMark    = ">"
	finishScopeMark   = "<"
	matchedMark       = "*"
	formatContentHTML = "[%05d|%v][%v]"
)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func init() {

	scanCmd.Flags().StringVarP(&input, "input", "i", ".", "Input file path (*.json) with scan commands.")
	scanCmd.Flags().StringVarP(&outputHTML, fOutputHTML, "o", "", "Output html report.")
	scanCmd.Flags().StringVarP(&outputJSON, fOutputJSON, "j", "", "Output json report.")
	scanCmd.Flags().BoolVarP(&trace, fTrace, "t", false, "Set trace mode.")
	scanCmd.Flags().BoolVarP(&show, fShow, "s", false, "Show result after scan.")

	rootCmd.AddCommand(scanCmd)
}

func scan(input string, outputhtml string, outputjson string, trace bool, show bool) error {

	durationStart := time.Now()
	logger := utils.CreateLogger("scan", trace)

	logger.Info().Msgf("READ SCAN CONFIGURATION. Command(s) file path : %v", input)

	scanConfig, err := common.ReadScanConfigurationFromFile(input)
	if err != nil {
		logger.Err(err)
		return err
	}

	var folder string = scanConfig.Folder
	var filter string = scanConfig.Filter
	var scanSummary common.ScanSummary = common.ScanSummary{
		Folder:       folder,
		Filter:       filter,
		Scopes:       scanConfig.Scopes,
		CreationTime: time.Now(),
		DurationTime: 0,
		Summary:      nil,
		ScanFiles:    0,
	}

	abs, err := filepath.Abs(folder)
	if err == nil {
		folder = abs
	} else {
		logger.Err(err)
	}

	logger.Info().Msgf("START SCAN. Folder [%v] with filter [%v]", folder, filter)

	// -----------------------------------------------------------------------------
	// read files and find scope(s)
	// -----------------------------------------------------------------------------
	go processFile(&logger, &cFile, &wgFile, scanConfig, &scanSummary)

	// -->
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !info.IsDir() {

			matched, merr := filepath.Match(filter, filepath.Base(path))
			if merr != nil {
				logger.Trace().Msgf("Filter match warning:%v", merr)
			} else {

				if matched {
					wgFile.Add(1)
					cFile <- path
				}
			}
		}

		return nil
	})

	wgFile.Wait()
	elapsed := time.Since(durationStart)
	scanSummary.DurationTime = elapsed

	// save

	if (outputhtml != "") || (outputjson != "") {
		logger.Info().Msg("SAVE...")

		if outputhtml != "" {

			var b []byte
			var e error
			if b, e = scanSummary.AsHTML(); e != nil {
				logger.Error().Msgf("Cannot save summary as HTML. %v", e.Error())
			} else {
				info, err := os.Stat(outputhtml)
				if !os.IsNotExist(err) {
					newName := outputhtml + ".backup"
					if err = os.Remove(newName); err != nil {
						logger.Info().Msgf("\tRemove old file [%v] ends with error : [%v]", newName, err.Error())
					}
					logger.Info().Msgf("\tRename previous file [%v] to [%v]", info.Name(), newName)
					if err = os.Rename(outputhtml, newName); err != nil {
						logger.Info().Msgf("\tRename file [%v] ends with error : [%v]", info.Name(), err.Error())
					}
				}

				logger.Info().Msgf("\tSave to HTML format file [%v]", outputhtml)
				os.WriteFile(outputhtml, b, os.ModeExclusive)
			}
		}

		if outputjson != "" {

			var b []byte
			var e error
			if b, e = scanSummary.AsJSON(); e != nil {
				logger.Error().Msgf("Cannot save summary as JSON. %v", e.Error())
			} else {
				info, err := os.Stat(outputjson)
				if !os.IsNotExist(err) {
					newName := outputjson + ".backup"
					if err = os.Remove(newName); err != nil {
						logger.Info().Msgf("\tRemove old file [%v] ends with error : [%v]", newName, err.Error())
					}
					logger.Info().Msgf("\tRename previous file [%v] to [%v]", info.Name(), newName)
					if err = os.Rename(outputjson, newName); err != nil {
						logger.Info().Msgf("\tRename file [%v] ends with error : [%v]", info.Name(), err.Error())
					}
				}

				logger.Info().Msgf("\tSave to JSON format file [%v]", outputjson)
				os.WriteFile(outputjson, b, os.ModeExclusive)
			}
		}
	} else {
		logger.Info().Msg("SAVE skipped")
	}

	// <--
	logger.Info().Msgf("FINISH SCAN after %s", elapsed)

	return err
}

// -----------------------------------------------------------------------------
// read files and find scope(s)
// -----------------------------------------------------------------------------
func processFile(logger *zerolog.Logger, channel *channelFile, wgFile *sync.WaitGroup, sc common.ScanConfig, summary *common.ScanSummary) {
	for {
		path := <-(*channel)
		pathHash := fileNameHash(path)
		logger.Info().Msgf("Begin process file [%v], hash=[%v]", path, pathHash)

		// -->

		fileScopeSummary := common.FileScopeSummary{
			FileName:   path,
			Scopes:     []common.ScopeSummary{},
			AllMatches: 0,
		}
		for _, s := range sc.Scopes {

			rxStart := regexp2.MustCompile(s.Begin, regexOpt)
			rxStop := regexp2.MustCompile(s.End, regexOpt)

			// read file line by line -->
			file, err := os.Open(path)
			if err != nil {
				logger.Err(err).Send()
			}
			//defer file.Close()

			scanner := bufio.NewScanner(file)
			index := 0
			scopeIsOpen := false
			var scopeSummary common.ScopeSummary
			var scan bool = true
			var line string
			for scan {
				scan = scanner.Scan()

				if scan {
					line = scanner.Text()
					index++
				}

				if checkIfBeginScope(line, rxStart, scopeIsOpen) {

					beginScope(logger, path, line, index, s.Name, &scopeIsOpen, &scopeSummary)

				} else if (checkIfBeginScope(line, rxStart, false)) && s.AutoClose && scopeIsOpen {

					logger.Trace().Msgf("\t[%v] End scope because of StartQueryCloseScope flag.", pathHash)

					endScope(logger, scan, line, index, s.Name, &scopeIsOpen, &scopeSummary, &s, &fileScopeSummary)
					beginScope(logger, path, line, index, s.Name, &scopeIsOpen, &scopeSummary)

				} else {
					if checkIfEndScope(line, rxStop, scopeIsOpen) || (scopeIsOpen && !scan) {

						endScope(logger, scan, line, index, s.Name, &scopeIsOpen, &scopeSummary, &s, &fileScopeSummary)

					} else {
						if scopeIsOpen {

							scopeSummary.Content = append(scopeSummary.Content, html.EscapeString(line))
							tmp := fmt.Sprintf(formatContentHTML, index, notMatchedMark, line)
							scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
						}
					}
				}
			}

			if err := scanner.Err(); err != nil {
				logger.Fatal().Err(err)
			}
		}

		mutex.Lock()
		summary.ScanFiles++
		if (fileScopeSummary.Scopes != nil) && (len(fileScopeSummary.Scopes) > 0) {
			logger.Trace().Msgf("ADD FILE [%v] MATCHES TO SUMMARY", fileScopeSummary.FileName)
			summary.Summary = append(summary.Summary, fileScopeSummary)
		}
		mutex.Unlock()

		// <--

		logger.Info().Msgf("End process file [%v] hash=[%v]", path, pathHash)

		wgFile.Done()
	}
}

func checkIfBeginScope(line string, rx *regexp2.Regexp, scopeIsOpen bool) bool {
	m, e := rx.MatchString(line)
	return !scopeIsOpen && m && (e == nil)
}

func checkIfEndScope(line string, rx *regexp2.Regexp, scopeIsOpen bool) bool {
	m, e := rx.MatchString(line)
	return scopeIsOpen && m && (e == nil)
}

func beginScope(logger *zerolog.Logger, fileName string, line string, index int, scopeName string, scopeIsOpen *bool, scopeSummary *common.ScopeSummary) {

	pathHash := fileNameHash(fileName)
	logger.Trace().Msgf("\t[%v] Begin scope [%v] in line [%v]", pathHash, scopeName, index)
	*scopeIsOpen = true
	*scopeSummary = common.ScopeSummary{
		Name:     scopeName,
		FileName: fileName,
		Started:  index,
		Finished: 0,
		Matches:  nil,
		Content:  nil,
	}
	scopeSummary.ResolveId()
	scopeSummary.Content = append(scopeSummary.Content, line)
	tmp := fmt.Sprintf(formatContentHTML, index, startScopeMark, line)
	scopeSummary.ContentAsHTML = append(scopeSummary.ContentAsHTML, html.EscapeString(tmp))
}

func endScope(logger *zerolog.Logger, scan bool, line string, index int, scopeName string, scopeIsOpen *bool, scopeSummary *common.ScopeSummary, scopeConfig *common.ScopeConfig, fileScopeSummary *common.FileScopeSummary) {

	pathHash := fileNameHash(fileScopeSummary.FileName)
	*scopeIsOpen = false
	(*scopeSummary).Finished = index

	var boeof string = ""

	if !scan {
		boeof = " because of EOF"
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

	logger.Trace().Msgf("\t[%v] End scope [%v] in line [%v]%v", pathHash, scopeName, index, boeof)

	s := findMatchesInScope(*logger, scopeConfig, scopeSummary)

	if len(s) > 0 {
		mutex.Lock()
		logger.Trace().Msgf("\t[%v] update summary. Add [%v] match(es) from [%v].", pathHash, len(s), scopeConfig.Name)

		scopeSummary.Matches = append(scopeSummary.Matches, s...)

		for _, m := range scopeSummary.Matches {
			logger.Trace().Msgf("\t\t[%v] match [%v][%v] -> [%v]", pathHash, m.Index, m.Line, strings.Join(m.MatchNames, ","))
		}

		if len(scopeSummary.Matches) > 0 {
			fileScopeSummary.Scopes = append(fileScopeSummary.Scopes, *scopeSummary)
			fileScopeSummary.AllMatches = len(fileScopeSummary.Scopes)
		}
		mutex.Unlock()
	} else {
		logger.Trace().Msgf("\t[%v] does not contains any match(es) from [%v].", pathHash, scopeConfig.Name)
	}
}

func findMatchesInScope(logger zerolog.Logger, scopeConfig *common.ScopeConfig, scopeSummary *common.ScopeSummary) []common.MatchLine {

	pathHash := fileNameHash(scopeSummary.FileName)
	logger.Trace().Msgf("\t[%v] Process scope [name=%v] in [%v][%06d..%06d]",
		pathHash, scopeSummary.Name, scopeSummary.FileName, scopeSummary.Started, scopeSummary.Finished)

	var matchLines []common.MatchLine
	var result []common.MatchLine

	for i, line := range scopeSummary.Content {
		matchLines = append(matchLines, common.MatchLine{Index: i + scopeSummary.Started, Line: line, MatchNames: []string{}})
	}

	for i := 0; i < len(matchLines); i++ {

		for _, m := range scopeConfig.Matches {
			matchesIndex := m.IsMatch(matchLines)
			if matchesIndex == nil {
				return []common.MatchLine{}
			} else {
				if len(matchesIndex) > 0 {
					for _, mi := range matchesIndex {
						if matchLines[i].Index == mi {
							matchLines[i].MatchNames = append(matchLines[i].MatchNames, m.Name)
						}
					}
				}
			}
		}

		if len(matchLines[i].MatchNames) > 0 {
			result = append(result, matchLines[i])
		}
	}

	return result
}

func fileNameHash(p string) string {
	hash := md5.Sum([]byte(p))
	return string([]rune(hex.EncodeToString(hash[:]))[0:5])
}

// func findAndMarkAsMatches(logger *zerolog.Logger, l *[]string, x string) {

// 	fx := "[" + x + "]"
// 	logger.Trace().Msgf("*** Start search [%v]", x)

// 	for i := 0; i < len(*l); i++ {
// 		v := (*l)[i]

// 		if strings.HasSuffix(v, fx) {
// 			logger.Trace().Msgf("*** Found and modify line %v", v)
// 			v = strings.Replace(v, "| ][", "|*][", 1)
// 			(*l)[i] = v
// 		}
// 	}
// }
