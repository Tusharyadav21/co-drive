// Shared browser client for the Co-Drive API.
//
// Every page talks to the backend through this file so that the transport rules
// live in one place: JSON encoding, the CSRF cookie/header pair the server
// requires on unsafe methods, error shape, and what an absent list means.

function csrfToken() {
    const match = document.cookie.match(/(^| )csrf_token=([^;]+)/);
    return match ? match[2] : null;
}

/**
 * Calls the API and returns the decoded body.
 * Throws an Error carrying `.status` when the server responds with an error.
 */
async function api(path, options = {}) {
    const headers = { ...(options.headers || {}) };

    let body = options.body;
    if (body && typeof body === 'object') {
        headers['Content-Type'] = 'application/json';
        body = JSON.stringify(body);
    }

    const token = csrfToken();
    if (token) {
        headers['X-CSRF-Token'] = token;
    }

    const res = await fetch(path, { ...options, headers, body });
    const data = await res.json().catch(() => null);

    if (!res.ok) {
        const err = new Error((data && data.error) || `HTTP ${res.status} Error`);
        err.status = res.status;
        err.data = data;
        throw err;
    }

    // Some responses (logout) have no body; hand back the raw response headers
    // via a non-enumerable slot so callers that need X-Challenge-ID can read it.
    if (data && typeof data === 'object') {
        Object.defineProperty(data, '_res', { value: res, enumerable: false });
    }
    return data;
}

/** Like api(), but normalises a null/absent collection to an empty array. */
api.list = async function (path, options) {
    const data = await api(path, options);
    return Array.isArray(data) ? data : [];
};

/** Sends the browser to the login page, preserving where it was headed. */
api.toLogin = function () {
    window.location.href = `/auth?return_to=${encodeURIComponent(window.location.pathname)}`;
};

/**
 * For pages that only make sense when signed in: any 401 (or network failure)
 * sends the user to /auth instead of surfacing an error.
 */
api.authed = async function (path, options) {
    try {
        return await api(path, options);
    } catch (err) {
        if (err.status === 401 || err.status === undefined) {
            api.toLogin();
        }
        throw err;
    }
};
