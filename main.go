package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

var (
	DatedFile bool
)

func init() {
	flag.BoolVar(&DatedFile, "dated-file", true, "Include date in output file name (e.g., scans_1970-01-01.csv)")
	flag.Parse()
}

func filename() string {
	if DatedFile {
		return fmt.Sprintf("scans_%s.csv", time.Now().Format(time.DateOnly))
	}
	return "scans.csv"
}

func writeToFile(b []byte) {
	file := filename()

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file", "filename", file)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s,%s\n", time.Now().Format(time.DateTime), b)
}

func readInputNoEcho() {
	for {
		in, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			slog.Error("error reading input",
				slog.String("error", err.Error()))
		}
		slog.Info("", "input", in)
		writeToFile(in)
	}
}

func readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			in := scanner.Bytes()
			slog.Info("", "input", in)
			writeToFile(in)
		}
	}
}

func main() {
	slog.Info("ready to scan")
	readInputNoEcho()
}
