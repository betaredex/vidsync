package main

import "time"

type Hub struct {
  clients map[*Client]bool
  event chan []byte
  join chan *Client
  leave chan *Client
}

func NewHub() *Hub {
  return &Hub {
    clients: make(map[*Client]bool)
    event: make(chan *Event)
    join: make(chan *Client)
    leave: make(chan *Client)
  }
}

func (h *Hub) run() {
  for {
    select {
    case client := <-h.join:
      h.clients[client] = true
    case client := <-h.leave:
      if _, ok := h.clients[client]; ok {
        delete(h.clients, client)
      }
      close(client.send)
    case event := <-h.event:
      max uint = 0
      for client := range h.clients {
        if client.meanLatency > max {
          max = client.meanLatency
        }
      }
      sync *Event = {
        timestamp = makeTimestamp()
        schedule = makeTimestamp() + max
        method = event.method
      }
      for client := range h.clients {
        client.send <- sync
      }
    }
  }
}

