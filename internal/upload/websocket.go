package upload

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/util"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Rely on other auth mechanisms
	},
}

type WSMessage struct {
	Type      string `json:"type"`
	Filename  string `json:"filename,omitempty"`
	Expiry    string `json:"expiry,omitempty"`
	Password  string `json:"password,omitempty"`
	Random    bool   `json:"random,omitempty"`
	Size      int64  `json:"size,omitempty"`
	DeleteKey string `json:"delete_key,omitempty"`
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	var upReq Request
	upReq.expiry = config.Default.MaxExpiry.Duration

	// 1. Read obfuscated metadata
	messageType, msg, err := conn.ReadMessage()
	if err != nil || messageType != websocket.BinaryMessage {
		return
	}

	// The client sends the XORed JSON as a Base64 string in a binary packet
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(msg)))
	n, err := base64.StdEncoding.Decode(decoded, msg)
	if err != nil {
		return
	}

	// XOR back to get original JSON
	originalJson := util.XorTransform(decoded[:n], 0)

	var wsMsg WSMessage
	if err := json.Unmarshal(originalJson, &wsMsg); err != nil {
		return
	}

	upReq.filename = wsMsg.Filename
	upReq.expiry = ParseExpiry(wsMsg.Expiry)
	upReq.accessKey = wsMsg.Password
	upReq.randomBarename = wsMsg.Random
	upReq.size = wsMsg.Size
	upReq.deleteKey = wsMsg.DeleteKey

	// 2. Stream file data
	pr, pw := io.Pipe()
	upReq.src = io.LimitReader(pr, upReq.size)

	go func() {
		defer pw.Close()
		buf := make([]byte, 32*1024)
		xorOffset := 0
		for {
			messageType, reader, err := conn.NextReader()
			if err != nil {
				return
			}
			if messageType == websocket.BinaryMessage {
				for {
					n, err := reader.Read(buf)
					if n > 0 {
						// XOR the chunk back to original
						transformed := util.XorTransform(buf[:n], xorOffset)
						xorOffset = (xorOffset + n) % len(util.ObfuscationKey)
						_, _ = pw.Write(transformed)
					}
					if err != nil {
						break
					}
				}
			}
		}
	}()

	// 3. Process upload
	upload, err := Process(r.Context(), upReq)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "error", "error": err.Error()})
		return
	}

	// 4. Send response
	resp, _ := json.Marshal(upload.JSONResponse(r))
	// XOR before Base64 encoding
	obfuscatedResp := util.XorTransform(resp, 0)
	encodedResp := base64.StdEncoding.EncodeToString(obfuscatedResp)
	_ = conn.WriteMessage(websocket.BinaryMessage, []byte(encodedResp))
}
