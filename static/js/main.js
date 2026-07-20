/* Haven — Client-side JS
   Theme persistence, mobile menu, form submissions, data-remove actions,
   toast notification dismiss */

(function () {
  'use strict';

  /* ---- Theme -------------------------------------------------- */
  const html = document.documentElement;
  const themeToggle = document.getElementById('theme-toggle');
  const themeSwitch = document.getElementById('theme-switch');
  const themeLabel = document.getElementById('theme-label');

  function getTheme() {
    const stored = localStorage.getItem('haven-theme');
    if (stored === 'dark' || stored === 'light') return stored;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function applyTheme(theme) {
    html.setAttribute('data-theme', theme);
    localStorage.setItem('haven-theme', theme);
    if (themeSwitch) themeSwitch.checked = (theme === 'dark');
    if (themeLabel) themeLabel.textContent = theme === 'dark' ? 'Dark' : 'Light';
  }

  applyTheme(getTheme());

  function toggleTheme() {
    const current = html.getAttribute('data-theme');
    applyTheme(current === 'dark' ? 'light' : 'dark');
  }

  if (themeToggle) themeToggle.addEventListener('click', toggleTheme);
  if (themeSwitch) themeSwitch.addEventListener('change', toggleTheme);

  /* ---- Mobile menu -------------------------------------------- */
  const mobileBtn = document.getElementById('mobile-menu-btn');
  const mobileNav = document.getElementById('mobile-nav');
  if (mobileBtn && mobileNav) {
    mobileBtn.addEventListener('click', function () {
      const open = mobileNav.classList.toggle('open');
      mobileBtn.setAttribute('aria-expanded', open ? 'true' : 'false');
    });
  }

  /* ---- Toast dismiss ------------------------------------------ */
  document.addEventListener('click', function (e) {
    const btn = e.target.closest('[data-dismiss-alert]');
    if (!btn) return;
    const toast = btn.closest('.triggered-toast');
    if (toast) {
      toast.classList.add('toast-dismissed');
      toast.addEventListener('animationend', function () { toast.remove(); });
    }
  });

  /* ---- Remove holding ----------------------------------------- */
  document.addEventListener('click', function (e) {
    const btn = e.target.closest('[data-remove-holding]');
    if (!btn) return;
    const id = btn.getAttribute('data-remove-holding');
    if (!confirm('Remove this holding?')) return;

    fetch('/api/portfolio/remove/' + encodeURIComponent(id), { method: 'DELETE' })
      .then(function (r) {
        if (!r.ok) return r.json().then(function (j) { throw new Error(j.error || 'Failed'); });
        window.location.reload();
      })
      .catch(function (err) { alert('Error: ' + err.message); });
  });

  /* ---- Remove alert ------------------------------------------- */
  document.addEventListener('click', function (e) {
    const btn = e.target.closest('[data-remove-alert]');
    if (!btn) return;
    const id = btn.getAttribute('data-remove-alert');
    if (!confirm('Delete this alert?')) return;

    fetch('/api/alerts/' + encodeURIComponent(id), { method: 'DELETE' })
      .then(function (r) {
        if (!r.ok) return r.json().then(function (j) { throw new Error(j.error || 'Failed'); });
        window.location.reload();
      })
      .catch(function (err) { alert('Error: ' + err.message); });
  });

  /* ---- Add holding form --------------------------------------- */
  var holdingForm = document.getElementById('add-holding-form');
  var holdingErr = document.getElementById('holding-error');
  if (holdingForm) {
    holdingForm.addEventListener('submit', function (e) {
      e.preventDefault();
      if (holdingErr) holdingErr.style.display = 'none';

      var fd = new FormData(holdingForm);
      var body = JSON.stringify({
        coin_id: fd.get('coin_id'),
        amount: parseFloat(fd.get('amount')),
        purchase_price: parseFloat(fd.get('purchase_price'))
      });

      fetch('/api/portfolio/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body
      })
        .then(function (r) {
          return r.json().then(function (j) {
            if (!r.ok) throw new Error(j.error || 'Failed to add holding');
            return j;
          });
        })
        .then(function () { window.location.reload(); })
        .catch(function (err) {
          if (holdingErr) { holdingErr.textContent = err.message; holdingErr.style.display = 'block'; }
        });
    });
  }

  /* ---- Add alert form ----------------------------------------- */
  var alertForm = document.getElementById('add-alert-form');
  var alertErr = document.getElementById('alert-error');
  if (alertForm) {
    alertForm.addEventListener('submit', function (e) {
      e.preventDefault();
      if (alertErr) alertErr.style.display = 'none';

      var fd = new FormData(alertForm);
      var body = JSON.stringify({
        coin_id: fd.get('coin_id'),
        direction: fd.get('direction'),
        target_price: parseFloat(fd.get('target_price'))
      });

      fetch('/api/alerts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body
      })
        .then(function (r) {
          return r.json().then(function (j) {
            if (!r.ok) throw new Error(j.error || 'Failed to create alert');
            return j;
          });
        })
        .then(function () { window.location.reload(); })
        .catch(function (err) {
          if (alertErr) { alertErr.textContent = err.message; alertErr.style.display = 'block'; }
        });
    });
  }
})();
