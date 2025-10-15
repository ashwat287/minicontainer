package main

import (
	"fmt"
	"os"
)

func setupCgroups() {
	fmt.Println("[cgroups] setting up resource limits...")

	// Check if cgroups v1 pids controller exists
	if _, err := os.Stat("/sys/fs/cgroup/pids"); err == nil {
		setupCgroupsV1()
	} else {
		setupCgroupsV2()
	}
}

func setupCgroupsV1() {
	fmt.Println("[cgroups] using cgroups v1...")
	cgroupPath := "/sys/fs/cgroup/pids/minicontainer"
	must(os.MkdirAll(cgroupPath, 0755))
	writeFile(cgroupPath+"/pids.max", "20")
	writeFile(cgroupPath+"/cgroup.procs", getPid())
}

func setupCgroupsV2() {
	fmt.Println("[cgroups] using cgroups v2...")
	cgroupPath := "/sys/fs/cgroup/minicontainer"
	must(os.MkdirAll(cgroupPath, 0755))

	// Try to enable pids controller (may fail if not available)
	os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+pids"), 0644)

	writeFile(cgroupPath+"/pids.max", "20")
	writeFile(cgroupPath+"/cgroup.procs", getPid())
}
