// Package xlog is a user-defined log, include multi-level highlight support, output redirection support
package xlog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var colorMap = map[string]int{
	"RED":       31,
	"red":       31,
	"GREEN":     32,
	"green":     32,
	"YELLOW":    33,
	"yellow":    33,
	"BLUE":      34,
	"blue":      34,
	"PURPLE":    35,
	"purple":    35,
	"DARKGREEN": 36,
	"darkgreen": 36,
	"WHITE":     37,
	"white":     37,
}

func getPatternMono(word, color string) string {
	return fmt.Sprintf("\x1b[%dm%s \x1b[0m", colorMap[color], word)
}

func getPatternMix(word, foreground, background string) string {
	return fmt.Sprintf("\x1b[%d;%dm%s \x1b[0m", colorMap[background], colorMap[foreground], word)

}

var (
	flag     = log.LstdFlags
	debugLog = log.New(os.Stdout, getPatternMono("[Debug]", "white"), flag|log.Lshortfile)
	errorLog = log.New(os.Stdout, getPatternMono("[Error]", "red"), flag)
	infoLog  = log.New(os.Stdout, getPatternMono("[Info]", "blue"), flag)
	fatalLog = log.New(os.Stdout, getPatternMix("[Fatal]", "purple", "white"), flag)
	warnlLog = log.New(os.Stdout, getPatternMix("[Warning]", "yellow", "white"), flag)

	loggers = []*log.Logger{errorLog, infoLog, fatalLog, debugLog, warnlLog}
	mu      sync.Mutex
)

//log alias
var (
	Warn   = warnlLog.Print
	Warnln = warnlLog.Println
	Warnf  = warnlLog.Printf

	Println = debugLog.Println
	Print   = debugLog.Print
	Printf  = debugLog.Printf

	Errorln = errorLog.Println
	Error   = errorLog.Print
	Errorf  = errorLog.Printf

	Infoln = infoLog.Println
	Infof  = infoLog.Printf
	Info   = infoLog.Print

	Fatalln = fatalLog.Fatalln
	Fatalf  = fatalLog.Fatalf
	Fatal   = fatalLog.Fatal
)

//log levels
const (
	Disabled = iota
	FatalLevel
	ErrorLevel
	InfoLevel
	DebugLevel
)

//SetLevel controls log level, the higher the level is, the colorful the output becomes from
//fatal->error->info->debug
func SetLevel(level int, w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if w == nil {
		w = os.Stdout
	}
	for _, logger := range loggers {
		logger.SetOutput(w)
	}

	if level < DebugLevel {
		debugLog.SetOutput(ioutil.Discard)
	}
	if level < InfoLevel {
		infoLog.SetOutput(ioutil.Discard)
	}
	if level < ErrorLevel {
		errorLog.SetOutput(ioutil.Discard)
	}
	if level < FatalLevel {
		errorLog.SetOutput(ioutil.Discard)
	}
}
