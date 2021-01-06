package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// CreateLogger deliver logger instance
func CreateLogger(moduleName string) zerolog.Logger {

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

	return zerolog.New(output).With().Timestamp().Caller().Logger()
}
