package prefetcher

import (
	"fmt"
	"internal/common"
	"path"
)

type PrefetchFactory struct {
}

func NewPrefetchFactory() *PrefetchFactory {
	return &PrefetchFactory{}
}

func (f *PrefetchFactory) CreatePrefetchMatcher(config *common.ServerConfig, matcherConfig common.MatcherConfig) (PrefetchMatcher, error) {
	var result PrefetchMatcher
	switch matcherConfig.Type {
	case "anchor":
		result = &PrefetchMatcherAnchor{
			file:     path.Join(config.SrcDir, matcherConfig.File),
			anchor:   matcherConfig.AnchorRegex,
			format:   matcherConfig.Format,
			regexStr: matcherConfig.Regex,
			maxLine:  matcherConfig.MaxLines,
		}
	case "regex":
		result = &PrefetchMatcherRegex{
			file:   path.Join(config.SrcDir, matcherConfig.File),
			regex:  matcherConfig.Regex,
			format: matcherConfig.Format,
		}
	case "hardcoded":
		result = &PrefetchMatcherHardcoded{
			hardcoded: matcherConfig.Format,
		}
	case "":
		result = &PrefetchMatcherNil{}
	default:
		return nil, fmt.Errorf("unsupported matcher name: %s", matcherConfig.Type)
	}

	return result, nil
}
