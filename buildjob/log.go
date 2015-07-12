package buildjob

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"time"
)

// BuildLog is a metadata blob to hold state about a log
type BuildLog struct {
	ID      string
	Success bool
	Time    time.Time
}

// ToString creats a gob encoded version of the BuildJob to store in text
// queues or other locations
func (b *BuildLog) ToString() (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// LogFromString reads a string representation created by ToBytes and
// returns a reference to the BuildJob object for further usage
func LogFromString(s string) (*BuildLog, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var tmp BuildLog
	err = dec.Decode(&tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}
