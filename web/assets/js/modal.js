// Modal generik dipakai di beberapa halaman (kasus-detail, pengguna, pengaturan).
function openModal(title, fieldsHtml, onSubmit, submitLabel) {
  const backdrop = document.createElement("div");
  backdrop.className = "modal-backdrop";
  backdrop.innerHTML = `
    <div class="modal">
      <h3>${escapeHtml(title)}</h3>
      <div id="modal-alert"></div>
      <form id="modal-form">${fieldsHtml}</form>
      <div class="modal-actions">
        <button class="btn btn-outline" id="modal-cancel" type="button">Batal</button>
        <button class="btn btn-primary" id="modal-submit" type="submit" form="modal-form">${submitLabel || "Simpan"}</button>
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
    } catch (err) {
      backdrop.querySelector("#modal-alert").innerHTML = '<div class="alert alert-error">' + escapeHtml(err.message) + "</div>";
      submitBtn.disabled = false;
    }
  });
  return backdrop;
}
function modalVal(root, id) { return root.querySelector("#" + id).value; }

function wireTabs(root) {
  const scope = root || document;
  Array.from(scope.querySelectorAll(".tab-btn")).forEach((btn) => {
    btn.addEventListener("click", () => {
      Array.from(scope.querySelectorAll(".tab-btn")).forEach((b) => b.classList.remove("active"));
      Array.from(scope.querySelectorAll(".tab-panel")).forEach((p) => p.classList.remove("active"));
      btn.classList.add("active");
      scope.querySelector('.tab-panel[data-panel="' + btn.dataset.tab + '"]').classList.add("active");
    });
  });
}
