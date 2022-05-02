package main

import (
	"regexp"
)

type crawlPolicy struct {
	CrawlReg   *regexp.Regexp
	CrawlWords []string
}

func newCrawlPolicy() crawlPolicy {
	cp := crawlPolicy{
		CrawlWords: []string{},
	}

	return cp
}

func (cp *crawlPolicy) addWord(word ...string) {
	cp.CrawlWords = append(cp.CrawlWords, word...)
	cp.compile()
}

func (cp *crawlPolicy) compile() {
	regstr := "("
	index := len(cp.CrawlWords) - 1

	for i, word := range cp.CrawlWords {
		regstr += word

		if index != i {
			regstr += "|"
		}
	}

	regstr += ")"
	cp.CrawlReg = regexp.MustCompile(regstr)
}

func (cp *crawlPolicy) getRelation(src string) int {
	return len(cp.CrawlReg.FindAllString(src, -1))
}
