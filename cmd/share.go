package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenUrl(cmUrl string) error {

	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", cmUrl).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", cmUrl).Start()
	case "darwin":
		exec.Command("open", cmUrl).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}
