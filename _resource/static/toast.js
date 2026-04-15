((g) => {

  const sheet = new CSSStyleSheet();
  sheet.replaceSync(`
      .toast {
        background-color: #0009;
        color: #fff;
        border-radius: 4px;
        padding: 0.625em 1em;
        z-index: 9999;
        position: fixed;
        transition: opacity 0.25s ease-in-out;
      }
    `);
  g.document.adoptedStyleSheets.push(sheet);

  async function sleep(msec) {
    await new Promise((resolve) => setTimeout(resolve, msec));
  }

  async function showToast(message, ev) {
    const target = ev.currentTarget;

    const toast = document.createElement('div');
    toast.classList.add('toast');
    toast.textContent = message;
    toast.style.left = (target.offsetLeft + target.offsetWidth) + "px";
    toast.style.top = target.offsetTop + "px";

    target.appendChild(toast);
    await sleep(1750);
    toast.style.opacity = "0";
    await sleep(250);
    target.removeChild(toast);
  }

  g.showToast = showToast;
})(this);
