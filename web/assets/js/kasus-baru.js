(async function () {
  const { user } = await window.AppReady;

  async function loadSatkerOptions() {
    const res = await Api.get("/api/v1/satuan-kerja");
    const sel = qs("#satker_id");
    (res.data || []).forEach((s) => {
      const opt = document.createElement("option");
      opt.value = s.id;
      opt.innerHTML = (s.jenis_satker === "Polsek" ? "&nbsp;&nbsp;— " : "") + escapeHtml(s.nama_satker);
      if (s.id === user.satker_id) opt.selected = true;
      sel.appendChild(opt);
    });
  }
  await loadSatkerOptions();

  qs("#form-lp").addEventListener("submit", async (e) => {
    e.preventDefault();
    const alertBox = qs("#alert-box");
    const btn = qs("#btn-submit");
    alertBox.innerHTML = "";
    btn.disabled = true;
    btn.textContent = "Menyimpan...";

    const payload = {
      no_lp: qs("#no_lp").value.trim(),
      tanggal_lp: qs("#tanggal_lp").value,
      tanggal_kejadian: qs("#tanggal_kejadian").value || null,
      jenis_perkara: qs("#jenis_perkara").value,
      satker_id: Number(qs("#satker_id").value),
      status_kasus: qs("#status_kasus").value,
      tkp_alamat: qs("#tkp_alamat").value.trim(),
      tkp_latitude: qs("#tkp_latitude").value ? Number(qs("#tkp_latitude").value) : null,
      tkp_longitude: qs("#tkp_longitude").value ? Number(qs("#tkp_longitude").value) : null,
    };

    try {
      const lpRes = await Api.post("/api/v1/laporan", payload);
      const lpId = lpRes.data.id;

      const noPolisi = qs("#k_no_polisi").value.trim();
      const noRangka = qs("#k_no_rangka").value.trim();
      const noMesin = qs("#k_no_mesin").value.trim();
      if (noPolisi || noRangka || noMesin) {
        await Api.post(`/api/v1/laporan/${lpId}/kendaraan`, {
          no_polisi: noPolisi, no_rangka_vin: noRangka, no_mesin: noMesin,
          merk_tipe: qs("#k_merk").value.trim(), warna: qs("#k_warna").value.trim(),
          tahun: qs("#k_tahun").value ? Number(qs("#k_tahun").value) : null,
          jenis_kendaraan: qs("#k_jenis").value, status_kendaraan: "Terlapor Hilang",
        });
      }

      window.location.href = "/kasus-detail.html?id=" + lpId;
    } catch (err) {
      alertBox.innerHTML = '<div class="alert alert-error">' + escapeHtml(err.message) + "</div>";
      btn.disabled = false;
      btn.textContent = "Simpan Laporan";
      window.scrollTo(0, 0);
    }
  });
})();
