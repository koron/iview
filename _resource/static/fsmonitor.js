(function() {
  function isDir() {
    return location.pathname.endsWith('/');
  }

  const pathOrPattern = isDir() ? new RegExp('^' + location.pathname + '[^/]*/?$') : location.pathname;

  const interestEvents = isDir() ? [ 'create', 'write', 'remove', 'rename' ] : ['write', 'create'];

  const matchPath = isDir() ? (p) => pathOrPattern.test(p) : (p) => p == pathOrPattern;

  const worker = new SharedWorker('/_/static/fsmonitor-worker.js');

  function isIntersect(a, b) {
    return a.filter(v => b.includes(v)).length > 0;
  }

  function isInterested(path, events) {
    return matchPath(path) && isIntersect(interestEvents, events);
  }

  worker.port.onmessage = (ev) => {
    switch (ev.data[0]) {
      case 'notify':
        if (isInterested(ev.data[1], ev.data[2])) {
          // Using htmx.ajax() can prevent from reloading shared worker
          htmx.ajax('GET', location.pathname, { target: '#main', select: '#main', swap: 'outerHTML' });
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

  worker.port.postMessage(['connect', pathOrPattern, interestEvents]);
})();
