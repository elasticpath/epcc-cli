package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenUrl(url string) error {
	if url == "" {
		return fmt.Errorf("No url available")
	}
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}
