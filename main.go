package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
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

func writeToFile(s string) {
	file := filename()

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file", "filename", file)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s,%s\n", time.Now().Format(time.DateTime), s)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	slog.Info("ready to scan")
	for {
		if scanner.Scan() {
			in := scanner.Text()
			slog.Info("", "input", in)
			writeToFile(in)
		}
	}
}
