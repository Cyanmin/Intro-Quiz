package main

import (
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("upgrade:", err)
        return
    }
    defer conn.Close()
    for {
        mt, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            break
        }
        log.Printf("recv: %s", message)
        if err = conn.WriteMessage(mt, message); err != nil {
            log.Println("write:", err)
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", echo)
    log.Println("Listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
