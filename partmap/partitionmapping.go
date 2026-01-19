package partmap

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
)

type (
	PartitionRangeStart uint32
	ServerIndex         uint16
)

// PartitionMapping is map of partition number(the beginning of the partition range in the 32-bit space) to server index
type PartitionMapping map[PartitionRangeStart]ServerIndex

// Marshal marshalls PartitionMapping to raw bytes
func (pm *PartitionMapping) Marshal() ([]byte, error) {
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	listLength := uint32(len(*pm))
	err := binary.Write(gzipWriter, binary.BigEndian, listLength)
	if err != nil {
		return nil, fmt.Errorf("failed to write list length: %w", err)
	}
	// Write each pair of integers to the binary data buffer
	for partition, serverIdx := range *pm {
		err := binary.Write(gzipWriter, binary.BigEndian, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to write partition range start: %w", err)
		}
		err = binary.Write(gzipWriter, binary.BigEndian, serverIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to write server index: %w", err)
		}
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return compressedBuffer.Bytes(), nil
}

// Unmarshal unmarshalls raw bytes to PartitionMapping
func (pm *PartitionMapping) Unmarshal(data []byte) error {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()
	var listLength uint32
	err = binary.Read(reader, binary.BigEndian, &listLength)
	if err != nil {
		return fmt.Errorf("reading list length: %w", err)
	}
	*pm = make(PartitionMapping)
	for i := uint32(0); i < listLength; i++ {
		var (
			partition PartitionRangeStart
			serverIdx ServerIndex
		)
		if err := binary.Read(reader, binary.BigEndian, &partition); err != nil {
			return fmt.Errorf("reading int value: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &serverIdx); err != nil {
			return fmt.Errorf("reading short value: %w", err)
		}
		(*pm)[partition] = serverIdx
	}
	return nil
}
