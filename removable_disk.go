package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func detectRemovableDisks() {
	cmd := exec.Command("lsblk", "-o", "label,uuid,mountpoint", "--filter", "RM == 1")
	output, _ := cmd.Output()
	fmt.Println(string(output))
}

type RemovableDisk struct {
	UUID       string
	DevicePath string
}

func executableExistsInPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NewRemovableDisk(uuid string) *RemovableDisk {
	requiredExecutables := []string{"lsblk", "udisksctl"}
	for _, e := range requiredExecutables {
		if !executableExistsInPath(e) {
			slog.Error("executable required to use removable disks is not found",
				slog.String("executable", e))
			os.Exit(1)
		}
	}

	return &RemovableDisk{
		UUID:       uuid,
		DevicePath: "/dev/disk/by-uuid/" + uuid,
	}
}

func (d *RemovableDisk) mountPoint() string {
	cmd := exec.Command("lsblk", "-n", "-o", "mountpoint", d.DevicePath)
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output))
}

func (d *RemovableDisk) isMounted() bool {
	return len(d.mountPoint()) > 0
}

func (d *RemovableDisk) unmount() error {
	if !d.isMounted() {
		slog.Info("disk is not mounted",
			slog.String("UUID", d.UUID))
		return nil
	}

	cmd := exec.Command("udisksctl", "unmount", "-b", d.DevicePath)
	_, err := cmd.Output()
	return err
}

func (d *RemovableDisk) poweroff() error {
	if d.isMounted() {
		d.unmount()
	}

	cmd := exec.Command("udisksctl", "power-off", "-b", d.DevicePath)
	_, err := cmd.Output()
	if err == nil {
		slog.Info("disk can be safely removed",
			slog.String("UUID", d.UUID))
	}
	return err
}

func (d *RemovableDisk) copyFile(src string) {
	if !d.isMounted() {
		slog.Error("disk is not mounted",
			slog.String("UUID", d.UUID))
		return
	}

	filename := filepath.Base(src)
	source, err := os.Open(src)
	if err != nil {
		slog.Error("error opening file",
			slog.String("source", src))
	}
	defer source.Close()

	dest, err := os.Create(d.mountPoint() + "/" + filename)
	defer dest.Close()

	slog.Info("copying file to removable disk",
		slog.String("source", filename),
		slog.String("destination", dest.Name()))

	_, err = io.Copy(dest, source)
	if err != nil {
		slog.Error("error copying file",
			slog.String("filename", filename),
			slog.String("msg", err.Error()))
	}

	err = dest.Sync()
	if err != nil {
		slog.Error("error syncing file to removable disk",
			slog.String("UUID", d.UUID),
			slog.String("filename", dest.Name()),
			slog.String("msg", err.Error()))
	}
}
