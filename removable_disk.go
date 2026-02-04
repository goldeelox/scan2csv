package main

import (
	"fmt"
	"io"
	"log"
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
	AutoMount  bool
}

func executableExistsInPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NewRemovableDisk(uuid string, automount bool) *RemovableDisk {
	requiredExecutables := []string{"lsblk", "udisksctl"}
	for _, e := range requiredExecutables {
		if !executableExistsInPath(e) {
			log.Printf("%s executable required to use removable disks is not found", e)
			os.Exit(1)
		}
	}

	return &RemovableDisk{
		UUID:       uuid,
		DevicePath: "/dev/disk/by-uuid/" + uuid,
		AutoMount:  automount,
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

func (d *RemovableDisk) isAttached() bool {
	_, err := os.Stat("/dev/disk/by-uuid/" + d.UUID)
	return err == nil
}

func (d *RemovableDisk) mount() error {
	cmd := exec.Command("udisksctl", "mount", "-b", d.DevicePath)
	_, err := cmd.Output()
	return err
}

func (d *RemovableDisk) unmount() error {
	if !d.isMounted() {
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
	return err
}

func (d *RemovableDisk) copyFile(src string) error {
	if !d.isMounted() {
		return nil
	}

	filename := filepath.Base(src)
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(d.mountPoint() + "/" + filename)
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	err = dest.Sync()
	if err != nil {
		return err
	}
	return nil
}
