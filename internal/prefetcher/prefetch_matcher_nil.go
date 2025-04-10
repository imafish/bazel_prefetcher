package prefetcher

type PrefetchMatcherNil struct {
}

func (p *PrefetchMatcherNil) Match() (bool, string, error) {
	return true, "", nil
}
