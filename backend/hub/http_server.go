package hub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// StartHub starts the HTTP server and handles shutdown on context cancellation.
func StartHub(ctx context.Context) {
	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/tools", toolHandler(ctx))
	http.HandleFunc("/ws/stream_tools", streamToolHandler)
	server := &http.Server{Addr: "0.0.0.0:9573"}
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	log.Fatal(server.ListenAndServe())
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func toolHandler(appCtx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		toolName := r.URL.Path[len("/api/cmd_tools/"):]
		if toolName[len(toolName)-1] == '/' {
			toolName = toolName[:len(toolName)-1]
		}
		if toolName == "" || strings.Contains(toolName, "/") {
			http.Error(w, "Invalid toolName", http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		tool, err := gorm.G[Tool](db).Where("name = ?", toolName).Take(ctx)
		if err != nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}
		switch tool.Type {
		case "command_line":
			cmdTool, err := gorm.G[CommandLineTool](db).Where("id = ?", tool.ID).First(ctx)
			if err != nil {
				http.Error(w, "CommandLineTool not found", http.StatusNotFound)
				return
			}
			if cmdTool.ConcurrencyGroupID != nil {
				cg, err := gorm.G[ConcurrencyGroup](db).Where("id = ?", cmdTool.ConcurrencyGroupID).First(ctx)
				if err != nil {
					http.Error(w, "ConcurrencyGroup not found", http.StatusInternalServerError)
					return
				}
				cmdTool.ConcurrencyGroup = &cg
			}
			cmdToolHandler(cmdTool, r, w)
		case "http":
			httpTool, err := gorm.G[HTTPTool](db).Preload("ConcurrencyGroup", func(_ gorm.PreloadBuilder) error { return nil }).Where("id = ?", tool.ID).Take(ctx)
			if err != nil {
				http.Error(w, "HttpTool not found", http.StatusNotFound)
				return
			}
			if httpTool.ConcurrencyGroupID != nil {
				cg, err := gorm.G[ConcurrencyGroup](db).Where("id = ?", httpTool.ConcurrencyGroupID).First(ctx)
				if err != nil {
					http.Error(w, "ConcurrencyGroup not found", http.StatusInternalServerError)
					return
				}
				httpTool.ConcurrencyGroup = &cg
			}
			httpToolHandler(ctx, httpTool, r, w)
		}
	}
}

var upgrader = websocket.Upgrader{}

func streamToolHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Echo message back
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
