(function() {
  const urlToCheck = window.location.href;
  const reloadInterval = 5 * 1000; // every 5 seconds.
  const prevTime = Date.now();

  async function checkAndReload() {
    try {
      const response = await fetch(urlToCheck, {
        method: 'HEAD',
        cache: 'no-store' // without cache
      });

      if (!response.ok) {
        return;
      }

      const dateHeader = response.headers.get('Date');
      if (!dateHeader) {
        return;
      }
      const serverTimestamp = new Date(dateHeader).getTime();

      if (serverTimestamp <= prevTime) {
        return;
      }

      // Force cache to be discarded and reloaded
      location.reload(true);
    } catch (error) {
      // TODO: stop to reload, and show error.
    }
  }

  // Check periodically
  setInterval(checkAndReload, reloadInterval);
})()
