package raft

import "github.com/Xynnn007/simple-raft/misc"

// Because after searching for a term-index pair, we often delete the elements after the target one
// vector, or the slice will be the best one

type Entries struct {
	Storage []*Entry
}

type Entry struct {
	Index int
	Term  int
	Content
}

type Content interface {
	GetContent() []byte
	Apply() bool
	SetContent([]byte)
}

func (es *Entries) Insert(e *Entry) {
	es.Storage = append(es.Storage, e)
}

// -1 no
// -2 conflict
// >=0 index
func (es *Entries) Find(term int, index int) int {
	if len(es.Storage) == 0 {
		return -1
	}

	pass := es.Storage[0].Index

	if index-pass >= len(es.Storage) {
		return -1
	}

	if es.Storage[index-pass].Term == term {
		return index - pass
	}

	return -1
}

func (es *Entries) FindTerm(index int) int {
	if len(es.Storage) == 0 {
		return -1
	}

	pass := es.Storage[0].Index

	if index-pass >= len(es.Storage) {
		return -1
	}

	return es.Storage[index-pass].Term
}

func (es *Entries) GetRange(st int, length int) []Entry {
	var res []Entry

	if len(es.Storage) == 0 {
		return res
	}

	pass := es.Storage[0].Index

	if st-pass >= len(es.Storage) {
		return res
	}

	for i := st - pass; i <= misc.Min(st-pass+length-1, len(es.Storage)-1); i++ {
		res = append(res, *es.Storage[i])
	}

	return res
}

// delete all elements after index
func (es *Entries) DeleteFrom(index int) {
	es.Storage = es.Storage[:index+1]
}
