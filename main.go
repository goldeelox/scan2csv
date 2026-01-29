package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

var (
	DatedFile            bool
	OutputDir            string
	DetectRemovableDisks bool
	RemovableDiskUUID    string
	Version              string = "dev"
	Revision             string = "unknown"
)

func init() {
	flag.BoolVar(&DatedFile, "dated-file", true, "Include date in output file name (e.g., scans_1970-01-01.csv)")
	flag.StringVar(&OutputDir, "output-dir", ".", "Directory to write CSV files to")
	flag.StringVar(&RemovableDiskUUID, "uuid", "", "UUID of removable disk used to backup scan files (Attach removable disk and run with -detect-removable-disks to get UUID)")
	flag.BoolVar(&DetectRemovableDisks, "detect-removable-disks", false, "Scans for attached removable disks then exits")
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

func backup(uuid string) {
	disk := NewRemovableDisk(uuid)

	for {
		if disk.isMounted() {
			slog.Info("removable disk detected",
				slog.String("UUID", disk.UUID),
				slog.String("path", disk.mountPoint()))

			fsys := os.DirFS(OutputDir)
			matches, _ := fs.Glob(fsys, "scans*.csv")

			for _, match := range matches {
				disk.copyFile(OutputDir + "/" + match)
			}

			slog.Info("backup complete")
			disk.unmount()
			disk.poweroff()
		}
		time.Sleep(2 * time.Second)
	}
}

func main() {
	if DetectRemovableDisks {
		detectRemovableDisks()
		return
	}

	fmt.Printf("scan2csv version=%s revision=%s\n", Version, Revision)

	if RemovableDiskUUID != "" {
		go backup(RemovableDiskUUID)
	}

	slog.Info("ready to scan")
	readInputNoEcho()
}
