package global

import (
	"sync"
	"syscall"
)

/**
	频繁创建对象时，为避免反复申请内存+gc，使用sync.Pool缓存对象，控制内存使用
**/
var global *sync.Pool

func GetByteArray() *[]byte {
	b := global.Get().(*[]byte)
	*b = (*b)[0:0]
	return b
}
func PutByteArray(b *[]byte) {
	global.Put(b)
}

func init() {
	global = &sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, syscall.Getpagesize())
			return &b
		},
	}
}
