// Widget "Import Massal Laporan" — dipasang di halaman mana pun yang punya
// elemen <div id="bulk-import-mount">. Satu sumber markup+logika supaya
// Dashboard, Data & Pencarian Kasus, dan Input Laporan Baru selalu konsisten
// tanpa duplikasi salinan HTML/JS yang bisa saling menyimpang saat diubah.
(async function () {
  await window.AppReady;

  const mount = qs("#bulk-import-mount");
  if (!mount) return; // halaman ini tidak menyediakan titik pasang widget

  mount.innerHTML = `
    <div class="card">
      <h2>Import Massal Laporan <span class="hint">— unggah banyak laporan sekaligus lewat berkas Excel (.xlsx)</span></h2>
      <p class="helper-text">
        Unduh template dulu (kolom sudah benar, ada dropdown validasi, dan sheet <strong>Petunjuk</strong>
        berisi daftar kode satuan kerja terbaru). Satu baris pada sheet "Data" = satu laporan
        (kolom kendaraan &amp; pelapor opsional). Kolom wajib: <code>no_lp</code>, <code>tanggal_lp</code>,
        <code>kode_satker</code>. Baris yang gagal tetap dilaporkan detail alasannya — baris lain yang valid tetap tersimpan.
      </p>
      <div class="toolbar">
        <a class="btn btn-outline" href="/api/v1/laporan/bulk-import/template" download>&#11015; Unduh Template Excel</a>
      </div>
      <form id="form-import" class="toolbar" style="margin-top:10px">
        <input type="file" id="file-input" accept=".xlsx" required>
        <button type="submit" class="btn btn-primary" id="btn-submit-import">Unggah &amp; Proses</button>
      </form>
      <div id="import-alert-box"></div>

      <div id="result-card" style="display:none; margin-top:14px">
        <div class="stat-grid" id="result-summary"></div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Baris</th>
                <th>No. LP</th>
                <th>Status</th>
                <th>Kendaraan</th>
                <th>Pelapor</th>
                <th>Keterangan</th>
              </tr>
            </thead>
            <tbody id="result-rows"></tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  qs("#form-import").addEventListener("submit", async (e) => {
    e.preventDefault();
    const alertBox = qs("#import-alert-box");
    const btn = qs("#btn-submit-import");
    const fileInput = qs("#file-input");
    const file = fileInput.files[0];

    alertBox.innerHTML = "";
    qs("#result-card").style.display = "none";

    if (!file) {
      alertBox.innerHTML = '<div class="alert alert-error">Pilih berkas .xlsx terlebih dahulu</div>';
      return;
    }

    btn.disabled = true;
    btn.textContent = "Memproses...";

    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await Api.post("/api/v1/laporan/bulk-import", fd);
      renderResult(res.data);
      alertBox.innerHTML =
        '<div class="alert alert-success">Import selesai — ' +
        res.data.berhasil + " berhasil, " + res.data.gagal + " gagal, " + res.data.dilewati + " dilewati.</div>";
      fileInput.value = "";
    } catch (err) {
      alertBox.innerHTML = '<div class="alert alert-error">' + escapeHtml(err.message) + "</div>";
    } finally {
      btn.disabled = false;
      btn.textContent = "Unggah & Proses";
    }
  });

  function renderResult(data) {
    const summary = qs("#result-summary");
    summary.innerHTML =
      tile("Total Baris", data.total_baris, "") +
      tile("Berhasil", data.berhasil, "accent-green") +
      tile("Gagal", data.gagal, "accent-red") +
      tile("Dilewati", data.dilewati, "accent-gold");

    const rows = data.hasil || [];
    const tbody = qs("#result-rows");
    tbody.innerHTML = rows.map((r) => `
      <tr>
        <td>${r.baris}</td>
        <td>${escapeHtml(r.no_lp || "-")}</td>
        <td><span class="${statusPillClass(r.status)}">${escapeHtml(r.status)}</span></td>
        <td>${r.kendaraan_dibuat ? "&#10003;" : "-"}</td>
        <td>${r.pelapor_dibuat ? "&#10003;" : "-"}</td>
        <td>${escapeHtml(r.pesan || "-")}</td>
      </tr>
    `).join("");

    qs("#result-card").style.display = "";
  }

  function tile(label, value, accent) {
    return `<div class="stat-tile ${accent}"><div class="label">${label}</div><div class="value">${value}</div></div>`;
  }

  function statusPillClass(status) {
    if (status === "berhasil") return "badge badge-selesai";
    if (status === "gagal") return "badge badge-dpo";
    return "badge badge-lp";
  }
})();
