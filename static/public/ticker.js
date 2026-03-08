(function () {
  var track = document.getElementById("ticker-track");
  var wrap = track && track.parentElement;
  if (!track || !wrap) return;

  var wrapWidth = wrap.offsetWidth;
  var oneSetWidth = track.scrollWidth;
  if (oneSetWidth === 0) return;

  var minCopies = Math.ceil(wrapWidth / oneSetWidth) + 1;
  var originalItems = Array.from(track.children);

  // Pad the first set so it is wider than the viewport
  var extra = minCopies - 1;
  while (extra--) {
    originalItems.forEach(function (item) {
      var clone = item.cloneNode(true);
      clone.setAttribute("aria-hidden", "true");
      track.appendChild(clone);
    });
  }

  // Duplicate the padded set to create a seamless second half
  var paddedItems = Array.from(track.children);
  paddedItems.forEach(function (item) {
    var clone = item.cloneNode(true);
    clone.setAttribute("aria-hidden", "true");
    track.appendChild(clone);
  });

  var oneLoop = track.scrollWidth / 2;
  var duration = oneLoop / 30;

  track.style.setProperty("--scroll-to", "-50%");
  track.style.setProperty("--duration", duration + "s");
  track.classList.add("scrolling");
})();
