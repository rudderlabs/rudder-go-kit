package partmap

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
)

type (
	PartitionRangeStart uint32
	PartitionIndex      uint16
	NodeIndex           uint16
)

// PartitionMapping is map of partition number(the beginning of the partition range in the 32-bit space) to node index
type PartitionMapping map[PartitionRangeStart]NodeIndex

// PartitionIndexMapping is map of partition index to node index
type PartitionIndexMapping map[PartitionIndex]NodeIndex

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
	for partition, nodeIdx := range *pm {
		err := binary.Write(gzipWriter, binary.BigEndian, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to write partition range start: %w", err)
		}
		err = binary.Write(gzipWriter, binary.BigEndian, nodeIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to write node index: %w", err)
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
			nodeIdx   NodeIndex
		)
		if err := binary.Read(reader, binary.BigEndian, &partition); err != nil {
			return fmt.Errorf("reading int value: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &nodeIdx); err != nil {
			return fmt.Errorf("reading short value: %w", err)
		}
		(*pm)[partition] = nodeIdx
	}
	return nil
}

// ToPartitionIndexMapping converts PartitionMapping to map of partition index to node index based on the provided number of partitions
func (pm PartitionMapping) ToPartitionIndexMappingWithNumPartitions(numPartitions uint32) (PartitionIndexMapping, error) {
	if numPartitions == 0 {
		return nil, fmt.Errorf("numPartitions must be greater than 0")
	}
	if numPartitions == 1 {
		if len(pm) != 1 {
			return nil, fmt.Errorf("invalid partition mapping: expected exactly one partition range for numPartitions=1, got %d", len(pm))
		}
		return PartitionIndexMapping{0: pm[0]}, nil // Return a single partition index mapping if numPartitions is 1
	}
	window := uint32((1 << 32) / int(numPartitions))
	result := make(PartitionIndexMapping)

	for rangeStart, nodeIdx := range pm {
		if mod := uint32(rangeStart) % window; mod != 0 {
			return nil, fmt.Errorf("invalid partition mapping: partition range start %d is not aligned with window size %d", rangeStart, window)
		}
		idx := rangeStart / PartitionRangeStart(window)
		result[PartitionIndex(idx)] = nodeIdx
	}
	return result, nil
}

// ToPartitionIndexMapping converts PartitionMapping to map of partition index to node index, inferring the number of partitions from the length of the mapping
func (pm PartitionMapping) ToPartitionIndexMapping() (PartitionIndexMapping, error) {
	return pm.ToPartitionIndexMappingWithNumPartitions(uint32(len(pm)))
}

// ToPartitionMapping converts PartitionIndexMapping to map of partition range start to node index based on the provided number of partitions
func (pim PartitionIndexMapping) ToPartitionMappingWithNumPartitions(numPartitions uint32) (PartitionMapping, error) {
	if numPartitions == 0 {
		return nil, fmt.Errorf("numPartitions must be greater than 0")
	}
	if numPartitions == 1 {
		if len(pim) != 1 {
			return nil, fmt.Errorf("invalid partition index mapping: expected exactly one partition for numPartitions=1, got %d", len(pim))
		}
		return PartitionMapping{0: pim[0]}, nil // Return a single partition mapping if numPartitions is 1
	}
	window := uint32((1 << 32) / int(numPartitions))
	result := make(PartitionMapping, len(pim))
	for partitionIdx, nodeIdx := range pim {
		rangeStart := uint32(partitionIdx) * window
		result[PartitionRangeStart(rangeStart)] = nodeIdx
	}
	return result, nil
}

// ToPartitionMapping converts PartitionIndexMapping to map of partition range start to node index, inferring the number of partitions from the length of the mapping
func (pim PartitionIndexMapping) ToPartitionMapping() (PartitionMapping, error) {
	return pim.ToPartitionMappingWithNumPartitions(uint32(len(pim)))
}

// ToPartitionRangeStart converts partition index to partition range start based on the provided number of partitions
func (partitionIdx PartitionIndex) ToPartitionRangeStart(numPartitions uint32) (PartitionRangeStart, error) {
	if numPartitions == 0 {
		return 0, fmt.Errorf("numPartitions must be greater than 0")
	}
	if numPartitions == 1 {
		return 0, nil // Return partition range start 0 for any partition index if numPartitions is 1
	}
	if partitionIdx >= PartitionIndex(numPartitions) {
		return 0, fmt.Errorf("partition index %d out of range for numPartitions %d", partitionIdx, numPartitions)
	}
	window := uint32((1 << 32) / int(numPartitions))
	rangeStart := uint32(partitionIdx) * window
	return PartitionRangeStart(rangeStart), nil
}

// ToPartitionIndex converts partition range start to partition index based on the provided number of partitions
func (rangeStart PartitionRangeStart) ToPartitionIndex(numPartitions uint32) (PartitionIndex, error) {
	if numPartitions == 0 {
		return 0, fmt.Errorf("numPartitions must be greater than 0")
	}
	if numPartitions == 1 {
		return 0, nil // Return partition index 0 for any range start if numPartitions is 1
	}
	window := uint32((1 << 32) / int(numPartitions))
	idx := rangeStart / PartitionRangeStart(window)
	return PartitionIndex(idx), nil
}
