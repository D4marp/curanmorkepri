(async function () {
  const { user, akses } = await window.AppReady;
  const canEdit = akses["kelola_pengguna"] === "penuh";
  if (!canEdit) qs("#btn-new-user").style.display = "none";

  wireTabs(document);

  let satkerList = [];
  let peranList = [];
  let userList = [];

  async function loadRefs() {
    const [satkerRes, peranRes] = await Promise.all([
      Api.get("/api/v1/satuan-kerja"),
      Api.get("/api/v1/peran"),
    ]);
    satkerList = satkerRes.data || [];
    peranList = peranRes.data || [];
  }

  function satkerOptions(selectedId) {
    return satkerList.map((s) =>
      `<option value="${s.id}" ${s.id === selectedId ? "selected" : ""}>${s.jenis_satker === "Polsek" ? "&nbsp;&nbsp;— " : ""}${escapeHtml(s.nama_satker)}</option>`
    ).join("");
  }
  function peranOptions(selectedId) {
    return peranList.map((p) =>
      `<option value="${p.id}" ${p.id === selectedId ? "selected" : ""}>${escapeHtml(p.nama_peran)}</option>`
    ).join("");
  }

  async function loadUsers() {
    const res = await Api.get("/api/v1/pengguna");
    userList = res.data || [];
    const tbody = qs("#tbl-users");
    tbody.innerHTML = userList.length === 0
      ? '<tr><td colspan="7" class="empty-state">Belum ada data pengguna.</td></tr>'
      : userList.map((u) => `
        <tr data-id="${u.id}">
          <td>${escapeHtml(u.nrp)}</td>
          <td>${escapeHtml(u.nama_lengkap)}</td>
          <td>${escapeHtml(u.peran_nama)}</td>
          <td>${escapeHtml(u.satker_nama)}</td>
          <td>${escapeHtml(u.no_whatsapp || "-")}</td>
          <td>${u.status_aktif ? '<span class="badge badge-selesai">Aktif</span>' : '<span class="badge badge-dpo">Nonaktif</span>'}</td>
          <td>
            ${canEdit ? `<button class="btn btn-sm btn-outline btn-edit">Ubah</button>
            <button class="btn btn-sm btn-outline btn-reset">Reset Sandi</button>` : ""}
          </td>
        </tr>`).join("");

    if (canEdit) {
      qsa(".btn-edit", tbody).forEach((btn) => btn.addEventListener("click", (e) => {
        const id = Number(e.target.closest("tr").dataset.id);
        openEditUser(userList.find((u) => u.id === id));
      }));
      qsa(".btn-reset", tbody).forEach((btn) => btn.addEventListener("click", (e) => {
        const id = Number(e.target.closest("tr").dataset.id);
        openResetPassword(id);
      }));
    }
  }

  function openNewUser() {
    openModal("Pengguna Baru", `
      <div class="field"><label>NRP</label><input id="m_nrp" required></div>
      <div class="field"><label>Nama Lengkap</label><input id="m_nama" required></div>
      <div class="field"><label>Pangkat</label><input id="m_pangkat"></div>
      <div class="field"><label>Jabatan</label><input id="m_jabatan"></div>
      <div class="field"><label>Peran</label><select id="m_peran">${peranOptions()}</select></div>
      <div class="field"><label>Satuan Kerja</label><select id="m_satker">${satkerOptions(user.satker_id)}</select></div>
      <div class="field"><label>No. WhatsApp</label><input id="m_wa" placeholder="0812xxxxxxx"></div>
      <div class="field"><label>Kata Sandi Awal</label><input id="m_pass" type="password" minlength="10" required></div>
      <div class="helper-text">Minimal 10 karakter. Sampaikan ke pengguna melalui jalur aman.</div>
    `, async (m) => {
      await Api.post("/api/v1/pengguna", {
        nrp: modalVal(m, "m_nrp"), nama_lengkap: modalVal(m, "m_nama"),
        pangkat: modalVal(m, "m_pangkat"), jabatan: modalVal(m, "m_jabatan"),
        peran_id: Number(modalVal(m, "m_peran")), satker_id: Number(modalVal(m, "m_satker")),
        no_whatsapp: modalVal(m, "m_wa") || null, password: modalVal(m, "m_pass"),
      });
      await loadUsers();
    });
  }

  function openEditUser(u) {
    openModal("Ubah Pengguna — " + u.nama_lengkap, `
      <div class="field"><label>Nama Lengkap</label><input id="m_nama" value="${escapeHtml(u.nama_lengkap)}" required></div>
      <div class="field"><label>Pangkat</label><input id="m_pangkat" value="${escapeHtml(u.pangkat || "")}"></div>
      <div class="field"><label>Jabatan</label><input id="m_jabatan" value="${escapeHtml(u.jabatan || "")}"></div>
      <div class="field"><label>Peran</label><select id="m_peran">${peranOptions(u.peran_id)}</select></div>
      <div class="field"><label>Satuan Kerja</label><select id="m_satker">${satkerOptions(u.satker_id)}</select></div>
      <div class="field"><label>No. WhatsApp</label><input id="m_wa" value="${escapeHtml(u.no_whatsapp || "")}"></div>
      <div class="field"><label><input type="checkbox" id="m_aktif" ${u.status_aktif ? "checked" : ""}> Akun Aktif</label></div>
    `, async (m) => {
      await Api.put(`/api/v1/pengguna/${u.id}`, {
        nama_lengkap: modalVal(m, "m_nama"), pangkat: modalVal(m, "m_pangkat"), jabatan: modalVal(m, "m_jabatan"),
        peran_id: Number(modalVal(m, "m_peran")), satker_id: Number(modalVal(m, "m_satker")),
        no_whatsapp: modalVal(m, "m_wa") || null, status_aktif: m.querySelector("#m_aktif").checked,
      });
      await loadUsers();
    });
  }

  function openResetPassword(id) {
    openModal("Reset Kata Sandi", `
      <div class="field"><label>Kata Sandi Baru</label><input id="m_pass" type="password" minlength="10" required></div>
    `, async (m) => {
      await Api.post(`/api/v1/pengguna/${id}/reset-password`, { password_baru: modalVal(m, "m_pass") });
    }, "Reset");
  }

  // ---------- Whitelist WA ----------
  async function loadWhitelist() {
    const res = await Api.get("/api/v1/whitelist-whatsapp");
    const list = res.data || [];
    const tbody = qs("#tbl-whitelist");
    tbody.innerHTML = list.length === 0
      ? '<tr><td colspan="5" class="empty-state">Belum ada nomor terdaftar.</td></tr>'
      : list.map((w) => `
        <tr data-id="${w.id}">
          <td>${escapeHtml(w.no_whatsapp)}</td>
          <td>${escapeHtml((userList.find(u => u.id === w.pengguna_id) || {}).nama_lengkap || ("#" + w.pengguna_id))}</td>
          <td>
            <select class="level-select sel-wl-status">
              <option value="aktif" ${w.status === "aktif" ? "selected" : ""}>Aktif</option>
              <option value="nonaktif" ${w.status === "nonaktif" ? "selected" : ""}>Nonaktif</option>
              <option value="diblokir" ${w.status === "diblokir" ? "selected" : ""}>Diblokir</option>
            </select>
          </td>
          <td>${fmtDate(w.tanggal_registrasi)}</td>
          <td><button class="btn btn-sm btn-danger btn-del-wl">Hapus</button></td>
        </tr>`).join("");

    qsa(".sel-wl-status", tbody).forEach((sel) => sel.addEventListener("change", async (e) => {
      const id = e.target.closest("tr").dataset.id;
      await Api.patch(`/api/v1/whitelist-whatsapp/${id}/status`, { status: e.target.value });
    }));
    qsa(".btn-del-wl", tbody).forEach((btn) => btn.addEventListener("click", async (e) => {
      if (!confirm("Hapus nomor ini dari whitelist?")) return;
      const id = e.target.closest("tr").dataset.id;
      await Api.del(`/api/v1/whitelist-whatsapp/${id}`);
      await loadWhitelist();
    }));
  }

  function openNewWhitelist() {
    openModal("Daftarkan Nomor WhatsApp", `
      <div class="field"><label>No. WhatsApp</label><input id="m_wa" placeholder="0812xxxxxxx" required></div>
      <div class="field"><label>Pengguna</label>
        <select id="m_user">${userList.map(u => `<option value="${u.id}">${escapeHtml(u.nama_lengkap)} (${escapeHtml(u.nrp)})</option>`).join("")}</select>
      </div>
    `, async (m) => {
      await Api.post("/api/v1/whitelist-whatsapp", {
        no_whatsapp: modalVal(m, "m_wa"), pengguna_id: Number(modalVal(m, "m_user")),
      });
      await loadWhitelist();
    });
  }

  await loadRefs();
  await loadUsers();
  await loadWhitelist();

  if (canEdit) {
    qs("#btn-new-user").addEventListener("click", openNewUser);
    qs("#btn-new-whitelist").addEventListener("click", openNewWhitelist);
  } else {
    qs("#btn-new-whitelist").style.display = "none";
  }
})();
