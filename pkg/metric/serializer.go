package metric

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
)

const (
	JSONType = iota + 0x1
	GOBType
	ProtoType
)

// Serializer is an interface for marshaling and unmarshaling data
type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	String() string
}

// ProtoSerializer is an interface for marshaling and unmarshaling protobuf data
type ProtoSerializer interface {
	Marshal(v proto.Message) ([]byte, error)
	Unmarshal(data []byte, v proto.Message) error
	String() string
}

type JSONSerialize struct{}

func (s *JSONSerialize) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerialize) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// String returns the name of the serializer
func (s *JSONSerialize) String() string {
	return "JSON"
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

func (s *GOBSerialize) String() string {
	return "GOB"
}

type ProtoSerialize struct{}

func (s *ProtoSerialize) Marshal(v proto.Message) ([]byte, error) {
	return proto.Marshal(v)
}

func (s *ProtoSerialize) Unmarshal(data []byte, v proto.Message) error {
	return proto.Unmarshal(data, v)
}

func (s *ProtoSerialize) String() string {
	return "Proto"
}

// Marshal marshals the given metric using the given serializer
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
		return nil, fmt.Errorf("marshal: unsupported value type for protobuf serialization: %T", metric)
	default:
		return nil, fmt.Errorf("marshal: unsupported serializer type: %T", s)
	}
}

// Unmarshal unmarshals the given data into the given metric using the given serializer
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
		return fmt.Errorf("unmarshal: unsupported value type for protobuf deserialization: %T", metric)
	default:
		return fmt.Errorf("unmarshal: unsupported deserializer type: %T", s)
	}
}

var dataTypeMap = map[byte]interface{}{
	JSONType:  &JSONSerialize{},
	GOBType:   &GOBSerialize{},
	ProtoType: &ProtoSerialize{},
}

// NewSerializer create serializer depending on dataType
func NewSerializer(dataType byte) (interface{}, error) {
	if s, ok := dataTypeMap[dataType]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("new serializer: unknown data type: %v", dataType)
}
