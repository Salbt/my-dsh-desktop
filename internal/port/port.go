package port

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

func Free() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func WaitReady(url string, timeout time.Duration, logf func(string, ...any)) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	lastLog := time.Time{}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		if logf != nil && time.Since(lastLog) > 30*time.Second {
			lastLog = time.Now()
			logf("仍在等待服务启动...")
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("等待服务就绪超时: %s", url)
}
