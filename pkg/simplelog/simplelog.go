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
	// GetVerbosity returns the current verbosity level
	GetVerbosity func() int

	// LoggerFatal is written to for FATAL level logging
	LoggerFatal = log.New(os.Stderr, "[FATAL] ", 0)

	// LoggerResult is written to for RESULT level logging
	LoggerResult = log.New(os.Stdout, "", 0)

	// LoggerWarning is written to for WARNING level logging
	LoggerWarning = log.New(os.Stderr, "[!] ", 0)

	// LoggerInfo is written to for INFO level logging
	LoggerInfo = log.New(os.Stdout, "[+] ", 0)

	// LoggerVerbose is written to for VERBOSE level logging
	LoggerVerbose = log.New(os.Stdout, "[>] ", 0)

	// LoggerDebug is written to for DEBUG level logging
	LoggerDebug = log.New(os.Stdout, "[D] ", 0)
)

// GetVerbosityWith sets the function used to load the current verbosity level
func GetVerbosityWith(v func() int) {
	GetVerbosity = v
}

func assertInitialized() {
	if GetVerbosity == nil {
		panic("Logger is not initialized!")
	}
}

// F logs FATAL errors to stderr and calls os.Exit(1)
func F(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= FATAL {
		LoggerFatal.Fatalln(v...)
	}
}

// R logs RESULTS to stdout
func R(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= RESULT {
		LoggerResult.Println(v...)
	}
}

// W logs WARNINGS to stderr
func W(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= WARNING {
		LoggerWarning.Println(v...)
	}
}

// I logs INFO to stdout
func I(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= INFO {
		LoggerInfo.Println(v...)
	}
}

// V logs VERBOSE messages to stdout
func V(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= VERBOSE {
		LoggerVerbose.Println(v...)
	}
}

// D logs DEBUG messages to stdout
func D(v ...interface{}) {
	assertInitialized()
	if GetVerbosity() >= DEBUG {
		LoggerDebug.Println(v...)
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
