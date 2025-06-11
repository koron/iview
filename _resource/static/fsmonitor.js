(function() {
  const INTEREST_EVENTS = ['write', 'create'];

  const worker = new SharedWorker('/_/static/fsmonitor-worker.js');

  worker.port.onmessage = (ev) => {
    switch (ev.data[0]) {
      case 'notify':
        if (ev.data[1] == location.pathname && INTEREST_EVENTS.filter(v => ev.data[2].includes(v)).length > 0) {
          // Using htmx.ajax() can prevent from reloading shared worker
          htmx.ajax('GET', location.pathname);
        }
        break;

      case 'ping':
        worker.port.postMessage(['pong']);
        const status = ev.data[1];
        // Update the stream status
        const el = document.querySelector('#status');
        if (status !== undefined) {
          el.innerText = status ? 'OK' : 'NG';
          el.classList.toggle('ok', status);
          el.classList.toggle('ng', !status);
        }
        break;
    }
  };

  worker.port.postMessage(['connect', location.pathname, INTEREST_EVENTS]);
})();
