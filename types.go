package main

type linkStruct struct {
	link     string
	relation int
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
