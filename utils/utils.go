package utils

import (
	"log"
	"syscall"
)

const BATCH_SIZE = 4096

type Memlist struct {
	Len    int
	Buffer [BATCH_SIZE]byte
	next   *Memlist
}

/**
读取fd中所有bytes
**/
func ReadBytesFromFd(fd int) []byte {
	tmpBuffer := new(Memlist)
	start := tmpBuffer
	allLen := 0
	for {

		bytesRead, err := syscall.Read(fd, tmpBuffer.Buffer[:])
		allLen = allLen + bytesRead
		tmpBuffer.Len = bytesRead
		if err != nil {
			log.Fatal(err)
		}

		if bytesRead == BATCH_SIZE {
			tmpBuffer.next = new(Memlist)
			tmpBuffer = tmpBuffer.next
			continue
		}
		break
	}
	result := make([]byte, allLen)
	p := start
	offset := 0
	for p != nil {
		copy(result[offset:offset+p.Len], p.Buffer[:p.Len])
		offset = offset + p.Len
		p = p.next
	}

	return result
}
