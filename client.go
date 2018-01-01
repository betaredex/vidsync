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
  CheckOrigin: func(r *http.Request) bool {return true}, // Allow requests from any origin
}

type Client struct {
  hub *Hub // The hub this client is connected to
  conn *websocket.Conn // The connection from the client object to the peer
  send chan *Event // events to send to the peer
  sentTime uint // the last time we pinged the peer
  meanLatency uint // average latency of the peer
  connections uint // number of times the peer connected to the client
}

func (c *Client) updateLatency(newLatency uint){
  c.meanLatency = (c.meanLatency*c.connections+newLatency)/(c.connections+1) // update mean
  c.connections++
}

func (c *Client) readPump() {
  var event Event
  defer func() { // leave the hub when the function breaks
    c.hub.leave <- c
    c.conn.Close()
  }()
  c.conn.SetReadDeadline(time.Now().Add(pongWait))
  c.conn.SetPongHandler(func(string) error {
      c.conn.SetReadDeadline(time.Now().Add(pongWait)) // Deadline will be reached if peer stops replying
      c.updateLatency((makeTimestamp()-c.sentTime)/2) // updates the latency based on the ping
      return nil
  })
  for {
    err := c.conn.ReadJSON(event) // stores the json from the connection in 'event', after decoding it. see event.go for details
    if err != nil {
      fmt.Printf("error: %v\n", err)
      break
    }
    c.updateLatency(makeTimestamp() - event.timestamp) // new latency measurement by comparing the timestamp of the event's creation with the current time
    c.hub.event <- &event // send the event to the hub
  }
}

func (c *Client) writePump() {
  ticker := time.NewTicker(pingPeriod) // timer to ping the peer once every pingPeriod
  defer func() { // cleanup when the fuction breaks
    ticker.Stop()
    c.conn.Close()
  }()
  for {
    select {
    case event := <-c.send: // new event from hub
      c.conn.SetWriteDeadline(time.Now().Add(writeWait))
      w, err := c.conn.NextWriter(websocket.TextMessage) // get a writer
      if err != nil {
        return
      }
      message, err := json.Marshal(*event) // convert the event to JSON
      if err != nil {
        fmt.Printf("error: %v\n", err)
        return
      }
      w.Write(message) // send the event to the peer
      if err := w.Close(); err != nil {
        return
      }
    case <-ticker.C: // timer is up
      c.conn.SetWriteDeadline(time.Now().Add(writeWait))
      err := c.conn.WriteMessage(websocket.PingMessage, []byte{}) // ping the peer
      if err != nil {
        fmt.Printf("error: %v\n", err)
        return
      }
      c.sentTime = makeTimestamp() // record when the ping was sent
    }
  }
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request){ // serves the websocket to the peer
  conn, err := upgrader.Upgrade(w, r, nil) // upgrades the connection from http
  if err != nil {
    fmt.Printf("error: %v\n", err)
    return
  }
  client := &Client { // initialize the client
    hub: hub,
    conn: conn,
    send: make(chan *Event),
    sentTime: 0,
    meanLatency: 0,
    connections: 0,
  }
  hub.join <- client // add the client to the hub

  go client.writePump() // start our pump functions
  go client.readPump()
}
