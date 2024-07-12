package hubur

import (
	"bytes"
	"errors"
	"net"
	"time"
)

type TimeoutConn struct {
	net.Conn
	timeout           time.Duration
	allowPartialWrite bool
}

func (t *TimeoutConn) AllowPartialWrite(b bool) {
	t.allowPartialWrite = b
}

func (t *TimeoutConn) Read(b []byte) (n int, err error) {
	return t.ReadTimeout(b, t.timeout)
}

func (t *TimeoutConn) ReadTimeout(b []byte, timeout time.Duration) (n int, err error) {
	if err := t.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	return t.Conn.Read(b)
}

func (t *TimeoutConn) Write(b []byte) (n int, err error) {
	return t.WriteTimeout(b, t.timeout)
}

func (t *TimeoutConn) WriteTimeout(b []byte, timeout time.Duration) (int, error) {
	cnt := 0
	for {
		if err := t.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return cnt, err
		}
		i := cnt - 1
		if i < 0 {
			i = 0
		}
		n, err := t.Conn.Write(b[i:])
		if err != nil {
			return cnt, err
		}
		cnt += n
		if cnt == len(b) || t.allowPartialWrite {
			return cnt, nil
		}
	}
}

// 读取数据直到 until 变量，如果没有找到第二个参数就返回 error
func (t *TimeoutConn) ReadUntil(until []byte, maxReadSize int) ([]byte, error) {
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 1)
	for i := 1; i <= maxReadSize; i++ {
		n, err := t.Read(tmp)
		if n == 1 {
			buf = append(buf, tmp[0])
		}
		if bytes.HasSuffix(buf, until) {
			return buf[:i], nil
		}
		if err != nil {
			return buf[:i], err
		}
	}
	return buf, errors.New("max read size exceeded")
}

// ReadAll 将缓冲区所有数据读完并返回, 相当于 conn 版的 ioutil.ReadAll
func (t *TimeoutConn) ReadAll() ([]byte, error) {
	var ret []byte
	for {
		m := make([]byte, 2048)
		n, err := t.ReadTimeout(m, t.timeout)
		if err != nil {
			if len(ret) == 0 {
				return nil, err
			} else {
				// 这种情况可能是正好是 2048 的 size，再去读就会失败，但这个失败不是期望的, 所以是 nil
				return ret, nil
			}
		}
		if n != 2048 {
			if len(ret) == 0 {
				return m[:n], nil
			} else {
				ret = append(ret, m[:n]...)
				return ret, nil
			}
		}
		ret = append(ret, m[:n]...)
	}

}

// NewTimeoutConn 返回一个 conn 的 wrapper, 对 conn 的 Read 方法和 Write 的方法做了 timeout 的限制，以防无限等待
func NewTimeoutConn(conn net.Conn, defaultTimeout time.Duration) net.Conn {
	return &TimeoutConn{
		Conn:    conn,
		timeout: defaultTimeout,
	}
}

func NewTimeoutConnFromAddr(addr string, defaultTimeout time.Duration) (*TimeoutConn, error) {
	conn, err := net.DialTimeout("tcp", addr, defaultTimeout)
	if err != nil {
		return nil, err
	}
	return NewTimeoutConn(conn, defaultTimeout).(*TimeoutConn), nil
}
