console.log('fsmonitor-worker start');

const PING_INTERVAL = 15000;

let clients = [];
let streamStatus = undefined;

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
      console.log(`disconnected: path=${c.path} type=${c.type} ping/pong=${c.ping}/0${c.pong} (len=${clients.length-1})`);
      return false;
    } else {
      // Ping with the stream status
      c.ping = now;
      c.port.postMessage(["ping", streamStatus]);
      return true;
    }
  });
}

// Checking if the client is alive
setInterval(() => pingAllClients(), PING_INTERVAL);

const eventSource = new EventSource('/_/stream/');

function isIntersect(a, b) {
  return a.filter(v => b.includes(v)).length > 0;
}

eventSource.onmessage = (ev) => {
  if (ev.data.length <= 0) {
    return;
  }
  var now = Date.now();
  var beDeleted = [];
  const data = JSON.parse(ev.data);
  for (const c of clients) {
    // Dispatch a message to watching clients
    if (c.pchk(data.path) && isIntersect(data.type, c.type)) {
      c.port.postMessage(['notify', data.path, data.type]);
    };
    //console.log('c.pchk', data.path, data.type, c.pchk(data.path), c.type);
  }
};

eventSource.onopen = (ev) => {
  if (streamStatus !== true) {
    console.log('eventSource: connected');
  }
  streamStatus = true;
  pingAllClients();
};

eventSource.onerror = (ev) => {
  if (streamStatus !== false) {
    console.log('eventSource: disconnected');
  }
  streamStatus = false;
  pingAllClients();
};

onconnect = (ev) => {
  const port = ev.ports[0];
  port.onmessage = (ev) => {
    switch (ev.data[0]) {
      case 'connect':
        const path = ev.data[1];
        const type = ev.data[2];
        const now = Date.now();
        clients.push({
          port: port,
          path: path,
          type: type,
          ping: now,
          pong: now,
          pchk: path instanceof RegExp ? (p) => path.test(p) : (p) => p == path,
        });
        port.postMessage(["ping", streamStatus]);
        console.log('connected:\n', 'path:', path, '\n', 'type:', type, '\n', 'clients.length:', clients.length);
        break;

      case 'pong':
        const c = getClient(port);
        if (c) {
          c.pong = Date.now();
        }
        break;
    }
  };
};
