package main

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"
)

type FileEvent struct {
	mask      uint32
	cookie    uint32
	filename  string
	event     string
	eventCode uint
}

/**
	为文件file添加inotify监控，并返回fd和watched句柄用于关闭
**/
func AddWatcherToFile(file string) (w int, fd int) {

	fd, e := syscall.InotifyInit()
	if e != nil {
		log.Fatal("初始化inotify失败")
	}

	watched, e := syscall.InotifyAddWatch(fd, file, syscall.IN_ALL_EVENTS)

	if e != nil {
		_ = syscall.Close(fd)
		log.Fatal("创建监控失败")
	}
	return watched, fd
}

// func RmWatcherFromFile(fd int, w uint32) {
// 	syscall.InotifyRmWatch()
// }

func EventParser(pointer *syscall.InotifyEvent) *FileEvent {
	pFileEvent := new(FileEvent)
	eventNums := 0
	// 保存cookie
	pFileEvent.cookie = pointer.Cookie

	// 保存mask
	pFileEvent.mask = pointer.Mask
	// 指向name的初始位置
	pName := unsafe.Pointer(uintptr(unsafe.Pointer(pointer)) + syscall.SizeofInotifyEvent)
	// 中文？
	filenameBytes := make([]byte, pointer.Len)

	copy(filenameBytes, (*(*[syscall.PathMax]byte)(pName))[:pointer.Len])

	pFileEvent.filename = strings.TrimRight(string(filenameBytes), "\x00")

	if (pointer.Mask & syscall.IN_ACCESS) > 0 {
		eventNums++
		pFileEvent.event = "read"
		pFileEvent.eventCode = syscall.IN_ACCESS
	}
	if (pointer.Mask & syscall.IN_ATTRIB) > 0 {
		eventNums++
		pFileEvent.event = "change_attrbute"
		pFileEvent.eventCode = syscall.IN_ATTRIB
	}
	if (pointer.Mask & syscall.IN_CLOSE_NOWRITE) > 0 {
		eventNums++
		pFileEvent.event = "close_no_write"
		pFileEvent.eventCode = syscall.IN_CLOSE_NOWRITE
	}
	if (pointer.Mask & syscall.IN_CLOSE_WRITE) > 0 {
		eventNums++
		pFileEvent.event = "close_write"
		pFileEvent.eventCode = syscall.IN_CLOSE_WRITE
	}
	if (pointer.Mask & syscall.IN_CREATE) > 0 {
		eventNums++
		pFileEvent.event = "create"
		pFileEvent.eventCode = syscall.IN_CREATE
	}
	if (pointer.Mask & syscall.IN_DELETE) > 0 {
		eventNums++
		pFileEvent.event = "delete"
		pFileEvent.eventCode = syscall.IN_DELETE
	}
	if (pointer.Mask & syscall.IN_DELETE_SELF) > 0 {
		eventNums++
		pFileEvent.event = "delete_self"
		pFileEvent.eventCode = syscall.IN_DELETE_SELF
	}
	// 触发该事件说明文件被删除后重新创建了，需要重新加
	if (pointer.Mask & syscall.IN_IGNORED) > 0 {
		eventNums++
		pFileEvent.event = "ignored"
		pFileEvent.eventCode = syscall.IN_IGNORED
	}
	if (pointer.Mask & syscall.IN_ISDIR) > 0 {
		eventNums++
		pFileEvent.event = "is_dir"
		pFileEvent.eventCode = syscall.IN_ISDIR
	}
	if (pointer.Mask & syscall.IN_MODIFY) > 0 {
		eventNums++
		pFileEvent.event = "is_modify"
		pFileEvent.eventCode = syscall.IN_MODIFY
	}
	if (pointer.Mask & syscall.IN_MOVE_SELF) > 0 {
		eventNums++
		pFileEvent.event = "move_self"
		pFileEvent.eventCode = syscall.IN_MOVE_SELF
	}
	if (pointer.Mask & syscall.IN_MOVED_FROM) > 0 {
		eventNums++
		pFileEvent.event = "moved_from"
		pFileEvent.eventCode = syscall.IN_MOVED_FROM
	}
	if (pointer.Mask & syscall.IN_MOVED_TO) > 0 {
		eventNums++
		pFileEvent.event = "moved_to"
		pFileEvent.eventCode = syscall.IN_MOVED_TO
	}
	if (pointer.Mask & syscall.IN_OPEN) > 0 {
		eventNums++
		pFileEvent.event = "open"
		pFileEvent.eventCode = syscall.IN_OPEN
	}
	if (pointer.Mask & syscall.IN_Q_OVERFLOW) > 0 {
		eventNums++
		pFileEvent.event = "q_overflow"
		pFileEvent.eventCode = syscall.IN_Q_OVERFLOW
	}
	if (pointer.Mask & syscall.IN_UNMOUNT) > 0 {
		eventNums++
		pFileEvent.event = "unmount"
		pFileEvent.eventCode = syscall.IN_UNMOUNT
	}
	if eventNums > 1 {
		log.Fatal("一次event出现了多个事件，需修改代码")
	}

	return pFileEvent
}

