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

func (f *PrefetchFactory) CreatePrefetchMatcher(srcDir string, matcherConfig common.MatcherConfig) (PrefetchMatcher, error) {
	var result PrefetchMatcher
	switch matcherConfig.Type {
	case "anchor":
		result = &PrefetchMatcherAnchor{
			file:     path.Join(srcDir, matcherConfig.File),
			anchor:   matcherConfig.AnchorRegex,
			format:   matcherConfig.Format,
			regexStr: matcherConfig.Regex,
			maxLine:  matcherConfig.MaxLines,
		}
	case "regex":
		result = &PrefetchMatcherRegex{
			file:   path.Join(srcDir, matcherConfig.File),
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

func CreatePrefetchersFromConfig(srcDir string, prefetchConfig *common.PrefetchConfig) ([]PrefetchMatchers, error) {

	prefetchFactory := NewPrefetchFactory()
	prefetchers := make([]PrefetchMatchers, 0, len(prefetchConfig.Items))
	for _, pf := range prefetchConfig.Items {
		urlMatcher, err := prefetchFactory.CreatePrefetchMatcher(srcDir, pf.UrlMatcherConfig)
		if err != nil {
			return nil, err
		}
		hashMatcher, err := prefetchFactory.CreatePrefetchMatcher(srcDir, pf.HashMatcherConfig)
		if err != nil {
			return nil, err
		}

		prefetchers = append(prefetchers, PrefetchMatchers{Name: pf.Name, UrlMatcher: urlMatcher, HashMatcher: hashMatcher})
	}

	return prefetchers, nil
}
