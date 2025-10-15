package main

import (
	"fmt"
	"os"
	"syscall"
)

const (
	lowerDir  = "rootfs-lower"
	upperDir  = "rootfs-upper"
	workDir   = "rootfs-work"
	mergedDir = "rootfs"
)

func setupOverlay() {
	fmt.Println("[overlay] setting up overlay filesystem...")
	must(os.MkdirAll(lowerDir, 0755))
	must(os.MkdirAll(upperDir, 0755))
	must(os.MkdirAll(workDir, 0755))
	must(os.MkdirAll(mergedDir, 0755))

	opts := "lowerdir=" + lowerDir + ",upperdir=" + upperDir + ",workdir=" + workDir
	must(syscall.Mount("overlay", mergedDir, "overlay", 0, opts))
}
