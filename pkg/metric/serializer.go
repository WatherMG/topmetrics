package metric

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
)

const (
	JSONType  = 0x1
	GOBType   = 0x2
	ProtoType = 0x3
)

type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type ProtoSerializer interface {
	Marshal(v proto.Message) ([]byte, error)
	Unmarshal(data []byte, v proto.Message) error
}

type JSONSerialize struct{}

func (s *JSONSerialize) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerialize) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type GOBSerialize struct{}

func (s *GOBSerialize) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (s *GOBSerialize) Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(v)
}

type ProtoSerialize struct{}

func (s *ProtoSerialize) Marshal(v proto.Message) ([]byte, error) {
	return proto.Marshal(v)
}

func (s *ProtoSerialize) Unmarshal(data []byte, v proto.Message) error {
	return proto.Unmarshal(data, v)
}

func Marshal(serializer interface{}, metric interface{}) ([]byte, error) {
	switch s := serializer.(type) {
	case *JSONSerialize:
		return s.Marshal(metric)
	case *GOBSerialize:
		return s.Marshal(metric)
	case *ProtoSerialize:
		if msg, ok := metric.(proto.Message); ok {
			return s.Marshal(msg)
		}
		return nil, fmt.Errorf("unsupported value type for protobuf serialization: %T", metric)
	default:
		return nil, fmt.Errorf("unsupported serializer type: %T", s)
	}
}

func Unmarshal(serializer interface{}, data []byte, metric interface{}) error {
	switch s := serializer.(type) {
	case *JSONSerialize:
		return s.Unmarshal(data, metric)
	case *GOBSerialize:
		return s.Unmarshal(data, metric)
	case *ProtoSerialize:
		if msg, ok := metric.(proto.Message); ok {
			return s.Unmarshal(data, msg)
		}
		return fmt.Errorf("unsupported value type for protobuf deserialization: %T", metric)
	default:
		return fmt.Errorf("unsupported deserializer type: %T", s)
	}
}
