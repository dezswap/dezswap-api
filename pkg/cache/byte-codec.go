package cache

import (
	"bytes"
	"encoding/gob"
)

var _ Codable = &byteCodecImpl{}

type byteCodecImpl struct{}

func NewByteCodec() Codable {
	return &byteCodecImpl{}
}

// Decode implements Encoder.
func (c *byteCodecImpl) Decode(val []byte, dest interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(val))
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	return nil
}

// Encode implements Encoder.
func (c *byteCodecImpl) Encode(value interface{}) ([]byte, error) {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
