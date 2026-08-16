package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Terminal struct {
	reader *bufio.Reader
}

const (
	colorDim   = "\033[2m"
	colorReset = "\033[0m"
)

func New(r io.Reader) Terminal {
	return Terminal{reader: bufio.NewReader(r)}
}

func (t Terminal) ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J")
}

func (t Terminal) ReadChoice() string {
	fmt.Println("Choose an option:")
	return t.readLine()
}

func (t Terminal) ReadInput(label string) string {
	fmt.Printf("%s: ", label)
	return t.readLine()
}

func (t Terminal) ReadInputWithDefault(label, defaultValue string) string {
	fmt.Printf("%s [%s%s%s]: ", label, colorDim, defaultValue, colorReset)

	value := t.readLine()
	if value == "" {
		return defaultValue
	}

	return value
}

func (t Terminal) ReadList(label string) []string {
	fmt.Printf("%s (comma-separated): ", label)
	raw := t.readLine()

	var items = []string{}
	for item := range strings.SplitSeq(raw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func (t Terminal) readLine() string {
	value, err := t.reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(value)
}

func (t Terminal) WaitForEnter() {
	fmt.Println("\nPress Enter to continue...")
	t.reader.ReadString('\n')
}
