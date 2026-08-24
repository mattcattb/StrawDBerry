package redis

type ObjectEncoding uint8

const (
	ObjectEncodingStringRaw ObjectEncoding = iota
	ObjectEncodingStringInt
	ObjectEncodingHashMap
	ObjectEncodingSetMap
	ObjectEncodingZSetSkiplist
)

func (e ObjectEncoding) String() string {
	switch e {
	case ObjectEncodingStringInt:
		return "int"
	case ObjectEncodingStringRaw:
		return "raw"
	case ObjectEncodingHashMap:
		return "hashtable"
	case ObjectEncodingSetMap:
		return "hashtable"

	case ObjectEncodingZSetSkiplist:
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
