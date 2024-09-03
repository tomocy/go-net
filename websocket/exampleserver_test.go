// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/websocket"
)

// ExampleServer_Handshake demonstrates the use of Server with a custom handshake.
// The handshake checks if the Origin request header matches the server's expected origin,
// and selects a supported subprotocol if one is present in the Sec-WebSocket-Protocol header.
func ExampleServer_Handshake() {
	origin, _ := url.Parse("http://localhost")
	server := httptest.NewServer(websocket.Server{
		Config: websocket.Config{
			Origin: origin,
		},
		Handshake: func(cfg *websocket.Config, r *http.Request) error {
			origin := r.Header.Get("Origin")
			if origin != cfg.Origin.String() {
				return websocket.ErrBadWebSocketOrigin
			}

			protocol := "supported_3"
			if slices.Contains(cfg.Protocol, protocol) {
				cfg.Protocol = []string{protocol}
			}

			return nil
		},
		Handler: func(conn *websocket.Conn) {
			_, _ = conn.Write([]byte("hello, websocket!"))
		},
	})
	defer server.Close()

	url := fmt.Sprintf("ws%s", strings.TrimPrefix(server.URL, "http"))

	{
		cfg, _ := websocket.NewConfig(url, fmt.Sprintf("%s/bad", origin))
		if _, err := cfg.DialContext(context.Background()); err != nil {
			fmt.Println("error: connected with bad origin")
		}
	}
	{
		cfg, _ := websocket.NewConfig(url, origin.String())
		cfg.Protocol = []string{"unsupported_1", "unsupported_2"}
		if _, err := cfg.DialContext(context.Background()); err != nil {
			fmt.Println("error: connected with unsupported protocols")
		}
	}
	{
		cfg, _ := websocket.NewConfig(url, origin.String())
		cfg.Protocol = []string{"supported_1", "supported_2", "supported_3"}
		ws, _ := cfg.DialContext(context.Background())

		msg := make([]byte, 1024)
		n, _ := ws.Read(msg)
		fmt.Printf("ok: %q\n", string(msg[:n]))
	}

	// Output:
	// error: connected with bad origin
	// error: connected with unsupported protocols
	// ok: "hello, websocket!"
}
