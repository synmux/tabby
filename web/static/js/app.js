// Tabby — Minimal client-side JS (clock, theme, SSE flash indicators)

(function () {
  "use strict";

  // ── Clock ──

  function updateClock() {
    const clockElement = document.getElementById("clock");
    const dateElement = document.getElementById("clock-date");
    if (!clockElement) return;

    const now = new Date();

    clockElement.textContent = now.toLocaleTimeString("en-GB", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });

    if (dateElement) {
      dateElement.textContent = now.toLocaleDateString("en-GB", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
      });
    }
  }

  updateClock();
  setInterval(updateClock, 1000);

  // ── SSE update flash indicator ──

  document.body.addEventListener("htmx:sseMessage", function (event) {
    const target = event.detail.elt;
    if (!target) return;

    const card = target.closest(".widget-card");
    if (!card) return;

    // Flash the indicator dot
    const indicator = card.querySelector(".widget-indicator");
    if (indicator) {
      indicator.classList.remove("widget-indicator--active");
      // Force reflow to restart the animation
      void indicator.offsetWidth;
      indicator.classList.add("widget-indicator--active");
    }

    // Flash the card border
    card.removeAttribute("data-sse-updated");
    void card.offsetWidth;
    card.setAttribute("data-sse-updated", "");
  });
})();
