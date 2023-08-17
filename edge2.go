package edge

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
)

type myResponseWriter struct {
}

func (*myResponseWriter) Header() http.Header {
	return make(http.Header)
}
func (*myResponseWriter) WriteHeader(statusCode int) {
}
func (w *myResponseWriter) Flush() {
}

type myResponseWriter2 struct {
	quic.Stream
	myResponseWriter
}

type myResponseWriter3 struct {
	handshake bool
	myResponseWriter2
	quic.Connection
}

func (w *myResponseWriter3) Write(b []byte) (int, error) {
	if !w.handshake {
		w.handshake = true
		return len(b), nil
	}
	println(string(b))
	return w.Stream.Write(b)
}

func (w *myResponseWriter3) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return net.Conn(w), bufio.NewReadWriter(bufio.NewReader(w), bufio.NewWriter(w)), nil
}

type Edge2Config struct {
	config.HTTP
	Origin string //源服务器地址
}

var Edge2Plugin = InstallPlugin(new(Edge2Config))

func (cfg *Edge2Config) OnEvent(event any) {
	switch event.(type) {
	case FirstConfig:
		if cfg.Origin != "" {
			go cfg.ConnectToOrigin()
		}
	}
}

func (cfg *Edge2Config) Connect() (wasConnected bool, err error) {
	ctx := Edge2Plugin.Context
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"monibuca"},
	}
	var conn quic.Connection
	conn, err = quic.DialAddr(ctx, cfg.Origin, tlsConf, &quic.Config{
		KeepAlivePeriod: time.Second * 10,
		EnableDatagrams: true,
	})
	wasConnected = err == nil
	panic(conn)
	// if stream := quic.Stream(nil); err == nil {
	// 	if stream, err = conn.OpenStreamSync(ctx); err == nil {
	// 	}
	// }
	// for err == nil {
	// 	var s quic.Stream
	// 	if s, err = conn.AcceptStream(ctx); err == nil {
	// 		go cfg.ReceiveRequest(s, conn)
	// 	}
	// }
	return wasConnected, err
}

func (cfg *Edge2Config) ConnectToOrigin() {
	retryDelay := [...]int{2, 3, 5, 8, 13}
	for i := 0; Edge2Plugin.Err() == nil; i++ {
		connected, err := cfg.Connect()
		if err == nil {
			// 不需要重试了，服务器返回了错误
			return
		}
		Edge2Plugin.Error("connect to origin server error", zap.Error(err))
		if connected {
			i = 0
		} else if i >= 5 {
			i = 4
		}
		time.Sleep(time.Second * time.Duration(retryDelay[i]))
	}
}

// func (cfg *Edge2Config) ReceiveRequest(s quic.Stream, conn quic.Connection) error {
// 	defer s.Close()
// 	wr := &myResponseWriter2{Stream: s}
// 	reader := bufio.NewReader(s)
// 	var req *http.Request
// 	url, _, err := reader.ReadLine()
// 	if err == nil {
// 		ctx, cancel := context.WithCancel(s.Context())
// 		defer cancel()
// 		req, err = http.NewRequestWithContext(ctx, "GET", string(url), reader)
// 		for err == nil {
// 			var h []byte
// 			if h, _, err = reader.ReadLine(); len(h) > 0 {
// 				if b, a, f := strings.Cut(string(h), ": "); f {
// 					req.Header.Set(b, a)
// 				}
// 			} else {
// 				break
// 			}
// 		}

// 		if err == nil {
// 			h, _ := cfg.mux.Handler(req)
// 			if req.Header.Get("Accept") == "text/event-stream" {
// 				go h.ServeHTTP(wr, req)
// 			} else if req.Header.Get("Upgrade") == "websocket" {
// 				var writer myResponseWriter3
// 				writer.Stream = s
// 				writer.Connection = conn
// 				req.Host = req.Header.Get("Host")
// 				if req.Host == "" {
// 					req.Host = req.URL.Host
// 				}
// 				if req.Host == "" {
// 					req.Host = "m7s.live"
// 				}
// 				h.ServeHTTP(&writer, req) //建立websocket连接,握手
// 			} else {
// 				h.ServeHTTP(wr, req)
// 			}
// 		}
// 		io.ReadAll(s)
// 	}
// 	if err != nil {
// 		log.Error("read console server error:", err)
// 	}
// 	return err
// }
