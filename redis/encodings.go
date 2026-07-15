package redis

import "github.com/huandu/skiplist"

type ObjectEncoding uint8

const (
	EncodingRaw ObjectEncoding = iota
	EncodingInt
	EncodingHashMap
	EncodingSetMap
	EncodingSkipList
)

func (e ObjectEncoding) StrRep() string {
	switch e {
	case EncodingInt:
		return "EncodingInt"
	case EncodingRaw:
		return "EncodingRaw"
	case EncodingHashMap:
		return "EncodingHashMap"
	case EncodingSetMap:
		return "EncodingSetMap"

	case EncodingSkipList:
		return "EncodingSkipList"
	default:
		return "UNKNOWN"
	}
}

type objectPayload interface {
	objectPayload()
}

type rawStringPayload string
type intStringPayload int
type hashMapPayload map[string]string
type setMapPayload map[string]struct{}
type skipListPayload struct {
	scores map[string]int
	sl     *skiplist.SkipList
}

func (rawStringPayload) objectPayload() {}
func (intStringPayload) objectPayload() {}
func (hashMapPayload) objectPayload()   {}
func (setMapPayload) objectPayload()    {}
func (skipListPayload) objectPayload()  {}
