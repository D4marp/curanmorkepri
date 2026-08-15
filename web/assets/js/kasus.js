(async function () {
  await window.AppReady;

  let currentPage = 1;
  const pageSize = 20;

  async function loadSatkerOptions() {
    const res = await Api.get("/api/v1/satuan-kerja");
    const sel = qs("#f-satker");
    (res.data || []).forEach((s) => {
      const opt = document.createElement("option");
      opt.value = s.id;
      opt.textContent = (s.jenis_satker === "Polsek" ? "&nbsp;&nbsp;— " : "") + s.nama_satker;
      opt.innerHTML = (s.jenis_satker === "Polsek" ? "&nbsp;&nbsp;— " : "") + escapeHtml(s.nama_satker);
      sel.appendChild(opt);
    });
  }

  function buildQuery(page) {
    const params = new URLSearchParams();
    const keyword = qs("#f-keyword").value.trim();
    const satker = qs("#f-satker").value;
    const status = qs("#f-status").value;
    const dari = qs("#f-dari").value;
    const sampai = qs("#f-sampai").value;
    if (keyword) params.set("keyword", keyword);
    if (satker) params.set("satker_id", satker);
    if (status) params.set("status", status);
    if (dari) params.set("dari", dari);
    if (sampai) params.set("sampai", sampai);
    params.set("page", page);
    params.set("page_size", pageSize);
    return params.toString();
  }

  async function loadResults(page) {
    currentPage = page || 1;
    qs("#tbl-body").innerHTML = '<tr><td colspan="6" class="empty-state">Memuat data...</td></tr>';
    const res = await Api.get("/api/v1/laporan?" + buildQuery(currentPage));
    const rows = res.data || [];
    const total = (res.meta && res.meta.total) || 0;

    qs("#result-count").textContent = total + " laporan ditemukan";

    if (rows.length === 0) {
      qs("#tbl-body").innerHTML = '<tr><td colspan="6" class="empty-state">Tidak ada data yang cocok dengan filter.</td></tr>';
    } else {
      qs("#tbl-body").innerHTML = rows.map((r) => `
        <tr>
          <td><a href="/kasus-detail.html?id=${r.id}"><strong>${escapeHtml(r.no_lp)}</strong></a></td>
          <td>${fmtDate(r.tanggal_kejadian)}</td>
          <td>${escapeHtml(r.satker_nama)}</td>
          <td>${escapeHtml(r.jenis_perkara)}</td>
          <td><span class="${statusBadgeClass(r.status_kasus)}">${escapeHtml(r.status_kasus)}</span></td>
          <td><a class="btn btn-sm btn-outline" href="/kasus-detail.html?id=${r.id}">Detail</a></td>
        </tr>
      `).join("");
    }

    const totalPages = Math.max(1, Math.ceil(total / pageSize));
    const pag = qs("#pagination");
    pag.innerHTML = "";
    if (totalPages > 1) {
      const prev = document.createElement("button");
      prev.className = "btn btn-sm btn-outline";
      prev.textContent = "‹ Sebelumnya";
      prev.disabled = currentPage <= 1;
      prev.onclick = () => loadResults(currentPage - 1);
      const next = document.createElement("button");
      next.className = "btn btn-sm btn-outline";
      next.textContent = "Berikutnya ›";
      next.disabled = currentPage >= totalPages;
      next.onclick = () => loadResults(currentPage + 1);
      const info = document.createElement("span");
      info.textContent = `Halaman ${currentPage} dari ${totalPages}`;
      pag.append(prev, info, next);
    }
  }

  await loadSatkerOptions();
  await loadResults(1);

  qs("#btn-search").addEventListener("click", () => loadResults(1));
  qs("#btn-reset").addEventListener("click", () => {
    qs("#f-keyword").value = "";
    qs("#f-satker").value = "";
    qs("#f-status").value = "";
    qs("#f-dari").value = "";
    qs("#f-sampai").value = "";
    loadResults(1);
  });
  qs("#f-keyword").addEventListener("keydown", (e) => { if (e.key === "Enter") loadResults(1); });
})();
