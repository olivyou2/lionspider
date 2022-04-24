package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

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

func startsWith(src string, target string) bool {
	if len(src) < len(target) {
		return false
	}
	if src[0:len(target)] == target {
		return true
	} else {
		return false
	}
}

func endsWith(src string, target string) bool {

	if len(src) < len(target) {
		return false
	}
	if src[len(src)-len(target):] == target {
		return true
	} else {
		return false
	}
}

type alertStruct struct {
	Src    string
	Origin string
	Words  []string
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

type linkStruct struct {
	link     string
	relation int
}

type linkQueue struct {
	link        []linkStruct
	linkHistory map[string]bool
	cp          *crawlPolicy
}

func newLinkQueue(cp *crawlPolicy) linkQueue {
	lq := linkQueue{
		link:        []linkStruct{},
		linkHistory: map[string]bool{},
		cp:          cp,
	}

	return lq
}

func (queue *linkQueue) getTop() string {
	sort.Slice(queue.link, func(i, j int) bool {
		return queue.link[i].relation > queue.link[j].relation
	})

	if len(queue.link) > 0 {
		top := queue.link[0]

		queue.link = queue.link[1:len(queue.link)]
		return top.link
	} else {
		return ""
	}
}

func (queue *linkQueue) append(links []string) {
	ls := []linkStruct{}

	for _, link := range links {
		_, ok := queue.linkHistory[link]
		if !ok {
			queue.linkHistory[link] = true
			relation := queue.cp.getRelation(link)

			ls = append(ls, linkStruct{
				link:     link,
				relation: relation,
			})
		}
	}

	queue.link = append(queue.link, ls...)
}

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

type timeElapsed struct {
	GetTime   float64
	ParseTime float64
	EmbedTime float64
	Link      string
	Details   map[string]int64
}

type searchResult struct {
	ok    bool
	links []linkRelation

	availableWords []string
}

type linkRelation struct {
	Link     string
	Relation int
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

func checkHtml(src string) bool {
	headResp, headErr := http.Head(src)

	if headErr != nil {
		return false
	}

	if len(headResp.Header["Content-Type"]) == 0 {

		return false
	}

	isHtml := startsWith(headResp.Header["Content-Type"][0], "text/html")
	return isHtml
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

func main() {
	start := time.Now()

	spid := newSpider()
	spid.maxLinks = 100
	spid.lq.cp.addWord("blog", "naver", "rainyrani119")

	spid.lq.append([]string{"https://blog.naver.com/PostView.naver?blogId=rainyrani119&logNo=222677366898&parentCategoryNo=&categoryNo=&viewDate=&isShowPopularPosts=false&from="})
	//spid.lq.append([]string{"http://www.naver.com"})

	spid.crawlProcessor(12)
	// Crawling Web

	fmt.Println("parsing duration:", time.Since(start))

	infos := spid.getSpiderInfo()
	fmt.Printf("%+v\n", infos)

	sort.Slice(spid.times, func(i, j int) bool {
		//iItem := spid.times[i]
		//jItem := spid.times[j]

		iElapsed := spid.times[i].EmbedTime + spid.times[i].GetTime + spid.times[i].ParseTime
		jElapsed := spid.times[j].EmbedTime + spid.times[j].GetTime + spid.times[j].ParseTime

		return iElapsed > jElapsed
	})

	marshalRes, _ := json.Marshal(spid.times)
	err := ioutil.WriteFile("./elapsed.json", marshalRes, 0644)

	if err != nil {
		panic(err)
	}

	res := spid.search("두께")

	searchRes, _ := json.Marshal(res.links)
	ioutil.WriteFile("./searchResult.json", searchRes, 0644)

	fmt.Println("Search Done!", len(res.links), "Links")

}
