package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	rootfsDir = "rootfs"
	procPath  = "rootfs/proc"
	hostName  = "minicontainer"
)

func Child() {
	log := func(msg string) { fmt.Println("[child]", msg) }

	log("starting container setup...")

	log("setting up overlay...")
	setupOverlay()

	log("binding host base system into rootfs...")
	if _, err := bindBaseSystem(false); err != nil {
		fatal(err)
	}

	log("setting up cgroups...")
	setupCgroups()

	log("setting hostname...")
	if err := syscall.Sethostname([]byte(hostName)); err != nil {
		fatal(err)
	}

	log("ensuring /proc dir + mount...")
	if err := mountProc(); err != nil {
		fatal(err)
	}

	// Enter new root
	if err := syscall.Chroot(rootfsDir); err != nil {
		fatal(err)
	}
	if err := os.Chdir("/"); err != nil {
		fatal(err)
	}

	setupEnv()

	cmdArgs, fallbackShell := resolveCommand(os.Args[2:])
	exitCode := runCommand(cmdArgs, fallbackShell)

	// Unmount /proc (path is /proc now inside chroot)
	if err := syscall.Unmount("/proc", 0); err != nil && !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.ENOENT) {
		fmt.Fprintf(os.Stderr, "[child] unmount /proc: %v\n", err)
	}

	os.Exit(exitCode)
}

// mountProc mounts proc inside rootfs prior to chroot.
func mountProc() error {
	if err := os.MkdirAll(procPath, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", procPath, err)
	}
	if err := syscall.Mount("proc", procPath, "proc", 0, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}
	return nil
}

func setupEnv() {
	_ = os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin")
	_ = os.Setenv("PS1", "# ")
}

func resolveCommand(args []string) (final []string, fallback bool) {
	if len(args) == 0 {
		fmt.Println("[child] no command provided; using /bin/bash")
		return []string{"/bin/bash"}, true
	}
	return args, false
}

func runCommand(cmdArgs []string, fallbackShell bool) int {
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		code := extractExitCode(err)
		if fallbackShell && code != 0 {
			// Normalize interactive shell exit anomalies
			return 0
		}
		return code
	}
	return 0
}

func extractExitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus()
		}
		return 1
	}
	fmt.Fprintf(os.Stderr, "[child] exec error: %v\n", err)
	return 1
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[child] fatal: %v\n", err)
	os.Exit(1)
}

func bindBaseSystem(readOnly bool) ([]string, error) {
	toBind := []string{"/bin", "/lib", "/lib64", "/usr", "/dev"}
	var mounted []string
	for _, src := range toBind {
		if _, err := os.Stat(src); err != nil {
			continue
		}
		target := rootfsDir + src
		if fi, err := os.Lstat(target); err == nil && !fi.IsDir() {
			if err := os.Remove(target); err != nil {
				return mounted, fmt.Errorf("remove non-dir %s: %w", target, err)
			}
		}
		if err := os.MkdirAll(target, 0o755); err != nil {
			return mounted, fmt.Errorf("mkdir %s: %w", target, err)
		}
		if err := syscall.Mount(src, target, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			return mounted, fmt.Errorf("bind %s -> %s: %w", src, target, err)
		}
		if readOnly {
			if err := syscall.Mount("", target, "", syscall.MS_BIND|syscall.MS_REMOUNT|syscall.MS_RDONLY, ""); err != nil {
				return mounted, fmt.Errorf("remount ro %s: %w", target, err)
			}
		}
		mounted = append(mounted, target)
	}
	return mounted, nil
}
