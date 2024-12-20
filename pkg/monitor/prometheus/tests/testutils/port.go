package testutils

import (
	"net"
	"testing"
)

// GetFreePort 获取一个空闲的端口号
func GetFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("无法解析TCP地址: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("无法监听TCP端口: %v", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}
