package simplelog

import (
	"log"
	"os"
)

// verbosity
const (
	_ = iota
	FATAL
	RESULT
	WARNING
	INFO
	VERBOSE
	DEBUG
)

var (
	getVerbosity  func() int
	loggerFatal   = log.New(os.Stderr, "[FATAL] ", 0)
	loggerResult  = log.New(os.Stdout, "", 0)
	loggerWarning = log.New(os.Stderr, "[!] ", 0)
	loggerInfo    = log.New(os.Stdout, "[+] ", 0)
	loggerVerbose = log.New(os.Stdout, "[>] ", 0)
	loggerDebug   = log.New(os.Stdout, "[D] ", 0)
)

// GetVerbosityWith sets the function used to load the current verbosity level
func GetVerbosityWith(v func() int) {
	getVerbosity = v
}

func assertInitialized() {
	if getVerbosity == nil {
		panic("Logger is not initialized!")
	}
}

// F logs FATAL errors to stderr and calls os.Exit(1)
func F(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= FATAL {
		loggerFatal.Fatalln(v...)
	}
}

// R logs RESULTS to stdout
func R(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= RESULT {
		loggerResult.Println(v...)
	}
}

// W logs WARNINGS to stderr
func W(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= WARNING {
		loggerWarning.Println(v...)
	}
}

// I logs INFO to stdout
func I(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= INFO {
		loggerInfo.Println(v...)
	}
}

// V logs VERBOSE messages to stdout
func V(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= VERBOSE {
		loggerVerbose.Println(v...)
	}
}

// D logs DEBUG messages to stdout
func D(v ...interface{}) {
	assertInitialized()
	if getVerbosity() >= DEBUG {
		loggerDebug.Println(v...)
	}
}

// FatalOnErr logs and calls os.Exit(1) if the error != nil
func FatalOnErr(s string, e error) {
	if e != nil {
		F("{"+s+"}", e)
	}
}

// WarnOnErr logs the error as a WARNING if the error != nil
func WarnOnErr(s string, e error) {
	if e != nil {
		W("{"+s+"}", e)
	}
}
