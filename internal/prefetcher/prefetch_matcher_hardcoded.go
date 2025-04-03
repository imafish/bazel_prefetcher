package prefetcher

type PrefetchMatcherHardcoded struct {
	hardcoded string
}

func (m *PrefetchMatcherHardcoded) Match() (bool, string, error) {
	return true, m.hardcoded, nil
}
