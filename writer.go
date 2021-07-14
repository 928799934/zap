package zap

import (
	"os"
	"strconv"
	"time"
)

const (
	// ExtFormat 文件后缀名格式
	ExtFormat string = ".2006-01-02-15"
)

type writer struct {
	file     string
	ext      string
	lastFile string
	f        *os.File
}

func newWriter(file string) *writer {
	return &writer{
		file: file + "." + strconv.Itoa(os.Getpid()),
		ext:  ExtFormat,
	}
}

func (w *writer) Write(b []byte) (int, error) {
	nowFile := w.file + time.Now().Format(w.ext)
	if nowFile != w.lastFile {
		if w.f != nil {
			w.close()
		}
		f, err := os.OpenFile(nowFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
		w.f = f
		w.lastFile = nowFile
	}
	return w.f.Write(b)
}

func (w *writer) Sync() error {
	return w.f.Sync()
}

func (w *writer) close() {
	if w.f == nil {
		return
	}
	_ = w.f.Sync()
	_ = w.f.Close()
	w.f = nil
	w.lastFile = ""
	return
}
