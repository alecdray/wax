(function () {
  var pending = 0;
  var bar = null;
  function getBar() { return bar || (bar = document.getElementById("global-progress")); }
  document.addEventListener("htmx:beforeSend", function () {
    var b = getBar();
    if (pending === 0 && b) {
      b.classList.remove("done");
      void b.offsetWidth; // force reflow so width resets to 0 before animating
      b.classList.add("loading");
    }
    pending++;
  });
  document.addEventListener("htmx:afterRequest", function () {
    pending = Math.max(0, pending - 1);
    var b = getBar();
    if (pending === 0 && b) {
      b.classList.remove("loading");
      b.classList.add("done");
    }
  });
})();
