package util

import (
	"io"
	"net"
)

func CopyBetween(conn1, conn2 net.Conn) {
	done := make(chan bool, 2)

	go func() {
		io.Copy(conn2, conn1)
		done <- true
		conn1.Close()
		conn2.Close()
	}()

	go func() {
		io.Copy(conn1, conn2)
		done <- true
		conn1.Close()
		conn2.Close()
	}()

	<-done
}
