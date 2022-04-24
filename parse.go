package main

import (
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type parseHtmlResult struct {
	words []string
	links []string

	timeElapsed map[string]int64
}

func parseHtml(src string) parseHtmlResult {
	res := parseHtmlResult{
		timeElapsed: map[string]int64{},
	}

	regTime := time.Now()

	srcReg := regexp.MustCompile("\"http([^\"]*)\"")

	/*
		styleReg := regexp.MustCompile("<style((.|\n)*)>((.|\n)*)</style>")
		tinReg := regexp.MustCompile("<!--(.*)-->")
		scriptReg := regexp.MustCompile("<script((.|\n)*?)>((.|\n)*?)</script>")
		tagReg := regexp.MustCompile("<([^>]*)>")
	*/
	spaceReg := regexp.MustCompile(`(\s{1,})`)

	links := srcReg.FindAllString(src, -1)
	for i := 0; i < len(links); i += 1 {
		links[i] = links[i][1 : len(links[i])-1]
	}

	//origin := src

	/*
		src = styleReg.ReplaceAllString(src, "")
		src = tinReg.ReplaceAllString(src, "")

		//scripts := scriptReg.FindAllString(src, -1)
		///for i, script := range scripts {
		//	fmt.Println(i, script)
		//}

		src = scriptReg.ReplaceAllLiteralString(src, "")
		src = tagReg.ReplaceAllString(src, "")
	*/

	parsed := ""

	dom := html.NewTokenizer(strings.NewReader(src))
	previousStartTokenTest := dom.Token()

loop:
	for {
		tt := dom.Next()
		switch {
		case tt == html.ErrorToken:
			break loop // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = dom.Token()
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" {
				continue
			}

			if previousStartTokenTest.Data == "style" {
				continue
			}

			parsed += strings.TrimSpace(html.UnescapeString(string(dom.Text())))
			//if len(TxtContent) > 0 {
			//		fmt.Printf("%s\n", TxtContent)
			//	}
		}
	}

	src = spaceReg.ReplaceAllString(parsed, " ")

	josaReg := regexp.MustCompile("[은는을를의에이가할된]")
	notJosaWords := []string{}

	words := strings.Split(src, " ")

	for _, word := range words {
		replaced := josaReg.ReplaceAllString(word, "")

		if replaced != word {
			notJosaWords = append(notJosaWords, replaced)
		}
	}

	words = append(words, notJosaWords...)

	regDuration := time.Since(regTime).Milliseconds()

	//fmt.Println("RegDuration :", regDuration)

	res.timeElapsed["regDuration"] = regDuration

	regTime = time.Now()

	//splitDuration := time.Since(regTime).Milliseconds()
	//fmt.Println("Split Elapsed :", splitDuration, "Words :", len(words))

	/*
		if len(words) < 100 {
			as := alertStruct{
				Words: words,
				Src:   src,
			}

			wordsJson, _ := JSONMarshal(as)
			ioutil.WriteFile("./Alert/"+fmt.Sprint(len(words))+".json", wordsJson, 0644)
			ioutil.WriteFile("./Page/"+fmt.Sprint(len(words))+".html", []byte(origin), 0644)
		}'
	*/

	res.words = words
	res.links = links
	return res
}
