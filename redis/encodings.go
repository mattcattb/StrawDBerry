package redis

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
		return "int"
	case EncodingRaw:
		return "raw"
	case EncodingHashMap:
		return "hashtable"
	case EncodingSetMap:
		return "hashtable"

	case EncodingSkipList:
		return "skiplist"
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

func (rawStringPayload) objectPayload() {}
func (intStringPayload) objectPayload() {}
func (hashMapPayload) objectPayload()   {}
func (setMapPayload) objectPayload()    {}
func (*zset) objectPayload()            {}
