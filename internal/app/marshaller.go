package app

// IMarshaller interface for marshalling/unmarshalling messages
type IMarshaller interface {
	Marshal(msg interface{}) ([]byte, error)
	Unmarshal(data []byte, msg interface{}) error
}
