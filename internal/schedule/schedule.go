package schedule

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const label = "com.rack-cli.expiry"

var plistTmpl = template.Must(template.New("plist").
	Funcs(template.FuncMap{"esc": html.EscapeString}).
	Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{esc .Bin}}</string>
		<string>--check</string>
	</array>
	<key>StartCalendarInterval</key>
	<dict>
		<key>Hour</key><integer>9</integer>
		<key>Minute</key><integer>0</integer>
	</dict>
	<key>StandardErrorPath</key><string>{{esc .Log}}</string>
</dict>
</plist>
`))

func EnsureInstalled(logDir string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("notifications are only supported on macOS")
	}

	bin, err := os.Executable()
	if err != nil {
		return err
	}
	if bin, err = filepath.EvalSymlinks(bin); err != nil {
		return err
	}
	if strings.Contains(bin, "go-build") {
		return errors.New("run the installed binary (go install), not go run")
	}

	var want bytes.Buffer
	data := map[string]string{
		"Label": label,
		"Bin":   bin,
		"Log":   filepath.Join(logDir, "expiry.log"),
	}
	if err := plistTmpl.Execute(&want, data); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	agents := filepath.Join(home, "Library", "LaunchAgents")
	plistPath := filepath.Join(agents, label+".plist")

	if current, err := os.ReadFile(plistPath); err == nil && bytes.Equal(current, want.Bytes()) {
		return nil
	}

	if err := os.MkdirAll(agents, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(plistPath, want.Bytes(), 0o644); err != nil {
		return err
	}

	domain := fmt.Sprintf("gui/%d", os.Getuid())
	_ = exec.Command("launchctl", "bootout", domain+"/"+label).Run()

	if out, err := exec.Command("launchctl", "bootstrap", domain, plistPath).CombinedOutput(); err != nil {
		_ = os.Remove(plistPath)
		return fmt.Errorf("launchctl bootstrap: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
