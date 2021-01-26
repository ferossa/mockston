package app

import (
	"bytes"
	"encoding/json"
)

// IMarshaller interface for marshalling/unmarshalling messages
type IMarshaller interface {
	Marshal(name string, msg interface{}) ([]byte, error)
	Unmarshal(name string, data []byte, msg interface{}) error
}

// JsonMarshaller default json marshaller
type JsonMarshaller struct {
}

// Marshal pack message to json format
func (m *JsonMarshaller) Marshal(name string, msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

// Unmarshal unpack message from json format
func (m *JsonMarshaller) Unmarshal(name string, data []byte, msg interface{}) error {
	r := bytes.NewReader(data)
	d := json.NewDecoder(r)
	d.UseNumber()

	return d.Decode(msg)
}
