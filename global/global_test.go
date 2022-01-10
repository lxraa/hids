package global

import (
	"fmt"
	"testing"
)

func TestGetByteArray(t *testing.T) {

	for i := 0; i < 1000; i++ {
		// time.Sleep(time.Duration(30) * time.Millisecond)
		byteArray := GetByteArray()
		fmt.Printf("%p\n", byteArray)
		for i := 0; i < 5000; i++ {
			*byteArray = append(*byteArray, '1')
		}

		PutByteArray(byteArray)

	}
}
