package clipboard

import (
	"errors"
	"io"
	"os/exec"
	"runtime"
)

var lookPath = exec.LookPath

func CopyFile(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	name, args, err := copyCommand(runtime.GOOS, path)
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func copyCommand(goos string, path string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "osascript", []string{"-e", `set the clipboard to (POSIX file "` + path + `")`}, nil
	default:
		if _, err := lookPath("xclip"); err == nil {
			return "xclip", []string{"-selection", "clipboard", "-t", "image/gif", "-i", path}, nil
		}
		if _, err := lookPath("wl-copy"); err == nil {
			return "sh", []string{"-c", "wl-copy --type image/gif < " + path}, nil
		}
		return "", nil, errors.New("no clipboard tool found (need xclip or wl-copy)")
	}
}
