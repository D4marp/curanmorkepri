(async function () {
  await window.AppReady;
  wireTabs(document);

  // ---------- Satuan Kerja ----------
  let satkerList = [];

  function satkerName(id) {
    const s = satkerList.find((x) => x.id === id);
    return s ? s.nama_satker : "-";
  }

  async function loadSatker() {
    const res = await Api.get("/api/v1/satuan-kerja");
    satkerList = res.data || [];
    const tbody = qs("#tbl-satker");
    tbody.innerHTML = satkerList.map((s) => `
      <tr data-id="${s.id}">
        <td>${escapeHtml(s.kode_satker)}</td>
        <td>${escapeHtml(s.nama_satker)}</td>
        <td>${escapeHtml(s.jenis_satker)}</td>
        <td>${s.induk_id ? escapeHtml(satkerName(s.induk_id)) : "-"}</td>
        <td>${escapeHtml(s.wilayah || "-")}</td>
        <td><button class="btn btn-sm btn-danger btn-del-satker">Hapus</button></td>
      </tr>`).join("");
    qsa(".btn-del-satker", tbody).forEach((btn) => btn.addEventListener("click", async (e) => {
      if (!confirm("Hapus satuan kerja ini? Tindakan ini gagal bila masih ada data terkait.")) return;
      const id = e.target.closest("tr").dataset.id;
      try {
        await Api.del(`/api/v1/satuan-kerja/${id}`);
        await loadSatker();
      } catch (err) { alert(err.message); }
    }));
  }

  function openNewSatker() {
    openModal("Satuan Kerja Baru", `
      <div class="field"><label>Kode Satker</label><input id="m_kode" placeholder="POLSEK-XXX-01" required></div>
      <div class="field"><label>Nama Satuan Kerja</label><input id="m_nama" required></div>
      <div class="field"><label>Jenis</label><select id="m_jenis"><option>Polda</option><option>Polres</option><option selected>Polsek</option></select></div>
      <div class="field"><label>Induk (opsional)</label>
        <select id="m_induk"><option value="">— Tidak ada —</option>${satkerList.map(s => `<option value="${s.id}">${escapeHtml(s.nama_satker)}</option>`).join("")}</select>
      </div>
      <div class="field"><label>Wilayah</label><input id="m_wilayah"></div>
    `, async (m) => {
      await Api.post("/api/v1/satuan-kerja", {
        kode_satker: modalVal(m, "m_kode"), nama_satker: modalVal(m, "m_nama"),
        jenis_satker: modalVal(m, "m_jenis"), induk_id: modalVal(m, "m_induk") ? Number(modalVal(m, "m_induk")) : null,
        wilayah: modalVal(m, "m_wilayah"),
      });
      await loadSatker();
    });
  }

  // ---------- Matriks RBAC ----------
  async function loadRbac() {
    const [peranRes, modulRes, matriksRes] = await Promise.all([
      Api.get("/api/v1/peran"), Api.get("/api/v1/modul-akses"), Api.get("/api/v1/rbac/matriks"),
    ]);
    const peranList = peranRes.data || [];
    const modulList = modulRes.data || [];
    const matriks = matriksRes.data || [];

    const head = qs("#rbac-head");
    head.innerHTML = "<th>Peran</th>" + modulList.map((m) => `<th>${escapeHtml(m.nama_modul)}</th>`).join("");

    const tbody = qs("#tbl-rbac");
    tbody.innerHTML = peranList.map((p) => {
      const cells = modulList.map((m) => {
        const cell = matriks.find((x) => x.peran_id === p.id && x.modul_id === m.id);
        const level = cell ? cell.level_akses : "ditolak";
        return `<td>
          <select class="level-select level-${level}" data-peran="${p.id}" data-modul="${m.id}">
            <option value="penuh" ${level === "penuh" ? "selected" : ""}>Penuh</option>
            <option value="terbatas" ${level === "terbatas" ? "selected" : ""}>Terbatas</option>
            <option value="ditolak" ${level === "ditolak" ? "selected" : ""}>Ditolak</option>
          </select>
        </td>`;
      }).join("");
      return `<tr><td><strong>${escapeHtml(p.nama_peran)}</strong></td>${cells}</tr>`;
    }).join("");

    qsa("select.level-select", tbody).forEach((sel) => {
      sel.addEventListener("change", async (e) => {
        const peranId = Number(e.target.dataset.peran);
        const modulId = Number(e.target.dataset.modul);
        const level = e.target.value;
        e.target.className = "level-select level-" + level;
        try {
          await Api.put("/api/v1/rbac/matriks", { peran_id: peranId, modul_id: modulId, level_akses: level });
        } catch (err) {
          alert(err.message);
        }
      });
    });
  }

  // ---------- API Keys ----------
  async function loadApiKeys() {
    const res = await Api.get("/api/v1/api-keys");
    const list = res.data || [];
    const tbody = qs("#tbl-apikeys");
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="6" class="empty-state">Belum ada API key.</td></tr>'
      : list.map((k) => `
        <tr data-id="${k.id}">
          <td>${escapeHtml(k.nama_layanan)}</td>
          <td><code>${escapeHtml(k.key_prefix)}...</code></td>
          <td>${k.status_aktif ? '<span class="badge badge-selesai">Aktif</span>' : '<span class="badge badge-dpo">Dicabut</span>'}</td>
          <td>${k.last_used_at ? fmtDateTime(k.last_used_at) : "Belum pernah"}</td>
          <td>${fmtDate(k.created_at)}</td>
          <td>${k.status_aktif ? '<button class="btn btn-sm btn-danger btn-revoke">Cabut</button>' : ""}</td>
        </tr>`).join("");
    qsa(".btn-revoke", tbody).forEach((btn) => btn.addEventListener("click", async (e) => {
      if (!confirm("Cabut API key ini? Layanan chatbot yang memakainya akan langsung kehilangan akses.")) return;
      const id = e.target.closest("tr").dataset.id;
      await Api.del(`/api/v1/api-keys/${id}`);
      await loadApiKeys();
    }));
  }

  function openNewApiKey() {
    openModal("Buat API Key Baru", `
      <div class="field"><label>Nama Layanan</label><input id="m_nama" placeholder="mis. Chatbot WA Ditreskrimum" required></div>
    `, async (m) => {
      const res = await Api.post("/api/v1/api-keys", { nama_layanan: modalVal(m, "m_nama") });
      await loadApiKeys();
      showRawKey(res.data.raw_key);
    }, "Buat");
  }

  function showRawKey(rawKey) {
    openModal("API Key Dibuat", `
      <div class="alert alert-info">Simpan kunci ini sekarang — tidak akan ditampilkan lagi setelah dialog ini ditutup.</div>
      <div class="field"><label>API Key</label><input id="m_key" readonly value="${escapeHtml(rawKey)}" style="font-family:monospace"></div>
      <div class="field"><label>Header yang dipakai layanan chatbot</label><input readonly value="X-API-Key: ${escapeHtml(rawKey)}" style="font-family:monospace;font-size:11.5px"></div>
    `, async () => {}, "Sudah Disimpan");
  }

  await loadSatker();
  await loadRbac();
  await loadApiKeys();

  qs("#btn-new-satker").addEventListener("click", openNewSatker);
  qs("#btn-new-apikey").addEventListener("click", openNewApiKey);
})();
