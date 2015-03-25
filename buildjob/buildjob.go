package buildjob

import (
	"bytes"
	"encoding/gob"
)

// BuildJob represents a job for the gobuilder to build including the
// repository and a number of tries already made to build it
type BuildJob struct {
	Repository         string
	NumberOfExecutions int
}

// ToByte creats a gob encoded version of the BuildJob to store in text
// queues or other locations
func (b *BuildJob) ToByte() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBytes reads a []byte representation created by ToBytes and
// returns a reference to the BuildJob object for further usage
func FromBytes(b []byte) (*BuildJob, error) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var tmp BuildJob
	err := dec.Decode(&tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}
