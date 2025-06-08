(function() {
  const worker = new SharedWorker('/_/static/fsmonitor-worker.js');

  //let last = Date.now();

  worker.port.onmessage = (ev) => {
    switch (ev.data[0]) {
      case 'notify':
        if (ev.data[1] == location.pathname && ev.data[2] == 'write') {
          location.reload(true);
        }
        break;

      case 'ping':
        //console.log(`ping: status=${status} (${Date.now() - last})`);
        //last = Date.now()
        worker.port.postMessage(['pong']);
        const status = ev.data[1];
        // TODO: Update the stream status
        break;
    }
  };

  worker.port.postMessage(['connect', location.pathname, 'write']);
})();
