(async function () {
  await window.AppReady;

  let map, markersLayer, chartTren, chartKategori;

  function initMap() {
    map = L.map("map").setView([0.9, 104.4], 8); // pusat: Kepulauan Riau
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
      maxZoom: 18,
    }).addTo(map);
    markersLayer = L.layerGroup().addTo(map);
  }

  async function loadRingkasan() {
    const res = await Api.get("/api/v1/dashboard/ringkasan");
    const d = res.data;
    const tiles = qsa("#stat-grid .value");
    tiles[0].textContent = d.total_kasus;
    tiles[1].textContent = d.kasus_belum_terungkap;
    tiles[2].textContent = d.kasus_selesai;
    tiles[3].textContent = d.kendaraan_ditemukan;
    tiles[4].textContent = d.kendaraan_belum_ditemukan;
  }

  async function loadPeta(dari, sampai) {
    let path = "/api/v1/dashboard/peta-sebaran";
    const params = new URLSearchParams();
    if (dari) params.set("dari", dari);
    if (sampai) params.set("sampai", sampai);
    if ([...params].length) path += "?" + params.toString();

    const res = await Api.get(path);
    const points = res.data || [];
    markersLayer.clearLayers();
    if (points.length === 0) return;

    const bounds = [];
    points.forEach((p) => {
      const color = p.sudah_terungkap ? "#1e7d47" : "#b3261e";
      const marker = L.circleMarker([p.latitude, p.longitude], {
        radius: 8, color, fillColor: color, fillOpacity: 0.75, weight: 1.5,
      }).bindPopup(
        "<b>" + escapeHtml(p.no_lp) + "</b><br>" +
        escapeHtml(p.nama_satker) + "<br>" +
        escapeHtml(p.tkp_alamat || "-") + "<br>" +
        "Tanggal: " + fmtDate(p.tanggal_kejadian) + "<br>" +
        "Status: " + escapeHtml(p.status_kasus)
      );
      marker.addTo(markersLayer);
      bounds.push([p.latitude, p.longitude]);
    });
    if (bounds.length) map.fitBounds(bounds, { padding: [30, 30], maxZoom: 13 });
  }

  async function loadTren() {
    const res = await Api.get("/api/v1/dashboard/tren-bulanan?bulan=6");
    const data = res.data || [];
    const ctx = qs("#chart-tren").getContext("2d");
    if (chartTren) chartTren.destroy();
    chartTren = new Chart(ctx, {
      type: "line",
      data: {
        labels: data.map((d) => d.bulan),
        datasets: [{
          label: "Jumlah Kasus", data: data.map((d) => d.jumlah_kasus),
          borderColor: "#274b7d", backgroundColor: "rgba(39,75,125,0.12)",
          tension: 0.3, fill: true, pointRadius: 3,
        }],
      },
      options: { plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true, ticks: { precision: 0 } } } },
    });
  }

  async function loadKategori() {
    const res = await Api.get("/api/v1/dashboard/kategori-kasus");
    const data = res.data || [];
    const ctx = qs("#chart-kategori").getContext("2d");
    if (chartKategori) chartKategori.destroy();
    chartKategori = new Chart(ctx, {
      type: "doughnut",
      data: {
        labels: data.map((d) => d.jenis_perkara),
        datasets: [{
          data: data.map((d) => d.jumlah),
          backgroundColor: ["#274b7d", "#d4a017", "#b3261e", "#1e7d47", "#8492a6"],
        }],
      },
      options: { plugins: { legend: { position: "bottom", labels: { boxWidth: 12, font: { size: 11 } } } } },
    });
  }

  initMap();
  await Promise.all([loadRingkasan(), loadPeta(), loadTren(), loadKategori()]);

  qs("#btn-filter-peta").addEventListener("click", () => {
    loadPeta(qs("#peta-dari").value, qs("#peta-sampai").value);
  });
  qs("#btn-reset-peta").addEventListener("click", () => {
    qs("#peta-dari").value = "";
    qs("#peta-sampai").value = "";
    loadPeta();
  });
})();
