package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func SetupLogger(debug bool) *zerolog.Logger {
	var zlog zerolog.Logger
	if debug {
		zerolog.TimestampFieldName = "Time"
		zerolog.LevelFieldName = "Level"
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
			return file + ":" + strconv.Itoa(line)
		}
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		consoleWriter.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		consoleWriter.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("__%s__", i)
		}
		consoleWriter.FormatFieldName = func(i interface{}) string {
			return fmt.Sprintf("%s:", i)
		}
		zlog = zerolog.New(consoleWriter).Level(zerolog.DebugLevel).With().Timestamp().Caller().Logger()
		return &zlog
	}
	zerolog.TimestampFieldName = "Time"
	zerolog.LevelFieldName = "Level"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	zlog = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()
	return &zlog
}
