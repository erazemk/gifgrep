const toast = (() => {
  const el = document.createElement("div");
  el.className = "toast";
  document.body.appendChild(el);
  let t = null;
  return (msg) => {
    el.textContent = msg;
    el.classList.add("on");
    clearTimeout(t);
    t = setTimeout(() => el.classList.remove("on"), 900);
  };
})();

async function copyText(s) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(s);
    return true;
  }

  const ta = document.createElement("textarea");
  ta.value = s;
  ta.style.position = "fixed";
  ta.style.top = "-1000px";
  ta.style.left = "-1000px";
  document.body.appendChild(ta);
  ta.focus();
  ta.select();
  const ok = document.execCommand("copy");
  document.body.removeChild(ta);
  return ok;
}

document.addEventListener("click", async (e) => {
  const btn = e.target.closest("button.copy");
  if (!btn) return;
  const cmd = btn.getAttribute("data-copy") || "";
  if (!cmd) return;
  try {
    await copyText(cmd);
    toast("copied");
  } catch {
    toast("copy failed (skill issue)");
  }
});

