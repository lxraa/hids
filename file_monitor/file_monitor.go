package file_monitor

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"
)

type FileEvent struct {
	Mask     uint32
	Cookie   uint32
	Filename string
	Events   []*Event
}

type Event struct {
	EventDec  string
	EventCode uint
}

var w int
var fd int

/**
	为文件file添加inotify监控，并返回fd和watched句柄用于关闭
**/
func addWatcherToFile(file string) (w int, fd int) {

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

func rmWatcherFromFile(fd int, w uint32) {
	syscall.InotifyRmWatch(fd, w)
}

/**
			注意，一个event可能会有多个时间，见linux文档:man inotify
 			struct inotify_event {
               int      wd;        Watch descriptor
               uint32_t mask;      Mask describing event
               uint32_t cookie;    Unique cookie associating related
                                     events (for rename(2))
               uint32_t len;       Size of name field
               char     name[];    Optional null-terminated name
           };
**/
func eventParser(pointer *syscall.InotifyEvent) *FileEvent {
	pFileEvent := new(FileEvent)
	// 保存cookie
	pFileEvent.Cookie = pointer.Cookie

	// 保存mask
	pFileEvent.Mask = pointer.Mask

	pFileEvent.Events = make([]*Event, 0, 1)

	// 指向name的初始位置
	pName := unsafe.Pointer(uintptr(unsafe.Pointer(pointer)) + syscall.SizeofInotifyEvent)
	// 中文？
	filenameBytes := make([]byte, pointer.Len)

	copy(filenameBytes, (*(*[syscall.PathMax]byte)(pName))[:pointer.Len])

	pFileEvent.Filename = strings.TrimRight(string(filenameBytes), "\x00")

	if (pointer.Mask & syscall.IN_ACCESS) > 0 {
		event := new(Event)
		event.EventCode = syscall.IN_ACCESS
		event.EventDec = "read"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_ATTRIB) > 0 {
		event := new(Event)
		event.EventCode = syscall.IN_ATTRIB
		event.EventDec = "change_attrbute"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_CLOSE_NOWRITE) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_CLOSE_NOWRITE
		event.EventDec = "close_no_write"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_CLOSE_WRITE) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_CLOSE_WRITE
		event.EventDec = "close_write"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_CREATE) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_CREATE
		event.EventDec = "create"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_DELETE) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_DELETE
		event.EventDec = "delete"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_DELETE_SELF) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_DELETE_SELF
		event.EventDec = "delete_self"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	// 触发该事件说明文件被删除后重新创建了，需要重新加
	if (pointer.Mask & syscall.IN_IGNORED) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_IGNORED
		event.EventDec = "ignored"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_ISDIR) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_ISDIR
		event.EventDec = "is_dir"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_MODIFY) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_MODIFY
		event.EventDec = "is_modify"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_MOVE_SELF) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_MOVE_SELF
		event.EventDec = "move_self"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_MOVED_FROM) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_MOVED_FROM
		event.EventDec = "moved_from"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_MOVED_TO) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_MOVED_TO
		event.EventDec = "moved_to"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_OPEN) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_OPEN
		event.EventDec = "open"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_Q_OVERFLOW) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_Q_OVERFLOW
		event.EventDec = "q_overflow"
		pFileEvent.Events = append(pFileEvent.Events, event)
	}
	if (pointer.Mask & syscall.IN_UNMOUNT) > 0 {

		event := new(Event)
		event.EventCode = syscall.IN_UNMOUNT
		event.EventDec = "unmount"
		pFileEvent.Events = append(pFileEvent.Events, event)
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

func Start(eventChannel chan interface{}) {
	filePath := "/home/liuxiaorui/test/test.txt"

	log.Println("application start.")

	// fd, err := unix.EpollCreate(10)
	// if err != nil {
	// 	unix.
	// }
	w, fd = addWatcherToFile(filePath)

	go func() {
		for {
			buffer := readBytesFromFd(fd)

			if len(buffer) < syscall.SizeofInotifyEvent {
				// No point trying if we don't have at least one event
				continue
			}

			fmt.Printf("Bytes read: %d\n", len(buffer))

			offset := 0
			//-syscall.SizeofInotifyEvent
			for offset < len(buffer) {
				// fmt.Printf("offset:%d\n", offset)
				event := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))
				fileEvent := eventParser(event)
				// log.Println(fileEvent.filename)
				// for k := 0; k < len(fileEvent.events); k++ {
				// 	fmt.Println(fileEvent.events[k].eventDec)
				// 	fmt.Println(fileEvent.filename)
				// }
				// fmt.Println("------------")
				// 输出到channel，由main统一处理
				eventChannel <- fileEvent
				if (fileEvent.Mask & syscall.IN_IGNORED) > 0 {
					log.Printf("重新绑定事件")
					rmWatcherFromFile(fd, uint32(w))
					w, fd = addWatcherToFile(filePath)
					break
				}

				// We need to account for the length of the name
				offset += syscall.SizeofInotifyEvent + int(event.Len)
			}
		}
	}()
}

func End() {
	log.Println("application end.")
	rmWatcherFromFile(fd, uint32(w))
	syscall.Close(fd)
}
