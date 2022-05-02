package main

import "sort"

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
