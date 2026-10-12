package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

func Send(title, body string) error {
	script := fmt.Sprintf("display notification %s with title %s", quote(body), quote(title))
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func quote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
