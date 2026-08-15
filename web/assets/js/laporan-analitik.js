(async function () {
  await window.AppReady;
  wireTabs(document);

  async function loadCharts() {
    const [trenRes, katRes] = await Promise.all([
      Api.get("/api/v1/dashboard/tren-bulanan?bulan=12"),
      Api.get("/api/v1/dashboard/kategori-kasus"),
    ]);
    const tren = trenRes.data || [];
    new Chart(qs("#chart-tren").getContext("2d"), {
      type: "bar",
      data: { labels: tren.map((d) => d.bulan), datasets: [{ label: "Jumlah Kasus", data: tren.map((d) => d.jumlah_kasus), backgroundColor: "#274b7d" }] },
      options: { plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true, ticks: { precision: 0 } } } },
    });
    const kat = katRes.data || [];
    new Chart(qs("#chart-kategori").getContext("2d"), {
      type: "pie",
      data: { labels: kat.map((d) => d.jenis_perkara), datasets: [{ data: kat.map((d) => d.jumlah), backgroundColor: ["#274b7d", "#d4a017", "#b3261e", "#1e7d47", "#8492a6"] }] },
      options: { plugins: { legend: { position: "bottom", labels: { boxWidth: 12, font: { size: 11 } } } } },
    });
  }

  let satkerList = [];
  async function loadPeriodik() {
    const [satkerRes, res] = await Promise.all([Api.get("/api/v1/satuan-kerja"), Api.get("/api/v1/laporan-periodik")]);
    satkerList = satkerRes.data || [];
    const list = res.data || [];
    qs("#tbl-periodik").innerHTML = list.length === 0
      ? '<tr><td colspan="5" class="empty-state">Belum ada laporan periodik.</td></tr>'
      : list.map((l) => `
        <tr>
          <td>${escapeHtml(l.jenis_laporan)}</td>
          <td>${fmtDate(l.periode_mulai)} — ${fmtDate(l.periode_selesai)}</td>
          <td>${l.satker_id ? escapeHtml((satkerList.find(s => s.id === l.satker_id) || {}).nama_satker || "-") : "Seluruh Polda"}</td>
          <td>${l.file_url ? `<a href="${l.file_url}" target="_blank">Unduh</a>` : "-"}</td>
          <td>${fmtDateTime(l.created_at)}</td>
        </tr>`).join("");
  }

  function openNewPeriodik() {
    openModal("Buat Entri Laporan Periodik", `
      <div class="field"><label>Jenis Laporan</label><select id="m_jenis"><option>Harian</option><option>Mingguan</option><option>Bulanan</option></select></div>
      <div class="field"><label>Periode Mulai</label><input id="m_mulai" type="date" required></div>
      <div class="field"><label>Periode Selesai</label><input id="m_selesai" type="date" required></div>
      <div class="field"><label>Satuan Kerja (kosongkan untuk seluruh Polda)</label>
        <select id="m_satker"><option value="">— Seluruh Polda —</option>${satkerList.map(s => `<option value="${s.id}">${escapeHtml(s.nama_satker)}</option>`).join("")}</select>
      </div>
      <div class="field"><label>URL Berkas (opsional)</label><input id="m_file" placeholder="/uploads/laporan-xxx.pdf"></div>
    `, async (m) => {
      await Api.post("/api/v1/laporan-periodik", {
        jenis_laporan: modalVal(m, "m_jenis"), periode_mulai: modalVal(m, "m_mulai"), periode_selesai: modalVal(m, "m_selesai"),
        satker_id: modalVal(m, "m_satker") ? Number(modalVal(m, "m_satker")) : null, file_url: modalVal(m, "m_file"),
      });
      await loadPeriodik();
    });
  }

  let auditPage = 1;
  async function loadAudit(page) {
    auditPage = page || 1;
    const res = await Api.get(`/api/v1/audit-log?page=${auditPage}&page_size=50`);
    const list = res.data || [];
    const total = (res.meta && res.meta.total) || 0;
    qs("#tbl-audit").innerHTML = list.length === 0
      ? '<tr><td colspan="5" class="empty-state">Belum ada aktivitas tercatat.</td></tr>'
      : list.map((a) => `
        <tr>
          <td>${fmtDateTime(a.waktu)}</td>
          <td>${escapeHtml(a.pengguna_nama || "-")}</td>
          <td>${escapeHtml(a.aktivitas)}</td>
          <td>${escapeHtml(a.modul || "-")}</td>
          <td>${escapeHtml(a.ip_address || "-")}</td>
        </tr>`).join("");

    const totalPages = Math.max(1, Math.ceil(total / 50));
    const pag = qs("#audit-pagination");
    pag.innerHTML = "";
    if (totalPages > 1) {
      const prev = document.createElement("button");
      prev.className = "btn btn-sm btn-outline"; prev.textContent = "‹ Sebelumnya";
      prev.disabled = auditPage <= 1; prev.onclick = () => loadAudit(auditPage - 1);
      const next = document.createElement("button");
      next.className = "btn btn-sm btn-outline"; next.textContent = "Berikutnya ›";
      next.disabled = auditPage >= totalPages; next.onclick = () => loadAudit(auditPage + 1);
      const info = document.createElement("span");
      info.textContent = `Halaman ${auditPage} dari ${totalPages}`;
      pag.append(prev, info, next);
    }
  }

  await loadCharts();
  await loadPeriodik();
  await loadAudit(1);

  qs("#btn-new-periodik").addEventListener("click", openNewPeriodik);
})();
