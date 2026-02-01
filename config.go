package main

import (
	"flag"
	"runtime"
)

type Config struct {
	AutoMount               bool
	DatedFile               bool
	DetectRemovableDisks    bool
	OutputDir               string
	RemovableDiskUUID       string
	RemovableDisksSupported bool
}

func ParseConfig() Config {
	c := Config{
		RemovableDisksSupported: runtime.GOOS == "linux",
	}

	flag.BoolVar(&c.AutoMount, "automount", false, "Auto mount removable disk")
	flag.BoolVar(&c.DatedFile, "dated-file", true, "Include date in output file name (e.g., scans_1970-01-01.csv)")
	flag.BoolVar(&c.DetectRemovableDisks, "detect-removable-disks", false, "Scans for attached removable disks then exits")
	flag.StringVar(&c.OutputDir, "output-dir", ".", "Directory to write CSV files to")
	flag.StringVar(&c.RemovableDiskUUID, "uuid", "", "UUID of removable disk used to backup scan files (Attach removable disk and run with -detect-removable-disks to get UUID)")
	flag.Parse()

	return c
}
