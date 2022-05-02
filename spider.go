package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type spider struct {
	linkToVector   map[string][]int
	vectorToLink   map[int][]string
	failLinkVector []string
	crawlLinks     int
	failLinks      int
	lq             linkQueue
	oh             onehotEmbeding
	times          []timeElapsed
	cp             *crawlPolicy
	maxLinks       int
	mutex          *sync.Mutex
}

func newSpider() spider {
	cp := newCrawlPolicy()

	spid := spider{
		linkToVector: map[string][]int{},
		vectorToLink: map[int][]string{},
		crawlLinks:   0,
		lq:           newLinkQueue(&cp),
		oh:           newOnehotEmbeding(),
		cp:           &cp,
		times:        []timeElapsed{},
		maxLinks:     100,
		mutex:        &sync.Mutex{},
	}

	return spid
}

func (spider *spider) containsVector(vector int) bool {
	for k := range spider.vectorToLink {
		if k == vector {
			return true
		}
	}

	return false
}

func (spider *spider) addLink(link string, vector []int) {
	spider.linkToVector[link] = vector

	for _, val := range vector {
		if spider.containsVector(val) {
			spider.vectorToLink[val] = append(spider.vectorToLink[val], link)
		} else {
			spider.vectorToLink[val] = []string{link}
		}
	}
}

func (spider spider) search(src string) searchResult {
	words := strings.Split(src, " ")
	vectors := []int{}

	res := searchResult{availableWords: []string{}}

	for _, word := range words {
		if spider.oh.containsWord(word) {
			vector := spider.oh.wordmap[word]
			res.availableWords = append(res.availableWords, word)

			vectors = append(vectors, vector)
		}
	}

	searchRes := map[string]int{}

	for _, vector := range vectors {
		links := spider.vectorToLink[vector]

		for _, link := range links {
			_, ok := searchRes[link]

			if !ok {
				searchRes[link] = 0
			}

			searchRes[link] += 1
		}
	}

	var ss []linkRelation
	for k, v := range searchRes {
		ss = append(ss, linkRelation{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Relation > ss[j].Relation
	})

	res.links = ss

	return res
}

type spiderInfo struct {
	vectorSize int
	linkSize   int
	failSize   int
}

func (spider spider) getSpiderInfo() spiderInfo {
	vectorKeys := make([]int, 0, len(spider.vectorToLink))

	for k := range spider.vectorToLink {
		vectorKeys = append(vectorKeys, k)
	}

	linkKeys := make([]string, 0, len(spider.linkToVector))

	for k := range spider.linkToVector {
		linkKeys = append(linkKeys, k)
	}

	return spiderInfo{
		vectorSize: len(vectorKeys),
		linkSize:   len(linkKeys),
		failSize:   spider.failLinks,
	}
}

func (spider spider) containsHistory(src string) bool {
	for k := range spider.linkToVector {
		if k == src {
			return true
		}
	}

	return false
}

func (spider *spider) crawlProcessor(worker int) {
	doneChans := make([]chan bool, worker)

	for i := 0; i < worker; i++ {
		doneChans[i] = make(chan bool)
		go spider.crawlManager(i, doneChans[i])
	}

	for i := 0; i < worker; i++ {
		<-doneChans[i]
	}
}

func (spider *spider) crawlManager(processid int, doneChannel chan bool) {
	banReg := regexp.MustCompile("(;amp)")

	topT := time.Now()
	loops := 0
bigLoop:
	for {

		if spider.crawlLinks >= spider.maxLinks {
			fmt.Println("Stop crawling")
			break
		}

		bannedEnds := []string{".css", ".png", ".exe", ".dmg", ".ico", ".gif", ".js", ".jpg"}

		top := ""

		for {
			loops += 1
			spider.mutex.Lock()
			top = spider.lq.getTop()

			if len(banReg.FindAllString(top, -1)) > 0 {
				spider.mutex.Unlock()
				continue bigLoop
			}

			if spider.containsHistory(top) {
				spider.mutex.Unlock()
				continue bigLoop
			}

			spider.mutex.Unlock()

			if top != "" {
				break
			}

			time.Sleep(time.Millisecond * 100)
		}
		cont := false

		for _, end := range bannedEnds {
			if endsWith(top, end) {
				cont = true
				break
			}
		}

		if cont {
			continue
		}

		isHtml := checkHtml(top)

		if !isHtml {
			continue
		}

		topTime := time.Since(topT).Milliseconds()

		topT = time.Now()
		spider.times = append(spider.times, timeElapsed{
			Details: map[string]int64{
				"FindTopTime": topTime,
				"Loops":       int64(loops),
			},
		})

		loops = 0

		spider.mutex.Lock()
		fmt.Println("[", processid, "] Crawl", spider.crawlLinks, "on", top, "Remains :", len(spider.lq.link))
		spider.crawlLinks += 1
		spider.mutex.Unlock()

		elapsed := timeElapsed{
			Link: top,
		}
		spider.crawl(top, &elapsed)
	}

	doneChannel <- true
}

func (spider *spider) crawl(entry string, elapsed *timeElapsed) {

	t := time.Now()

	resp, err := http.Get(entry)

	elapsed.GetTime = float64(time.Since(t).Milliseconds())

	if err != nil {
		spider.mutex.Lock()

		spider.failLinks += 1
		spider.failLinkVector = append(spider.failLinkVector, entry)
		spider.times = append(spider.times, *elapsed)

		spider.mutex.Unlock()
		elapsed.GetTime = float64(time.Since(t).Milliseconds())
	} else {
		defer resp.Body.Close()

		// 결과 출력
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		parseRes := parseHtml(string(data))

		t := time.Now()
		elapsed.ParseTime = float64(time.Since(t).Milliseconds())

		elapsed.Details = parseRes.timeElapsed

		spider.mutex.Lock()
		vectors := spider.oh.onehotBulk(parseRes.words)

		spider.lq.append(parseRes.links)

		spider.addLink(entry, vectors)
		spider.times = append(spider.times, *elapsed)
		spider.mutex.Unlock()
	}
}
