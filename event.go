package main

type Event struct {
  timestamp uint // when the event was created
  schedule uint // when the event is scheduled to resolve
  method string // method of the event (e.g. pause)
  owner uint // id of owner client
  vidTime uint // what time the video was at when the event happened
}

/*
In JSON, events look like this:
  {
    "timestamp": 123456789
    "schedule": 123456790
    "method": "pause"
  }
*/
