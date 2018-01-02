var ws = new WebSocket("ws://localhost:8080/websocket"); // temporary

var keybinds = {
  modkey: "TAB",
  pause: modkey+"+p",
  sync: modkey+"+s",
  seek_r: modkey+"+RIGHT",
  seek_l: modkey+"+LEFT",
  seek_u: modkey+"+UP",
  seek_d: modkey+"+DOWN"
}


ws.onmessage = function(event){
  var cmd = event.method.split(' ');
  switch(cmd[0]){
    case "pause":
      setTimeout(function(){
        mp.set_property_bool("pause", true);
      }, event.schedule-Date.now
      );
    
    case "play"
      setTimeout(function(){
        mp.set_property_bool("pause", false);
      }, event.schedule-Date.now
      );
    case "seek":
      setTimeout(function(){
        mp.commandv("seek", cmd[1], "absolute");
      }, event.schedule-Date.now
      );
  }
}
      
function pause_handler(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: mp.getProperty("pause") ? "play" : "pause"
  });
}
function sync_handler(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: "seek " + mp.getProperty("time-pos")
  });
}

function seek_handler_r(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: "seek " + (mp.getProperty("time-pos") + 5)
  });
}
function seek_handler_l(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: "seek " + (mp.getProperty("time-pos") - 5)
  });
}
function seek_handler_u(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: "seek " + (mp.getProperty("time-pos") + 60)
  });
}
function seek_handler_d(){
  ws.send(JSON.stringify({
    timestamp: Math.floor(Date.now),
    method: "seek " + (mp.getProperty("time-pos") - 60)
  });
}

mp.add_key_binding(keybinds.pause, pause_handler);
mp.add_key_binding(keybinds.sync, sync_handler);
mp.add_key_binding(keybinds.seek_r, seek_handler_r);
mp.add_key_binding(keybinds.seek_l, seek_handler_l);
mp.add_key_binding(keybinds.seek_u, seek_handler_u);
mp.add_key_binding(keybinds.seek_d, seek_handler_d);
