package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	mystrings "github.com/5vnetwork/vx-core/common/strings"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

// TODO: upload logger
const ipv4Pattern = `\b(?:\d{1,3}\.){3}\d{1,3}\b`
const ipv6Pattern = `\b(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{1,4}\b|\b(?:[0-9a-fA-F]{1,4}:)*::(?:[0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{1,4}\b`
const domainPattern = `\b(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}\b`

type Logger struct {
	files          []*os.File
	redactionRegex *regexp.Regexp
}

func SetLog(config *configs.LoggerConfig) (*Logger, error) {
	if config == nil {
		config = &configs.LoggerConfig{
			LogLevel: configs.Level_DISABLED,
		}
	}

	l := &Logger{}
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	prelogger := log.With()

	if config.ShowCaller {
		// file name and line number
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return filepath.Base(file) + ":" + strconv.Itoa(line)
		}
		prelogger = prelogger.Caller()
	}

	log.Logger = prelogger.Logger()

	var output io.Writer
	if config.FilePath != "" {
		f, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = f
		l.files = append(l.files, f)
	} else if config.LogFileDir != "" {
		f, err := os.OpenFile(filepath.Join(config.LogFileDir, time.Now().Format("2006-01-02T15:04:05")+".txt"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = f
		l.files = append(l.files, f)
	} else {
		output = os.Stderr
	}

	if config.ConsoleWriter {
		writer := zerolog.ConsoleWriter{Out: output, NoColor: !config.ShowColor, TimeFormat: time.TimeOnly}
		if config.Redact {

			l.redactionRegex = regexp.MustCompile(ipv4Pattern + "|" + ipv6Pattern + "|" + domainPattern)

			writer.FormatFieldValue = func(i interface{}) string {
				if i == nil {
					return ""
				}
				if err, ok := i.(error); ok {
					return l.RedactSensitiveData(err.Error())
				}
				return l.RedactSensitiveData(mystrings.ToString(i))
			}
			writer.FormatErrFieldValue = writer.FormatFieldValue
		}
		log.Logger = log.Output(writer)
	} else {
		log.Logger = log.Output(output)
	}

	return l, nil
}

func (l *Logger) Close() error {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	if l.files != nil {
		for _, f := range l.files {
			f.Close()
		}
	}
	return nil
}

func (l *Logger) RedactSensitiveData(input string) string {
	return l.redactionRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Check if it's an IPv6 address
		if strings.Contains(match, ":") {
			// IPv6 - handle compressed addresses with ::
			if strings.Contains(match, "::") {
				// For compressed IPv6, redact the first half of visible parts
				parts := strings.Split(match, "::")
				if len(parts) == 2 {
					leftParts := strings.Split(parts[0], ":")
					rightParts := strings.Split(parts[1], ":")

					// Redact first half of left parts
					if len(leftParts) > 0 && leftParts[0] != "" {
						halfLen := len(leftParts) / 2
						if halfLen == 0 {
							halfLen = 1
						}
						leftParts = leftParts[:halfLen]
					}

					// Redact first half of right parts
					if len(rightParts) > 0 && rightParts[0] != "" {
						halfLen := len(rightParts) / 2
						if halfLen == 0 {
							halfLen = 1
						}
						rightParts = rightParts[:halfLen]
					}

					leftStr := strings.Join(leftParts, ":")
					rightStr := strings.Join(rightParts, ":")

					if leftStr != "" && rightStr != "" {
						return leftStr + "::" + rightStr + ":***"
					} else if leftStr != "" {
						return leftStr + "::***"
					} else if rightStr != "" {
						return "::" + rightStr + ":***"
					} else {
						return "::***"
					}
				}
			} else {
				// Full IPv6 without compression
				parts := strings.Split(match, ":")
				halfLen := len(parts) / 2
				if halfLen == 0 {
					halfLen = 1
				}
				return strings.Join(parts[:halfLen], ":") + ":***"
			}
		} else if strings.Contains(match, ".") && !strings.Contains(match, ":") {
			// Check if it's an IPv4 address (4 numeric parts) or domain
			parts := strings.Split(match, ".")
			if len(parts) == 4 {
				// Check if all parts are numeric (IPv4)
				isIPv4 := true
				for _, part := range parts {
					if len(part) == 0 || len(part) > 3 {
						isIPv4 = false
						break
					}
					for _, char := range part {
						if char < '0' || char > '9' {
							isIPv4 = false
							break
						}
					}
					if !isIPv4 {
						break
					}
				}

				if isIPv4 {
					// IPv4 - redact first half
					halfLen := len(parts) / 2
					return strings.Join(parts[:halfLen], ".") + ".***"
				}
			}

			// Domain - redact subdomain part but keep main domain
			if len(parts) >= 2 {
				// Keep only the final part (.tld) and redact the rest
				return "***." + strings.Join(parts[len(parts)-1:], ".")
			}
			// Fallback for single part domains
			return "***." + match
		}
		return match // fallback
	})
}

// func (l *Logger) SetLevel(level zerolog.Level) {
// 	l.currentLevel = level
// 	zerolog.SetGlobalLevel(level)
// 	log.Info().Msgf("log level changed to %s", level)
// }

// func (l *Logger) AddOutputToGlobalLogger(writer io.Writer) error {
// 	l.Lock()
// 	defer l.Unlock()
// 	l.allOutputs = append(l.allOutputs, writer)
// 	multi := zerolog.MultiLevelWriter(l.allOutputs...)
// 	log.Logger = log.Output(multi)
// 	return nil
// }

// func (l *Logger) RemoveOutputFromGlobalLogger(writer io.Writer) error {
// 	l.Lock()
// 	defer l.Unlock()
// 	for i, output := range l.allOutputs {
// 		if output == writer {
// 			l.allOutputs = append(l.allOutputs[:i], l.allOutputs[i+1:]...)
// 			break
// 		}
// 	}
// 	multi := zerolog.MultiLevelWriter(l.allOutputs...)
// 	log.Logger = log.Output(multi)
// 	return nil
// }

// func (l *Logger) WithHook(hook zerolog.Hook) *Logger {
// 	l.Lock()
// 	defer l.Unlock()
// 	l.hooks = append(l.hooks, hook)
// 	log.Logger = log.Logger.Hook(hook)
// 	return l
// }

// func (l *Logger) RemoveHook(hook zerolog.Hook) *Logger {
// 	l.Lock()
// 	defer l.Unlock()
// 	for i, h := range l.hooks {
// 		if h == hook {
// 			l.hooks = append(l.hooks[:i], l.hooks[i+1:]...)
// 			break
// 		}
// 	}
// 	log.Logger = log.Logger.Hook(l.hooks...)
// 	return l
// }
