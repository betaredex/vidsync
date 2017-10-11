package main

import (
  "fmt"
  "encoding/json"
  "time"
  "net/http"
  "github.com/gorilla/websocket"
)

const (
  writeWait = 10 * time.Second
  pongWait = 60 * time.Second
  pingPeriod = 55 * time.Second // Has to be smaller than pongWait
)

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024
}

type Client struct {
  hub *Hub
  conn *websocket.Conn
  send chan *Event
  sentTime uint
  meanLatency uint
  connections uint
}

func (c *Client) updateLatency(newLatency uint){
  c.meanLatency = (c.meanLatency*c.connections+newLatency)/(c.connections+1) // update mean
  c.connections++
}

func (c *Client) readPump() {
  var event Event
  defer func() {
    c.hub.unregister <- c
    c.conn.Close()
  }()
  c.conn.SetReadDeadline(time.Now().Add(pongWait))
  c.conn.SetPongHandler(
    func(string) error {
      c.conn.SetReadDeadline(time.Now().Add(pongWait)) // Deadline will be reached if peer stops replying
      c.updateLatency((makeTimestamp-c.sentTime)/2)
      return nil
    }
  }
  for {
    err := c.readJson(event)
    if err != nil {
      fmt.Printf("error: %v\n", err)
      break
    }
    c.updateLatency(makeTimestamp() - event.timestamp)
    c.hub.event <- *event
  }
}

func (c *Client) writePump() {
  var event Event
  ticker := time.NewTicker(pingPeriod)
  defer func() {
    ticker.Stop()
    c.conn.Close()
    for {
      select {
      case event := *<-c.send:
        c.conn.SetWriteDeadline(time.Now().Add(writeWait))
        w, err := c.conn.NextWriter(websocket.TextMessage)
        if err != nil {
          return
        }
        message, err := json.Marshal(event)
        if err != nil {
          fmt.Printf("error: %v\n", err)
          return
        }
        w.Write(message)
        if err := w.Close(); err != nil {
          return
        }
      case <-ticker.C:
        c.conn.SetWriteDeadline(time.Now().Add(writeWait))
        err := c.conn.WriteMessage(websocket.pingMessage, []byte{})
        if err != nil {
          fmt.Printf("error: %v\n", err)
          return
        }
      }
    }
  }
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request){
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Printf("error: %v\n", err)
    return
  }
  client := &Client {
    hub: hub
    conn: conn
    send: make(chan *Event)
    sentTime: 0
    meanLatency: 0
    connections: 0
  }
  hub.register <- client

  go client.writePump()
  go client.readPump()
}
