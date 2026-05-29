package cli

import (
	"context"
	"os/exec"
	"runtime"
)

// openFile opens path in the default application for the current OS.
func openFile(path string) error {
	ctx := context.Background()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case OSWindows:
		cmd = exec.CommandContext(ctx, OpenCmdWindows, OpenArgWindows, path)
	case OSDarwin:
		cmd = exec.CommandContext(ctx, OpenCmdDarwin, path)
	default:
		cmd = exec.CommandContext(ctx, OpenCmdLinux, path)
	}
	return cmd.Start()
}
