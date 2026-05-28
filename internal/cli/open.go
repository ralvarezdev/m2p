package cli

import (
	"os/exec"
	"runtime"
)

// openFile opens path in the default application for the current OS.
func openFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case OSWindows:
		cmd = exec.Command(OpenCmdWindows, OpenArgWindows, path)
	case OSDarwin:
		cmd = exec.Command(OpenCmdDarwin, path)
	default:
		cmd = exec.Command(OpenCmdLinux, path)
	}
	return cmd.Start()
}
