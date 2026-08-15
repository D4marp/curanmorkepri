(async function () {
  const { akses } = await window.AppReady;
  const canEdit = akses["kelola_data"] === "penuh";

  const params = new URLSearchParams(window.location.search);
  const lpId = params.get("id");
  if (!lpId) {
    qs("#content-root").innerHTML = '<div class="alert alert-error">ID laporan tidak ditemukan pada URL.</div>';
    return;
  }

  let lp = null;

  // ---------- Modal generik ----------
  function openModal(title, fieldsHtml, onSubmit) {
    const backdrop = document.createElement("div");
    backdrop.className = "modal-backdrop";
    backdrop.innerHTML = `
      <div class="modal">
        <h3>${escapeHtml(title)}</h3>
        <div id="modal-alert"></div>
        <form id="modal-form">${fieldsHtml}</form>
        <div class="modal-actions">
          <button class="btn btn-outline" id="modal-cancel" type="button">Batal</button>
          <button class="btn btn-primary" id="modal-submit" type="submit" form="modal-form">Simpan</button>
        </div>
      </div>`;
    document.body.appendChild(backdrop);
    backdrop.querySelector("#modal-cancel").onclick = () => backdrop.remove();
    backdrop.querySelector("#modal-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const submitBtn = backdrop.querySelector("#modal-submit");
      submitBtn.disabled = true;
      try {
        await onSubmit(backdrop);
        backdrop.remove();
        await loadAndRender();
      } catch (err) {
        backdrop.querySelector("#modal-alert").innerHTML = '<div class="alert alert-error">' + escapeHtml(err.message) + "</div>";
        submitBtn.disabled = false;
      }
    });
    return backdrop;
  }
  function val(root, id) { return root.querySelector("#" + id).value; }

  // ---------- Load & render ----------
  async function loadAndRender() {
    const res = await Api.get("/api/v1/laporan/" + lpId);
    lp = res.data;
    render();
  }

  function render() {
    document.title = lp.no_lp + " — CURANMOR AI";
    qs("#title-no-lp").textContent = lp.no_lp;

    const root = qs("#content-root");
    root.innerHTML = "";
    const tpl = qs("#tpl-content").content.cloneNode(true);
    root.appendChild(tpl);

    qsa("[data-f]", root).forEach((el) => {
      const key = el.getAttribute("data-f");
      let v = lp[key];
      if (key === "status_kasus") { el.innerHTML = '<span class="' + statusBadgeClass(v) + '">' + escapeHtml(v) + "</span>"; return; }
      if (key === "tanggal_lp" || key === "tanggal_kejadian") v = fmtDate(v);
      el.textContent = v || "-";
    });

    // Status actions (dropdown ubah status)
    const statusActions = qs("#status-actions");
    if (akses["kelola_data"] === "penuh") {
      statusActions.innerHTML = `
        <select id="sel-status" class="btn btn-outline btn-sm" style="padding:8px 10px">
          <option value="">Ubah status ke...</option>
          <option value="LP">LP</option><option value="SP2HP">SP2HP</option>
          <option value="DPO">DPO</option><option value="Selesai">Selesai</option>
        </select>`;
      qs("#sel-status", statusActions).addEventListener("change", async (e) => {
        const statusBaru = e.target.value;
        if (!statusBaru) return;
        const keterangan = window.prompt("Keterangan perubahan status (opsional):", "") || "";
        try {
          await Api.patch(`/api/v1/laporan/${lpId}/status`, { status_baru: statusBaru, keterangan });
          await loadAndRender();
        } catch (err) {
          alert(err.message);
        }
      });
    }

    renderTabs(root);
    renderKendaraan(root);
    renderPihak(root);
    renderBB(root);
    renderDok(root);
    renderPenyidikan(root);
    renderStatus(root);
    wireAddButtons(root);

    qsa("[data-module]", root).forEach((el) => {
      if (akses[el.getAttribute("data-module")] === "ditolak" || (!canEdit && el.tagName !== "DIV")) {
        if (akses[el.getAttribute("data-module")] === "ditolak") el.style.display = "none";
      }
    });
  }

  function renderTabs(root) {
    qsa(".tab-btn", root).forEach((btn) => {
      btn.addEventListener("click", () => {
        qsa(".tab-btn", root).forEach((b) => b.classList.remove("active"));
        qsa(".tab-panel", root).forEach((p) => p.classList.remove("active"));
        btn.classList.add("active");
        qs('.tab-panel[data-panel="' + btn.dataset.tab + '"]', root).classList.add("active");
      });
    });
  }

  function actionCell(onDelete) {
    if (!canEdit) return "";
    return `<button class="btn btn-sm btn-danger btn-del">Hapus</button>`;
  }

  function renderKendaraan(root) {
    const tbody = qs("#tbl-kendaraan", root);
    const list = lp.kendaraan || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="6" class="empty-state">Belum ada data kendaraan.</td></tr>'
      : list.map((k) => `
        <tr data-id="${k.id}">
          <td>${escapeHtml(k.no_polisi || "-")}</td>
          <td>${escapeHtml(k.no_rangka_vin || "-")}</td>
          <td>${escapeHtml(k.no_mesin || "-")}</td>
          <td>${escapeHtml(k.merk_tipe || "-")}</td>
          <td>
            ${canEdit ? `<select class="level-select sel-status-kendaraan">
              ${["Terlapor Hilang","Ditemukan","Diamankan","Dikembalikan"].map(s => `<option ${s===k.status_kendaraan?"selected":""}>${s}</option>`).join("")}
            </select>` : `<span class="${statusBadgeClass(k.status_kendaraan)}">${escapeHtml(k.status_kendaraan)}</span>`}
          </td>
          <td>${actionCell()}</td>
        </tr>`).join("");

    qsa(".sel-status-kendaraan", tbody).forEach((sel) => {
      sel.addEventListener("change", async (e) => {
        const id = e.target.closest("tr").dataset.id;
        await Api.patch(`/api/v1/kendaraan/${id}/status`, { status_kendaraan: e.target.value });
        await loadAndRender();
      });
    });
    qsa(".btn-del", tbody).forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        if (!confirm("Hapus data kendaraan ini?")) return;
        const id = e.target.closest("tr").dataset.id;
        await Api.del(`/api/v1/kendaraan/${id}`);
        await loadAndRender();
      });
    });
  }

  function renderPihak(root) {
    const tbody = qs("#tbl-pihak", root);
    const list = lp.pihak_terkait || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="6" class="empty-state">Belum ada data pihak terkait.</td></tr>'
      : list.map((p) => `
        <tr data-id="${p.id}">
          <td>${escapeHtml(p.jenis_pihak)}</td>
          <td>${escapeHtml(p.nama)}</td>
          <td>${escapeHtml(p.nik_masked || "-")}</td>
          <td>${escapeHtml(p.no_telp || "-")}</td>
          <td>${escapeHtml(p.alamat || "-")}</td>
          <td>${actionCell()}</td>
        </tr>`).join("");
    qsa(".btn-del", tbody).forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        if (!confirm("Hapus pihak terkait ini?")) return;
        const id = e.target.closest("tr").dataset.id;
        await Api.del(`/api/v1/pihak-terkait/${id}`);
        await loadAndRender();
      });
    });
  }

  function renderBB(root) {
    const tbody = qs("#tbl-bb", root);
    const list = lp.barang_bukti || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="5" class="empty-state">Belum ada barang bukti.</td></tr>'
      : list.map((b) => `
        <tr data-id="${b.id}">
          <td>${escapeHtml(b.jenis_bb || "-")}</td>
          <td>${escapeHtml(b.no_registrasi_bb || "-")}</td>
          <td>${escapeHtml(b.lokasi_penyimpanan || "-")}</td>
          <td>${fmtDate(b.tanggal_diamankan)}</td>
          <td>${actionCell()}</td>
        </tr>`).join("");
    qsa(".btn-del", tbody).forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        if (!confirm("Hapus barang bukti ini?")) return;
        const id = e.target.closest("tr").dataset.id;
        await Api.del(`/api/v1/barang-bukti/${id}`);
        await loadAndRender();
      });
    });
  }

  function renderDok(root) {
    const tbody = qs("#tbl-dok", root);
    const list = lp.dokumentasi || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="4" class="empty-state">Belum ada dokumentasi.</td></tr>'
      : list.map((d) => `
        <tr data-id="${d.id}">
          <td>${escapeHtml(d.jenis_dokumen)}</td>
          <td><a href="${d.file_url}" target="_blank" rel="noopener">Lihat berkas</a></td>
          <td>${fmtDateTime(d.waktu_unggah)}</td>
          <td>${actionCell()}</td>
        </tr>`).join("");
    qsa(".btn-del", tbody).forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        if (!confirm("Hapus dokumentasi ini?")) return;
        const id = e.target.closest("tr").dataset.id;
        await Api.del(`/api/v1/dokumentasi/${id}`);
        await loadAndRender();
      });
    });
  }

  function renderPenyidikan(root) {
    const tbody = qs("#tbl-penyidikan", root);
    const list = lp.riwayat_penyidikan || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="4" class="empty-state">Belum ada jurnal penyidikan.</td></tr>'
      : list.map((r) => `
        <tr data-id="${r.id}">
          <td>${fmtDate(r.tanggal)}</td>
          <td>${escapeHtml(r.kegiatan || "-")}</td>
          <td>${escapeHtml(r.catatan || "-")}</td>
          <td>${actionCell()}</td>
        </tr>`).join("");
    qsa(".btn-del", tbody).forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        if (!confirm("Hapus entri jurnal ini?")) return;
        const id = e.target.closest("tr").dataset.id;
        await Api.del(`/api/v1/riwayat-penyidikan/${id}`);
        await loadAndRender();
      });
    });
  }

  function renderStatus(root) {
    const tbody = qs("#tbl-status", root);
    const list = lp.riwayat_status || [];
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="5" class="empty-state">Belum ada perubahan status.</td></tr>'
      : list.map((r) => `
        <tr>
          <td>${fmtDateTime(r.tanggal_perubahan)}</td>
          <td>${escapeHtml(r.status_sebelumnya || "-")}</td>
          <td><span class="${statusBadgeClass(r.status_baru)}">${escapeHtml(r.status_baru)}</span></td>
          <td>${escapeHtml(r.keterangan || "-")}</td>
          <td>${escapeHtml(r.diubah_oleh_nama || "-")}</td>
        </tr>`).join("");
  }

  function wireAddButtons(root) {
    qs("#btn-add-kendaraan", root)?.addEventListener("click", () => {
      openModal("Tambah Kendaraan", `
        <div class="field"><label>No. Polisi</label><input id="m_no_polisi"></div>
        <div class="field"><label>No. Rangka (VIN)</label><input id="m_no_rangka"></div>
        <div class="field"><label>No. Mesin</label><input id="m_no_mesin"></div>
        <div class="field"><label>Merk / Tipe</label><input id="m_merk"></div>
        <div class="field"><label>Warna</label><input id="m_warna"></div>
        <div class="field"><label>Tahun</label><input id="m_tahun" type="number" min="1900" max="2100"></div>
        <div class="field"><label>Jenis Kendaraan</label><select id="m_jenis"><option>Roda 2</option><option>Roda 4</option><option>Lainnya</option></select></div>
      `, async (m) => {
        await Api.post(`/api/v1/laporan/${lpId}/kendaraan`, {
          no_polisi: val(m, "m_no_polisi"), no_rangka_vin: val(m, "m_no_rangka"), no_mesin: val(m, "m_no_mesin"),
          merk_tipe: val(m, "m_merk"), warna: val(m, "m_warna"),
          tahun: val(m, "m_tahun") ? Number(val(m, "m_tahun")) : null,
          jenis_kendaraan: val(m, "m_jenis"), status_kendaraan: "Terlapor Hilang",
        });
      });
    });

    qs("#btn-add-pihak", root)?.addEventListener("click", () => {
      openModal("Tambah Pihak Terkait", `
        <div class="field"><label>Jenis Pihak</label><select id="m_jenis"><option>Pelapor</option><option>Terlapor</option><option>Saksi</option></select></div>
        <div class="field"><label>Nama</label><input id="m_nama" required></div>
        <div class="field"><label>NIK <span class="helper-text">(dienkripsi at-rest, tidak ditampilkan utuh)</span></label><input id="m_nik" maxlength="20"></div>
        <div class="field"><label>No. Telp</label><input id="m_telp"></div>
        <div class="field"><label>Alamat</label><textarea id="m_alamat"></textarea></div>
      `, async (m) => {
        await Api.post(`/api/v1/laporan/${lpId}/pihak-terkait`, {
          jenis_pihak: val(m, "m_jenis"), nama: val(m, "m_nama"), nik: val(m, "m_nik"),
          no_telp: val(m, "m_telp"), alamat: val(m, "m_alamat"),
        });
      });
    });

    qs("#btn-add-bb", root)?.addEventListener("click", () => {
      openModal("Tambah Barang Bukti", `
        <div class="field"><label>Jenis Barang Bukti</label><input id="m_jenis" placeholder="mis. STNK asli, kunci kontak"></div>
        <div class="field"><label>No. Registrasi BB</label><input id="m_noreg"></div>
        <div class="field"><label>Lokasi Penyimpanan</label><input id="m_lokasi" placeholder="mis. Gudang Barang Bukti Polres"></div>
        <div class="field"><label>Tanggal Diamankan</label><input id="m_tgl" type="date"></div>
        <div class="field"><label>Deskripsi</label><textarea id="m_desk"></textarea></div>
      `, async (m) => {
        await Api.post(`/api/v1/laporan/${lpId}/barang-bukti`, {
          jenis_bb: val(m, "m_jenis"), no_registrasi_bb: val(m, "m_noreg"),
          lokasi_penyimpanan: val(m, "m_lokasi"), tanggal_diamankan: val(m, "m_tgl") || null,
          deskripsi: val(m, "m_desk"),
        });
      });
    });

    qs("#btn-add-penyidikan", root)?.addEventListener("click", () => {
      openModal("Tambah Jurnal Penyidikan", `
        <div class="field"><label>Tanggal</label><input id="m_tgl" type="date" required></div>
        <div class="field"><label>Kegiatan</label><input id="m_kegiatan" placeholder="mis. pemeriksaan saksi, olah TKP"></div>
        <div class="field"><label>Catatan</label><textarea id="m_catatan"></textarea></div>
      `, async (m) => {
        await Api.post(`/api/v1/laporan/${lpId}/riwayat-penyidikan`, {
          tanggal: val(m, "m_tgl"), kegiatan: val(m, "m_kegiatan"), catatan: val(m, "m_catatan"),
        });
      });
    });

    qs("#btn-add-dok", root)?.addEventListener("click", async () => {
      const fileInput = qs("#dok-file", root);
      const jenis = qs("#dok-jenis", root).value;
      if (!fileInput.files.length) { alert("Pilih berkas terlebih dahulu."); return; }
      const btn = qs("#btn-add-dok", root);
      btn.disabled = true;
      btn.textContent = "Mengunggah...";
      try {
        const url = await Api.upload(fileInput.files[0]);
        await Api.post(`/api/v1/laporan/${lpId}/dokumentasi`, { jenis_dokumen: jenis, file_url: url });
        await loadAndRender();
      } catch (err) {
        alert(err.message);
      } finally {
        btn.disabled = false;
        btn.textContent = "Unggah";
      }
    });
  }

  await loadAndRender();
})();
