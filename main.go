package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

var (
	Version  string = "dev"
	Revision string = "unknown"
)

func loadRows(filename string) []table.Row {
	source, _ := os.Open(filename)
	defer source.Close()

	source.Seek(-500, io.SeekEnd)
	scanner := bufio.NewScanner(source)
	// skip first line in case we missed the start
	scanner.Scan()
	var rows []table.Row
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), ",")
		rows = append(rows, split)
	}

	return rows
}

func main() {
	config := ParseConfig()

	if config.Version {
		fmt.Printf("version=%s revision=%s\n", Version, Revision)
		return
	}
	if !config.RemovableDisksSupported {
		fmt.Printf("removable disk support not available\n")
		os.Exit(1)
	}
	if config.RemovableDisksSupported && config.DetectRemovableDisks {
		detectRemovableDisks()
		return
	}

	m := New(config)
	p := tea.NewProgram(m)

	if config.RemovableDisksSupported && len(config.RemovableDiskUUID) > 0 {
		disk := NewRemovableDisk(config.RemovableDiskUUID, config.AutoMount)
		go func() {
			for {
				time.Sleep(2 * time.Second)

				if disk.AutoMount && disk.isAttached() {
					if err := disk.mount(); err != nil {
						p.Send(err)
					}
				}

				if !disk.isMounted() {
					continue
				}

				m.wg.Add(1)
				func() {
					defer m.wg.Done()
					fsys := os.DirFS(m.config.OutputDir)
					matches, _ := fs.Glob(fsys, "scans*.csv")
					for _, match := range matches {
						disk.copyFile(m.config.OutputDir + "/" + match)
					}
					p.Send(StatusOk("backup complete"))

					if err := disk.unmount(); err != nil {
						p.Send(err)
						return
					}

					if err := disk.poweroff(); err != nil {
						p.Send(err)
					}
				}()
			}
		}()
	}

	if _, err := p.Run(); err != nil {
		log.Printf("ut oh %v", err)
	}

	log.Printf("waiting for all jobs to finish...")
	m.wg.Wait()
	log.Printf("bye")
}
