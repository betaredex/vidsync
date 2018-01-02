package main

type Hub struct {
  clients map[*Client]bool // clients currently attached to hub
  event chan *Event // event queue
  join chan *Client // join queue
  leave chan *Client // leave queue
}

func newHub() *Hub { // basic constructor, nothing special here
  return &Hub {
    clients: make(map[*Client]bool),
    event: make(chan *Event),
    join: make(chan *Client),
    leave: make(chan *Client),
  }
}

func (h *Hub) run() {
  for {
    select {
    case client := <-h.join: // add joining clients to the hub
      h.clients[client] = true
    case client := <-h.leave: // remove leaving clients from the hub
      if _, ok := h.clients[client]; ok {
        delete(h.clients, client)
      }
      close(client.send) // close connection
    case event := <-h.event:
      var max uint = 0
      for client := range h.clients { // find the client with the maximum latency and use that for scheduling
        if client.meanLatency > max {
          max = client.meanLatency
        }
      }
      sync := &Event {
        timestamp: makeTimestamp(),
        schedule: makeTimestamp() + max,
        method: event.method,
      }
      for client := range h.clients { // send the event to all the clients
        client.send <- sync
      }
    }
  }
}
