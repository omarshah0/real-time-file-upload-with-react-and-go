package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For development only
	},
}

type ChunkData struct {
	Data     string `json:"data"`
	Offset   int64  `json:"offset"`
	Total    int64  `json:"total"`
	FileName string `json:"fileName"`
}

type Message struct {
	Type    string    `json:"type"`
	Payload ChunkData `json:"payload"`
}

type UploadState struct {
	file     *os.File
	fileName string
	total    int64
	received int64
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	uploadState := &UploadState{}

	for {
		_, p, err := conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			cleanup(uploadState)
			return
		}
		if err != nil {
			log.Println("Read error:", err)
			cleanup(uploadState)
			return
		}

		var message Message
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println("JSON decode error:", err)
			continue
		}

		switch message.Type {
		case "chunk":
			if err := handleChunk(conn, &message, uploadState); err != nil {
				log.Println("Chunk handling error:", err)
				cleanup(uploadState)
				return
			}
		}
	}
}

func handleChunk(conn *websocket.Conn, message *Message, state *UploadState) error {
	chunk := message.Payload

	// Initialize file on first chunk
	if state.file == nil {
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %v", err)
		}

		fileName := filepath.Join(uploadDir, chunk.FileName)
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		state.file = file
		state.fileName = chunk.FileName
		state.total = chunk.Total
	}

	// Extract and decode base64 data
	dataURL := chunk.Data
	commaIndex := strings.Index(dataURL, ",")
	if commaIndex != -1 {
		dataURL = dataURL[commaIndex+1:]
	}

	data, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %v", err)
	}

	// Write chunk to file
	if _, err := state.file.Write(data); err != nil {
		return fmt.Errorf("failed to write chunk: %v", err)
	}

	state.received += int64(len(data))

	// Calculate and send progress
	progress := int((float64(state.received) / float64(state.total)) * 100)
	if err := conn.WriteJSON(map[string]interface{}{
		"type":    "progress",
		"payload": fmt.Sprintf("%d", progress),
	}); err != nil {
		return fmt.Errorf("failed to send progress: %v", err)
	}

	// Request next chunk if not complete
	if state.received < state.total {
		if err := conn.WriteJSON(map[string]interface{}{
			"type": "ready-next-chunk",
		}); err != nil {
			return fmt.Errorf("failed to request next chunk: %v", err)
		}
	} else {
		// Send final success message
		if err := conn.WriteJSON(map[string]interface{}{
			"type":    "progress",
			"payload": "100",
		}); err != nil {
			return fmt.Errorf("failed to send completion status: %v", err)
		}
		cleanup(state)
	}

	return nil
}

func cleanup(state *UploadState) {
	if state.file != nil {
		state.file.Close()
		state.file = nil
	}
}
