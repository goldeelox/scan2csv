package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

var (
	AutoMount            bool
	DatedFile            bool
	DetectRemovableDisks bool
	OutputDir            string
	RemovableDiskUUID    string

	Version  string = "dev"
	Revision string = "unknown"
)

func init() {
	flag.BoolVar(&AutoMount, "automount", false, "Auto mount removable disk")
	flag.BoolVar(&DatedFile, "dated-file", true, "Include date in output file name (e.g., scans_1970-01-01.csv)")
	flag.BoolVar(&DetectRemovableDisks, "detect-removable-disks", false, "Scans for attached removable disks then exits")
	flag.StringVar(&OutputDir, "output-dir", ".", "Directory to write CSV files to")
	flag.StringVar(&RemovableDiskUUID, "uuid", "", "UUID of removable disk used to backup scan files (Attach removable disk and run with -detect-removable-disks to get UUID)")
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

func backup(ctx context.Context, wg *sync.WaitGroup, uuid string) {
	defer wg.Done()

	disk := NewRemovableDisk(uuid, AutoMount)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
			if disk.AutoMount && disk.isAttached() {
				slog.Info("removable disk detected",
					slog.String("UUID", disk.UUID))
				disk.mount()
			}

			if !disk.isMounted() {
				continue
			}

			slog.Info("removable disk is mounted",
				slog.String("UUID", disk.UUID),
				slog.String("mountPoint", disk.mountPoint()))

			fsys := os.DirFS(OutputDir)
			matches, _ := fs.Glob(fsys, "scans*.csv")
			for _, match := range matches {
				disk.copyFile(OutputDir + "/" + match)
			}
			slog.Info("backup complete")

			if err := disk.unmount(); err != nil {
				slog.Error("error unmounting disk",
					slog.String("msg", err.Error()))
				continue
			}

			if err := disk.poweroff(); err != nil {
				slog.Error("error powering off disk",
					slog.String("msg", err.Error()))
			}
		}
	}
}

func main() {
	removableDiskSupported := runtime.GOOS == "linux"

	if DetectRemovableDisks {
		if !removableDiskSupported {
			fmt.Printf("removable disk support not available in %s\n", runtime.GOOS)
			return
		}
		detectRemovableDisks()
		return
	}

	fmt.Printf("scan2csv version=%s revision=%s\n", Version, Revision)

	var wg sync.WaitGroup
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	if RemovableDiskUUID != "" {
		if removableDiskSupported {
			wg.Add(1)
			go backup(ctx, &wg, RemovableDiskUUID)
		} else {
			slog.Warn("removable disk support not available for this OS")
		}
	}

	slog.Info("ready to scan")
	go readInputNoEcho()

	select {
	case <-ctx.Done():
		slog.Warn("shutting down")
		wg.Wait()
	}
}
