//go:build !unix

package record

import "os/exec"

func processAlive(pid int) bool {
	if pid <= 1 {
		return false
	}
	return false
}

func prepareDetachedChild(cmd *exec.Cmd) {}
