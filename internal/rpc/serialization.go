package rpc

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

func serialize(i interface{}) ([]byte, error) {
	js, err := json.Marshal(i)
	if err != nil {
		return js, fmt.Errorf("error marshaling json: %s", err)
	}

	bb := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(bb)
	_, err = gz.Write(js)
	if err != nil {
		return nil, fmt.Errorf("error writing gzip: %s", err)
	}
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %s", err)
	}

	return bb.Bytes(), nil
}

func deserialize(msg []byte, i interface{}) error {
	b := bytes.NewReader(msg)
	zr, err := gzip.NewReader(b)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %s", err)
	}
	defer zr.Close()

	js, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("error reading gzip: %s", err)
	}

	return json.Unmarshal(js, i)
}
