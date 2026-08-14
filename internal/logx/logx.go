package logx

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu sync.Mutex
	f  *os.File
}

func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &Logger{f: f}, nil
}

func (l *Logger) Printf(format string, args ...any) {
	entry := time.Now().Format("2006-01-02 15:04:05.000 ") + fmt.Sprintf(format, args...)
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Println(entry)
	if l.f != nil {
		_, _ = fmt.Fprintln(l.f, entry)
	}
}

type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(p []byte) (int, error) { return w(p) }

func Writer(logf func(string, ...any)) io.Writer {
	return writerFunc(func(p []byte) (int, error) {
		s := strings.TrimRight(string(p), "\r\n")
		for _, line := range strings.Split(s, "\n") {
			if strings.TrimSpace(line) != "" {
				logf("%s", line)
			}
		}
		return len(p), nil
	})
}
