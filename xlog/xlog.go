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
	"RED" : 31,
	"red" : 31,
	"GREEN":32,
	"green":32,
	"YELLOW":33,
	"yellow":33,
	"BLUE":34,
	"blue":34,
	"PURPLE":35,
	"purple":35,
	"DARKGREEN":36,
	"darkgreen":36,
	"WHITE":37,
	"white":37,
}

func getPatternMono(word, color string) string{
	return fmt.Sprintf("\x1b[%dm%s \x1b[0m", colorMap[color], word)
}

func getPatternMix(word, foreground, background string) string{
	return fmt.Sprintf("\x1b[%d;%dm%s \x1b[0m",colorMap[background], colorMap[foreground], word)

}

var (
	normalLog = log.New(os.Stdout, getPatternMono("[Normal]", "white"), log.LstdFlags|log.Lshortfile)
	errorLog = log.New(os.Stdout, getPatternMono("[Error]", "red"), log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, getPatternMono("[Info]", "blue"), log.LstdFlags|log.Lshortfile)
	fatalLog = log.New(os.Stdout, getPatternMix("[Fatal]", "purple", "white"), log.LstdFlags|log.Lshortfile)

	loggers = []*log.Logger{errorLog, infoLog, fatalLog}
	mu      sync.Mutex
)

//log alias
var (
	Println  = normalLog.Println
	Print  = normalLog.Print
	Printf = normalLog.Printf

	Errorln  = errorLog.Println
	Error  = errorLog.Print
	Errorf = errorLog.Printf

	Infoln   = infoLog.Println
	Infof  = infoLog.Printf
	Info  = infoLog.Print

	Fatalln  = fatalLog.Fatalln
	Fatalf = fatalLog.Fatalf
	Fatal = fatalLog.Fatal
)

//log levels
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

//SetLevel controls log level
func SetLevel(level int, w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if w == nil {
		w = os.Stdout
	}
	for _, logger := range loggers {
		logger.SetOutput(w)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
}
