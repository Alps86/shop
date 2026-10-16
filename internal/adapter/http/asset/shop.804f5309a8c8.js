document.addEventListener('htmx:beforeRequest', function (event) {
  const region = document.querySelector('#results');
  if (region) region.setAttribute('aria-busy', 'true');
});
document.addEventListener('htmx:afterRequest', function () {
  const region = document.querySelector('#results');
  if (region) region.setAttribute('aria-busy', 'false');
});
document.addEventListener('htmx:beforeSwap', function (event) {
  if ([409, 410, 422, 503].includes(event.detail.xhr.status)) {
    event.detail.shouldSwap = true;
    event.detail.isError = false;
  }
});
document.addEventListener('click', function (event) {
  if (event.target.closest('.buy[disabled]')) event.preventDefault();
});
document.addEventListener('htmx:afterSwap', function () {
  const alert = document.querySelector('[role="alert"]');
  if (alert) {
    alert.tabIndex = -1;
    alert.focus();
  }
});
