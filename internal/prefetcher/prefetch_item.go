package prefetcher

type PrefetchMatchers struct {
	Name        string
	UrlMatcher  PrefetchMatcher
	HashMatcher PrefetchMatcher
}

type PrefetchItem struct {
	// initial information
	Url  string
	Hash string

	// updated after download
	Path      string
	HashOfUrl string
	Size      int64

	Error error
}

type PrefetchMatcher interface {
	// Match checks if the given URL matches the prefetch item.
	// It returns true if it matches, false otherwise.
	Match() (bool, string, error)
}
