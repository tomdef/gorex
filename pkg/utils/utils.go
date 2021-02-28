package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// CreateLogger deliver logger instance
func CreateLogger(moduleName string, trace bool) zerolog.Logger {

	fmt.Printf("trace=%v\n", trace)

	if moduleName != "" {
		moduleName = fmt.Sprintf("› %v ›", moduleName)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.NoColor = false

	output.FormatCaller = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%v", moduleName))
	}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("[%-6s]", i))
	}

	output.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%v", i)
	}

	output.FormatFieldName = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("[%v:", i)
	}

	output.FormatFieldValue = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%v]", i)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if trace == true {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	return zerolog.New(output).With().Timestamp().Caller().Logger()
}

// RemoveFile backup file into new
func RemoveFile(logger zerolog.Logger, file string) {
	info, err := os.Stat(file)
	if !os.IsNotExist(err) {
		newName := file + ".backup"
		logger.Info().Msgf("\tRename previous file [%v] to [%v]", info.Name(), newName)
		os.Rename(file, newName)
	}
}
