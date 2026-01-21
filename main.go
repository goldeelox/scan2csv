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
	OutputDir string
	Version   string = "dev"
	Revision  string = "unknown"
)

func init() {
	flag.BoolVar(&DatedFile, "dated-file", true, "Include date in output file name (e.g., scans_1970-01-01.csv)")
	flag.StringVar(&OutputDir, "output-dir", ".", "Directory to write CSV files to")
	flag.Parse()
}

func filename() string {
	if DatedFile {
		return fmt.Sprintf("%s/scans_%s.csv", OutputDir, time.Now().Format(time.DateOnly))
	}
	return fmt.Sprintf("%s/scans.csv", OutputDir)
}

func writeToFile(b []byte) {
	if isEmpty(b) {
		slog.Info("ignoring empty input")
		return
	}

	file := filename()
	slog.Info("writing to file",
		slog.String("path", file),
		slog.Any("input", b))

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file", "filename", file)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s,%s\n", time.Now().Format(time.DateTime), b)
}

func isEmpty(b []byte) bool {
	return len(b) == 0
}

func readInputNoEcho() {
	for {
		in, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			slog.Error("error reading input",
				slog.String("error", err.Error()))
		}

		writeToFile(in)
	}
}

func readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			writeToFile(scanner.Bytes())
		}
	}
}

func main() {
	fmt.Printf("scan2csv version=%s revision=%s\n", Version, Revision)
	slog.Info("ready to scan")
	readInputNoEcho()
}
