package common

import (
	"fmt"
	"log"
	"strings"
)

const (
	lightBlue = "\033[38;5;39m"
	blue      = "\033[38;5;27m"
	orange    = "\033[38;5;202m"
	reset     = "\033[0m"
)

func Imafish() string {
	// ANSI color codes
	const (
		fishBlue  = "\033[38;5;27m"
		fishLight = "\033[38;5;39m"
		eyeWhite  = "\033[97m"
		eyeBlack  = "\033[30m"
		fins      = "\033[38;5;202m"
		tail      = "\033[38;5;208m"
		textPink  = "\033[38;5;205m"
	)

	// Create the fish logo line by line
	fishLines := []string{
		"",
		fishBlue + "                          ____",
		fishLight + "               _____/" + fishBlue + "____/" + fishLight + "|\\" + fishBlue + "_____",
		fishLight + "          _____/" + fishBlue + "_____________/ | \\" + fishLight + "_____",
		fishBlue + "         /                             \\",
		fishBlue + "        /" + fins + "><" + fishBlue + "                               \\",
		fishBlue + "       / " + eyeWhite + "o" + eyeBlack + "•" + fishBlue + "                               \\",
		fishBlue + "      |                                   |",
		fishBlue + "      |                                   |",
		fishBlue + "       \\                                 /",
		fishBlue + "        \\" + tail + "\\___________" + fishBlue + "                /",
		fishBlue + "         \\___________" + tail + "/" + fishBlue + "___________/",
		textPink + "                  I M A F I S H" + reset,
		"",
	}

	// Join all lines with newline characters
	return strings.Join(fishLines, "\n")
}

func LogSeparator(text string) {
	n := 21 - len(text)
	if n < 0 {
		n = 0
	} else {
		n = n / 2
	}
	log.Print(strings.Repeat("=", 21))
	log.Print(strings.Repeat(" ", n) + text)
	log.Print(strings.Repeat("=", 21))
}

type LoggerWithPrefix struct {
	Prefix string
}

func NewLoggerWithPrefixAndColor(prefix string) *LoggerWithPrefix {
	return &LoggerWithPrefix{Prefix: colorPrefix(prefix)}
}

func colorPrefix(prefix string) string {
	return lightBlue + prefix + reset
}

func (l *LoggerWithPrefix) SmallSeparator(text string, args ...interface{}) {
	n := 49 - len(text) - 4
	if n < 7 {
		n = 7
	}
	m := n / 2
	if n%2 == 1 {
		m = m + 1
	}
	n = n / 2

	log.Print(l.Prefix, blue, strings.Repeat(">", n), reset, "  ", fmt.Sprintf(text, args...), "  ", blue, strings.Repeat("<", m), reset)
}

func (l *LoggerWithPrefix) Printf(format string, args ...interface{}) {
	prefixedFormat := l.Prefix + format
	log.Printf(prefixedFormat, args...)
}
func (l *LoggerWithPrefix) Print(args ...interface{}) {
	prefixedMessage := l.Prefix + fmt.Sprint(args...)
	log.Print(prefixedMessage)
}
func (l *LoggerWithPrefix) Println(args ...interface{}) {
	prefixedMessage := l.Prefix + fmt.Sprint(args...)
	log.Println(prefixedMessage)
}
func (l *LoggerWithPrefix) Fatalf(format string, args ...interface{}) {
	prefixedFormat := l.Prefix + format
	log.Fatalf(prefixedFormat, args...)
}
func (l *LoggerWithPrefix) Fatal(args ...interface{}) {
	prefixedMessage := l.Prefix + fmt.Sprint(args...)
	log.Fatal(prefixedMessage)
}
