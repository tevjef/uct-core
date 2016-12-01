package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
)

func MarshalMessage(format string, m University) (*bytes.Reader, error) {
	var out []byte
	var err error
	if format == Json {
		out, err = json.Marshal(m)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode message")
		}
	} else if format == Protobuf {
		out, err = m.Marshal()
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode message")
		}
	}
	return bytes.NewReader(out), nil
}

func UnmarshalMessage(format string, r io.Reader, m *University) error {
	if format == Json {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(r, &*m); err != nil {
			return err
		}
	} else if format == Protobuf {
		data, err := ioutil.ReadAll(r)
		if err = m.Unmarshal(data); err != nil {
			return err
		}
	}
	if m.Equal(University{}) {
		return fmt.Errorf("%s Reason %s", "Failed to unmarshal message:", "empty data")
	}
	return nil
}
