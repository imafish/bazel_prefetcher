package prefetcher

import (
	"fmt"
	"internal/common"
	"log"
	"os"
	"regexp"
	"strings"
)

type PrefetchMatcherAnchor struct {
	file     string
	anchor   string
	regexStr string
	maxLine  int
	format   string
}

func (m *PrefetchMatcherAnchor) Match() (bool, string, error) {
	l := common.NewLoggerWithPrefixAndColor("PrefetchMatcherAnchor: ")
	l.Printf("Analyzing file: %s", m.file)

	regex, err := regexp.Compile(m.regexStr)
	if err != nil {
		return false, "", err
	}

	anchorRegex, err := regexp.Compile(m.anchor)
	if err != nil {
		return false, "", err
	}

	content, err := os.ReadFile(m.file)
	if err != nil {
		log.Printf("Failed to read file %s: %v", m.file, err)
		return false, "", err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		matchedAnchor := anchorRegex.MatchString(line)
		if matchedAnchor {
			for j := i; j < i+m.maxLine; j++ {
				line2 := lines[j]
				if matches := regex.FindStringSubmatch(line2); len(matches) >= 2 {
					result := fmt.Sprintf(m.format, matches[1])
					l.Printf("Found match at line %d: %s", j, result)
					return true, result, nil
				}
			}
		}
	}

	l.Printf("No match found for item %s in file %s", m.regexStr, m.file)
	return false, "", nil
}
