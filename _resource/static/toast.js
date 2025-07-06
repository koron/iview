async function showToast(message, ev) {
  async function sleep(msec) {
    await new Promise((resolve) => setTimeout(resolve, msec));
  }

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