const BATCH_SIZE = 2048

type Memlist struct {
	Len    int
	Buffer [BATCH_SIZE]byte
	next   *Memlist
}

/**
读取fd中所有bytes
**/
func readBytesFromFd(fd int) []byte {

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

func main() {
	log.Println("application starting....")
	defer log.Println("application end.")

	// fd, err := unix.EpollCreate(10)
	// if err != nil {
	// 	unix.
	// }
	w, fd := AddWatcherToFile("/home/liuxiaorui/test/")
	defer func() {
		syscall.InotifyRmWatch(fd, uint32(w))
		syscall.Close(fd)
	}()

	// fileEvents := make(chan FileChangeEvent)

	for {
		// Room for at least 128 events
		buffer := readBytesFromFd(fd)

		if len(buffer) < syscall.SizeofInotifyEvent {
			// No point trying if we don't have at least one event
			continue
		}

		fmt.Printf("Bytes read: %d\n", len(buffer))

		offset := 0
		//-syscall.SizeofInotifyEvent
		for offset < len(buffer) {
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))
			fileEvent := EventParser(event)
			// log.Println(fileEvent.filename)
			log.Println(fileEvent.event)
			if fileEvent.event == "" {
				log.Println(fileEvent.mask)
			}
			// We need to account for the length of the name
			offset += syscall.SizeofInotifyEvent + int(event.Len)
		}
	}

	// //1、创建epoll
	// watcher, err := fsnotify.NewWatcher()

	// if err != nil {
	// 	log.Fatal("new watcher failed", err)
	// }
	// defer watcher.Close()
	// //2、创建文件监控句柄
	// fileMoniterChannel := make(chan FileChangeEvent)

	// go func() {
	// 	defer close(fileMoniterChannel)

	// 	for {
	// 		select {
	// 		case event, ok := <-watcher.Events:
	// 			if !ok {
	// 				return
	// 			}
	// 			// 接收安全事件
	// 			log.Printf("%s %s\n", event.Name, event.Op)
	// 		case err, ok := <-watcher.Errors:
	// 			if !ok {
	// 				return
	// 			}
	// 			// 接收错误
	// 			log.Println("err:", err)
	// 		}
	// 	}
	// }()
	// // 3、添加文件监控
	// err = watcher.Add("/etc/shadow")
	// if err != nil {
	// 	log.Fatal("添加/etc/shadow监控失败", err)
	// }
	// err = watcher.Add("/etc/passwd")
	// if err != nil {
	// 	log.Fatal("添加/etc/passwd监控失败", err)
	// }
	// err = watcher.Add("/etc/cron.d")
	// if err != nil {
	// 	log.Fatal("添加/etc/cron.d失败", err)
	// }
	// err = watcher.Add("/etc/crontab")
	// if err != nil {
	// 	log.Fatal("添加/etc/cron.d失败", err)
	// }
	// err = watcher.Add("/var/spool/cron/")
	// if err != nil {
	// 	log.Fatal("添加/etc/cron.d失败", err)
	// }
	// err = watcher.Add("/var/spool/cron/crontabs/")
	// if err != nil {
	// 	log.Fatal("添加/etc/cron.d失败", err)
	// }

	// for {
	// 	fileChangeEvent := <-fileMoniterChannel
	// 	fmt.Println(fileChangeEvent.filePath)
	// 	fmt.Println(fileChangeEvent.event)
	// }

}
