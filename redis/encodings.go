package redis

type ObjectEncoding uint8

const (
	EncodingRaw ObjectEncoding = iota
	EncodingInt
	EncodingHashMap
	EncodingSetMap
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
