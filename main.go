package main

import (
	"context"
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
	Version  string = "dev"
	Revision string = "unknown"
)

func writeToFile(wg *sync.WaitGroup, b []byte, file string) {
	defer wg.Done()

	if len(b) == 0 {
		slog.Info("ignoring empty input")
		return
	}

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

func (c Config) backup(ctx context.Context, wg *sync.WaitGroup, disk *RemovableDisk) {
	defer wg.Done()

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

			fsys := os.DirFS(c.OutputDir)
			matches, _ := fs.Glob(fsys, "scans*.csv")
			for _, match := range matches {
				disk.copyFile(c.OutputDir + "/" + match)
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
	var wg sync.WaitGroup
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	fmt.Printf("scan2csv version=%s revision=%s\n", Version, Revision)

	config := ParseConfig()

	if !config.RemovableDisksSupported {
		fmt.Printf("removable disk support not available in %s\n", runtime.GOOS)
	}

	if config.RemovableDisksSupported && config.DetectRemovableDisks {
		detectRemovableDisks()
		return
	}

	if config.RemovableDisksSupported && len(config.RemovableDiskUUID) > 0 {
		disk := NewRemovableDisk(config.RemovableDiskUUID, config.AutoMount)
		wg.Add(1)
		go config.backup(ctx, &wg, disk)
	}

	slog.Info("ready to scan")
	go func() {
		for {
			in, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				slog.Error("error reading input",
					slog.String("error", err.Error()))
			}

			var file string
			if config.DatedFile {
				file = fmt.Sprintf("%s/scans_%s.csv", config.OutputDir, time.Now().Format(time.DateOnly))
			} else {
				file = fmt.Sprintf("%s/scans.csv", config.OutputDir)
			}
			wg.Add(1)
			writeToFile(&wg, in, file)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Warn("shutting down")
		wg.Wait()
	}
}
