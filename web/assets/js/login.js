const params = new URLSearchParams(window.location.search);
if (params.get("expired") === "1") {
  qs("#alert-box").innerHTML = '<div class="alert alert-info">Sesi Anda berakhir, silakan masuk kembali.</div>';
}

qs("#login-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const btn = qs("#btn-submit");
  const alertBox = qs("#alert-box");
  alertBox.innerHTML = "";
  btn.disabled = true;
  btn.textContent = "Memproses...";
  try {
    await Api.post("/api/v1/auth/login", {
      nrp: qs("#nrp").value.trim(),
      password: qs("#password").value,
    }, { skipAuthRedirect: true });
    window.location.href = "/dashboard.html";
  } catch (err) {
    alertBox.innerHTML = '<div class="alert alert-error">' + escapeHtml(err.message) + '</div>';
    btn.disabled = false;
    btn.textContent = "Masuk";
  }
});
