package main

type onehotEmbeding struct {
	lastIndex int
	wordmap   map[string]int
	vectormap map[int]string
}

func newOnehotEmbeding() onehotEmbeding {
	oh := onehotEmbeding{
		lastIndex: 0,
		wordmap:   map[string]int{},
		vectormap: map[int]string{},
	}

	return oh
}

func (embeding *onehotEmbeding) onehot(src string) int {
	ind, ok := embeding.wordmap[src]

	if ok {
		return ind
	} else {
		embeding.wordmap[src] = embeding.lastIndex
		embeding.vectormap[embeding.lastIndex] = src

		embeding.lastIndex += 1

		return embeding.lastIndex - 1
	}
}

func (embeding *onehotEmbeding) onehotBulk(src []string) []int {
	res := []int{}

	for _, embed := range src {
		res = append(res, embeding.onehot(embed))
	}

	return res
}

func (embeding *onehotEmbeding) containsWord(src string) bool {
	for _, word := range embeding.vectormap {
		if word == src {
			return true
		}
	}

	return false
}
