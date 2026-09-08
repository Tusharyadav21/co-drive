/**
 * Co-Drive Layout Components & Common UI Utilities
 */

// Define <co-drive-header> Web Component
class CoDriveHeader extends HTMLElement {
    connectedCallback() {
        const active = this.getAttribute('active') || '';
        this.innerHTML = `
        <header>
            <div class="header-inner">
                <div style="display: flex; align-items: center; gap: 20px;">
                    <a href="/dashboard" class="logo">
                        <div class="logo-mark">C</div>
                        Co-Drive
                    </a>
                    <nav class="nav-links">
                        <a href="/dashboard" class="nav-btn ${active === 'dashboard' ? 'active' : ''}">Dashboard</a>
                        <a href="/logs" class="nav-btn ${active === 'logs' ? 'active' : ''}">Logs</a>
                        <a href="/account" class="nav-btn ${active === 'account' ? 'active' : ''}">Account</a>
                    </nav>
                </div>
                <div class="header-actions">
                    <a href="/account" class="user-pill" title="Driver Account">
                        <div class="user-avatar" id="user-avatar">U</div>
                        <span id="user-email-text">Driver</span>
                    </a>
                </div>
            </div>
        </header>
        `;
    }
}
customElements.define('co-drive-header', CoDriveHeader);

// Define <co-drive-footer> Web Component
class CoDriveFooter extends HTMLElement {
    connectedCallback() {
        this.innerHTML = `
        <footer>
            <div class="footer-inner">
                <div>&copy; 2026 Co-Drive. Drive Responsibly</div>
            </div>
        </footer>
        `;
    }
}
customElements.define('co-drive-footer', CoDriveFooter);

// Shared Global Toast
let toastTimer = null;
window.showToast = function (msg, isError = false) {
    let toast = document.getElementById('toast');
    if (!toast) {
        toast = document.createElement('div');
        toast.id = 'toast';
        document.body.appendChild(toast);
    }
    toast.textContent = msg;
    toast.style.background = isError ? 'var(--error, #E11D48)' : 'var(--primary, #0F172A)';
    toast.style.display = 'block';

    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        toast.style.display = 'none';
    }, 2800);
};

// Global User Header Session Sync
window.currentUser = null;
window.initUserHeader = async function () {
    if (window.currentUser) {
        return window.currentUser;
    }
    try {
        const fetchFn = typeof api === 'function'
            ? api
            : (url) => fetch(url).then(r => { if (!r.ok) throw new Error('Auth error'); return r.json(); });

        const user = await fetchFn('/api/users/me');
        window.currentUser = user;

        const name = user.full_name || user.name || user.email || 'Driver';
        const avatar = name[0].toUpperCase();

        const avatarEl = document.getElementById('user-avatar');
        const emailEl = document.getElementById('user-email-text');
        if (avatarEl) avatarEl.textContent = avatar;
        if (emailEl) emailEl.textContent = name;

        return user;
    } catch (e) {
        if (window.location.pathname !== '/auth' && window.location.pathname !== '/') {
            window.location.href = '/auth';
        }
        throw e;
    }
};

// Auto-run user header population when document is ready
document.addEventListener('DOMContentLoaded', () => {
    // Only attempt auto-sync on protected pages with co-drive-header
    if (document.querySelector('co-drive-header') && window.location.pathname !== '/auth' && window.location.pathname !== '/') {
        window.initUserHeader().catch(() => { });
    }
});
