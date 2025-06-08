console.log('fsmonitor-worker start');

const PING_INTERVAL = 15000;

let clients = [];
let streamStauts = '(N/A)';

function getClient(port) {
  for (const c of clients) {
    if (c.port === port) {
      return c
    }
  }
  return null;
}

function pingAllClients() {
  const now = Date.now();
  clients = clients.filter((c) => {
    if (c.ping > c.pong) {
      // Delete the client connection
      console.log(`disconnected: path=${c.path} type=${c.type} (len=${clients.length-1})`);
      return false;
    } else {
      // Ping with the stream status
      c.ping = now;
      c.port.postMessage(["ping", streamStauts]);
      return true;
    }
  });
}

// Checking if the client is alive
setInterval(() => pingAllClients(), PING_INTERVAL);

const eventSource = new EventSource('/_/stream/');

eventSource.onmessage = (ev) => {
  if (ev.data.length <= 0) {
    return;
  }
  var now = Date.now();
  var beDeleted = [];
  const data = JSON.parse(ev.data)
  for (const c of clients) {
    // Dispatch a message to watching clients
    if (data.path == c.path && data.type.includes(c.type)) {
      c.port.postMessage(['notify', data.path, data.type]);
    };
  }
};

eventSource.onopen = (ev) => {
  if (streamStauts != 'connected') {
    console.log('eventSource: connected');
  }
  streamStauts = 'connected';
  pingAllClients();
};

eventSource.onerror = (ev) => {
  if (streamStauts != 'disconnected') {
    console.log('eventSource: disconnected');
  }
  streamStauts = 'disconnected';
  pingAllClients();
};

onconnect = (ev) => {
  const port = ev.ports[0];
  port.onmessage = (ev) => {
    switch (ev.data[0]) {
      case 'connect':
        const path = ev.data[1];
        const type = ev.data[2];
        clients.push({
          port: port,
          path: path,
          type: type,
          ping: 0,
          pong: Date.now(),
        });
        console.log(`connected: path=${path} type=${type} (len=${clients.length})`);
        break;

      case 'pong':
        const c = getClient(port);
        c.pong = Date.now();
        break;
    }
  };
};
