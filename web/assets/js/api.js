// Klien API tipis untuk dashboard CURANMOR AI.
// Sesi diautentikasi lewat cookie httpOnly (diset oleh /api/v1/auth/login),
// sehingga fetch() cukup menyertakan credentials: "include".
const Api = (() => {
  async function request(method, path, body, opts) {
    opts = opts || {};
    const headers = {};
    let payload = body;
    if (body && !(body instanceof FormData)) {
      headers["Content-Type"] = "application/json";
      payload = JSON.stringify(body);
    }
    const res = await fetch(path, {
      method,
      headers,
      body: payload,
      credentials: "include",
    });
    let json = null;
    try {
      json = await res.json();
    } catch (e) {
      /* respons tanpa body (mis. 204) */
    }
    if (res.status === 401 && !opts.skipAuthRedirect) {
      window.location.href = "/index.html?expired=1";
      return Promise.reject(new Error("Sesi berakhir"));
    }
    if (!res.ok) {
      const msg = (json && json.error && json.error.message) || "Terjadi kesalahan (" + res.status + ")";
      const err = new Error(msg);
      err.status = res.status;
      err.body = json;
      throw err;
    }
    return json;
  }

  return {
    get: (path, opts) => request("GET", path, null, opts),
    post: (path, body, opts) => request("POST", path, body, opts),
    put: (path, body, opts) => request("PUT", path, body, opts),
    patch: (path, body, opts) => request("PATCH", path, body, opts),
    del: (path, opts) => request("DELETE", path, null, opts),
    async upload(file) {
      const fd = new FormData();
      fd.append("file", file);
      const json = await request("POST", "/api/v1/uploads", fd);
      return json.data.url;
    },
  };
})();

// Format util
function fmtDate(s) {
  if (!s) return "-";
  const d = new Date(s);
  if (isNaN(d.getTime())) return s;
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}
function fmtDateTime(s) {
  if (!s) return "-";
  const d = new Date(s);
  if (isNaN(d.getTime())) return s;
  return d.toLocaleString("id-ID", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });
}
function statusBadgeClass(status) {
  const map = {
    LP: "badge-lp", SP2HP: "badge-sp2hp", DPO: "badge-dpo", Selesai: "badge-selesai",
    "Terlapor Hilang": "badge-hilang", Ditemukan: "badge-ditemukan",
    Diamankan: "badge-diamankan", Dikembalikan: "badge-dikembalikan",
  };
  return "badge " + (map[status] || "badge-lp");
}
function escapeHtml(s) {
  if (s === null || s === undefined) return "";
  return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}
function qs(sel, root) { return (root || document).querySelector(sel); }
function qsa(sel, root) { return Array.from((root || document).querySelectorAll(sel)); }
