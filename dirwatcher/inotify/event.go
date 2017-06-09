// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// +build linux

package inotify

import (
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

type Flags uint32

// Inotify specific masks are legal, implemented events that are guaranteed to
// work with notify package on linux-based systems.
const (
	InAccess       = Flags(unix.IN_ACCESS)        // File was accessed
	InAttrib       = Flags(unix.IN_ATTRIB)        // Metadata changed
	InCloseWrite   = Flags(unix.IN_CLOSE_WRITE)   // Writtable file was closed
	InCloseNowrite = Flags(unix.IN_CLOSE_NOWRITE) // Unwrittable file closed
	InCreate       = Flags(unix.IN_CREATE)        // Subfile was created
	InDelete       = Flags(unix.IN_DELETE)        // Subfile was deleted
	InDeleteSelf   = Flags(unix.IN_DELETE_SELF)   // Self was deleted
	InModify       = Flags(unix.IN_MODIFY)        // File was modified
	InMoveSelf     = Flags(unix.IN_MOVE_SELF)     // Self was moved
	InMovedFrom    = Flags(unix.IN_MOVED_FROM)    // File was moved from X
	InMovedTo      = Flags(unix.IN_MOVED_TO)      // File was moved to Y
	InOpen         = Flags(unix.IN_OPEN)          // File was opened
)

// Inotify behavior events are not **currently** supported by notify package.
const (
	InDontFollow = Flags(unix.IN_DONT_FOLLOW)
	InExclUnlink = Flags(unix.IN_EXCL_UNLINK)
	InMaskAdd    = Flags(unix.IN_MASK_ADD)
	InOneshot    = Flags(unix.IN_ONESHOT)
	InOnlydir    = Flags(unix.IN_ONLYDIR)
)

const (
	inIgnored   = Flags(unix.IN_IGNORED)
	inIsDir     = Flags(unix.IN_ISDIR)
	inQOverflow = Flags(unix.IN_Q_OVERFLOW)
	inUnmount   = Flags(unix.IN_UNMOUNT)
)

var osestr = map[Flags]string{
	InAccess:       "notify.InAccess",
	InModify:       "notify.InModify",
	InAttrib:       "notify.InAttrib",
	InCloseWrite:   "notify.InCloseWrite",
	InCloseNowrite: "notify.InCloseNowrite",
	InOpen:         "notify.InOpen",
	InMovedFrom:    "notify.InMovedFrom",
	InMovedTo:      "notify.InMovedTo",
	InCreate:       "notify.InCreate",
	InDelete:       "notify.InDelete",
	InDeleteSelf:   "notify.InDeleteSelf",
	InMoveSelf:     "notify.InMoveSelf",
}

func (e Flags) String() string {
	var s []string
	for ev, str := range osestr {
		if e&ev == ev {
			s = append(s, str)
		}
	}
	return strings.Join(s, "|")
}

type Watcher struct {
	ifd   int
	efd   int
	pipe  []int
	paths map[string]int
	fds   map[int]string
	evts  chan<- Flags
	errs  chan<- error
	done  chan struct{}
}

func New(events chan<- Flags, errors chan<- error) (*Watcher, error) {
	var err error
	w := &Watcher{
		ifd:   -1,
		efd:   -1,
		pipe:  []int{-1, -1},
		paths: make(map[string]int),
		fds:   make(map[int]string),
		evts:  events, errs: errors,
		done: make(chan struct{}),
	}
	defer func() {
		if err != nil {
			w.Close()
		}
	}()
	w.ifd, err = unix.InotifyInit1(unix.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	err = unix.Pipe(w.pipe)
	if err != nil {
		return nil, err
	}

	w.efd, err = unix.EpollCreate(1)
	if err != nil {
		return nil, err
	}
	err = unix.EpollCtl(w.efd, unix.EPOLL_CTL_ADD, w.ifd, &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLERR})
	if err != nil {
		return nil, err
	}
	err = unix.EpollCtl(w.efd, unix.EPOLL_CTL_ADD, w.pipe[0], &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLERR})
	if err != nil {
		return nil, err
	}
	runtime.SetFinalizer(w, func(w *Watcher) {
		w.close()
	})
	go w.wait()
	return w, nil
}

func (w *Watcher) Add(path string, flags Flags) error {
	fd, err := unix.InotifyAddWatch(w.ifd, path, uint32(flags))
	if err != nil {
		return err
	}
	w.paths[path] = fd
	w.fds[fd] = path
	return nil
}

func (w *Watcher) Close() {
	if w.pipe[0] == -1 {
		return
	}
	unix.Close(w.pipe[1])
	w.pipe[0] = -1
	<-w.done
	w.close()
}

func (w *Watcher) close() {
	for fd, _ := range w.fds {
		unix.Close(fd)
	}
	for _, fd := range []int{w.ifd, w.efd, w.pipe[0], w.pipe[1]} {
		if fd != -1 {
			unix.Close(fd)
		}
	}
}

func (w *Watcher) wait() {
	events := make([]unix.EpollEvent, 2)
	for {
		n, err := unix.EpollWait(w.efd, events, -1)
		if err != nil {
			w.errs <- err
		}
		for i := 0; i < n; i++ {
			if events[i].Fd == int32(w.pipe[1]) {
				break
			} else if events[i].Fd == int32(w.ifd) {
				w.read()
			}
		}
	}
	close(w.done)
}

func (w *Watcher) read() {}
