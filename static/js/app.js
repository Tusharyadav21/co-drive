// Minimal App JS - Theme & HTMX Event Handlers
(function() {
    var theme = localStorage.getItem('theme') || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    document.documentElement.classList.toggle('dark', theme === 'dark');
})();

function toggleTheme() {
    var isDark = document.documentElement.classList.toggle('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
}

document.addEventListener('DOMContentLoaded', function() {
    document.body.addEventListener('htmx:configRequest', function(evt) {
        var match = document.cookie.match(new RegExp('(^| )__csrf=([^;]+)'));
        if (match) {
            evt.detail.headers['X-CSRF-Token'] = match[2];
        }
    });

    document.body.addEventListener('htmx:afterSwap', function(e) {
        var input = e.target.querySelector('input:not([type=hidden]), textarea, select');
        if (input) input.focus();
        e.target.querySelectorAll('.flash-message[data-auto-dismiss]').forEach(function(el) {
            setTimeout(function() { if (window.htmx) htmx.remove(el); else el.remove(); }, 5000);
        });
    });
});
