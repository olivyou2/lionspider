package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"time"
)

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
