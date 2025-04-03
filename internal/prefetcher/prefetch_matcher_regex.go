package prefetcher

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"internal/common"
)

type PrefetchMatcherRegex struct {
	file   string
	regex  string
	format string
}

func (m *PrefetchMatcherRegex) Match() (bool, string, error) {
	l := common.NewLoggerWithPrefixAndColor("PrefetchMatcherRegex: ")
	l.Printf("Analyzing file: %s", m.file)

	pattern, err := regexp.Compile(m.regex)
	if err != nil {
		return false, "", err
	}

	content, err := os.ReadFile(m.file)
	if err != nil {
		log.Printf("Failed to read file %s: %v", m.file, err)
		return false, "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			result := fmt.Sprintf(m.format, matches[1])
			return true, result, nil
		}
	}

	l.Printf("No match found for item %s in file %s", m.regex, m.file)
	return false, "", nil
}
