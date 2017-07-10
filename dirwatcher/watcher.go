package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	ErrNotStarted     = fmt.Errorf("Watcher not started")
	ErrAlreadyStarted = fmt.Errorf("Watcher already started")
)

type OverflowError struct {
	ev Event
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("Event discarded: %v", e.ev)
}

// Event represents a single file system notification.
type Event struct {
	Name string // Relative path to the file or directory.
	Op   Op     // File operation that triggered the event.
}

// Op describes a set of file operations.
type Op uint32

// These are the generalized file operations that can trigger a notification.
const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

type Watcher struct {
	events chan<- Event
	errors chan<- error
	w      *fsnotify.Watcher
	wg     sync.WaitGroup
	dir    string
}

func New(dir string, events chan<- Event, errors chan<- error) *Watcher {
	return &Watcher{dir: dir, events: events, errors: errors}
}

func (w *Watcher) Start() error {
	var err error
	if w.w != nil {
		return ErrAlreadyStarted
	}
	w.w, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.wg.Add(1)
	go w.monitor()
	return w.walkdir(w.dir)
}

func (w *Watcher) Stop() error {
	if w.w == nil {
		return ErrNotStarted
	}
	err := w.w.Close()
	w.wg.Wait()
	return err
}

func (w *Watcher) walkdir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	werr := w.w.Add(dir)
	if werr != nil {
		return werr
	}
	for {
		fis, err := d.Readdir(100)
		for _, fi := range fis {
			if !fi.IsDir() {
				continue
			}
			dirpath := filepath.Join(dir, fi.Name())
			err := w.walkdir(dirpath)
			if err != nil {
				return err
			}
		}
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			return nil
		}
	}
}

func (w *Watcher) monitor() {
	defer w.wg.Done()
	for {
		select {
		case ev, ok := <-w.w.Events:
			if !ok {
				return
			}
			w.sendEvent(ev)
		case err, ok := <-w.w.Errors:
			if !ok {
				return
			}
			w.sendError(err)
		}
	}
}

func (w *Watcher) sendEvent(ev fsnotify.Event) {
	e := Event{Name: ev.Name, Op: Op(ev.Op)}
	select {
	case w.events <- e:
	default:
		w.sendError(OverflowError{e})
	}
}

func (w *Watcher) sendError(err error) {
	select {
	case w.errors <- err:
	default:
	}
}

func main() {
	done := make(chan struct{})
	events := make(chan Event)
	errors := make(chan error)
	go func(events <-chan Event, errors <-chan error) {
		defer close(done)
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return
				}
				fmt.Println(ev)
			case err, ok := <-errors:
				if !ok {
					return
				}
				fmt.Println(err)
			}
		}
	}(events, errors)
	w := New(os.Args[1], events, errors)
	err := w.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	<-done
}
