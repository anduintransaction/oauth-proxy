package utils

import (
	"runtime"
	"testing"
	"time"
)

func TestOs(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	cmd, done, err := Fork("sleep", "1000")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		err = Kill(cmd)
		if err != nil {
			t.Fatal(err)
		}
	}()
	ticker := time.NewTicker(time.Second)
	select {
	case <-done:
		return
	case <-ticker.C:
		t.Fatal("timeout")
	}
}
