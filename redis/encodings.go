package redis

type ObjectEncoding uint8

const (
	EncodingRaw ObjectEncoding = iota
	EncodingInt
	EncodingHashMap
	EncodingSetMap
)

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
