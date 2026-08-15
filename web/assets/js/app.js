// Kerangka bersama seluruh halaman dashboard: memuat profil pengguna,
// menyembunyikan menu sesuai RBAC (peran_modul_akses), dan mengisi info
// sidebar. Setiap halaman menunggu `window.AppReady` sebelum memuat data
// spesifik halamannya.
window.AppReady = (async function initApp() {
  const [meRes, matriksRes] = await Promise.all([
    Api.get("/api/v1/auth/me"),
    Api.get("/api/v1/rbac/matriks"),
  ]);
  const user = meRes.data;
  const matriks = matriksRes.data || [];

  const akses = {};
  matriks.filter((m) => m.peran_id === user.peran_id).forEach((m) => {
    akses[m.kode_modul] = m.level_akses;
  });

  // ---------- Sidebar user info ----------
  const nameEl = qs("#sidebar-user-name");
  const roleEl = qs("#sidebar-user-role");
  const satkerEl = qs("#sidebar-user-satker");
  if (nameEl) nameEl.textContent = user.nama_lengkap;
  if (roleEl) roleEl.textContent = user.peran_nama;
  if (satkerEl) satkerEl.textContent = user.satker_nama;

  // ---------- Sembunyikan nav sesuai modul yang 'ditolak' ----------
  qsa("[data-module]").forEach((el) => {
    const modul = el.getAttribute("data-module");
    if (akses[modul] === "ditolak") {
      el.style.display = "none";
    }
  });

  // ---------- Highlight halaman aktif ----------
  const currentPage = document.body.getAttribute("data-page");
  qsa(".nav-link[data-page]").forEach((el) => {
    if (el.getAttribute("data-page") === currentPage) el.classList.add("active");
  });

  // ---------- Logout ----------
  const logoutBtn = qs("#btn-logout");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", async () => {
      await Api.post("/api/v1/auth/logout", {}, { skipAuthRedirect: true });
      window.location.href = "/index.html";
    });
  }

  // ---------- Mobile sidebar toggle ----------
  const toggleBtn = qs("#btn-toggle-sidebar");
  if (toggleBtn) {
    toggleBtn.addEventListener("click", () => qs(".sidebar").classList.toggle("open"));
  }

  return { user, akses };
})();
