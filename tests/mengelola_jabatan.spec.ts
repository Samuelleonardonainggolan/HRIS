import { test, expect, Page } from "@playwright/test";

// ─── Helper: masuk ke halaman daftar jabatan melalui departemen pertama ────────
// Karena halaman jabatan membutuhkan ?departmentId, kita navigasi dari
// halaman departemen dengan mengklik salah satu baris departemen.
async function navigateToJabatanList(page: Page) {
  await page.goto("/dashboard/manager-hr/departemen");

  // Tunggu tabel departemen dimuat
  await page.waitForTimeout(1000);

  // Klik baris pertama departemen untuk membuka halaman jabatan
  await page.locator("tbody tr").first().click();

  // Verifikasi sudah masuk ke halaman jabatan dengan query departmentId
  await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jabatan\?departmentId=/);

  // Tunggu tabel jabatan dimuat
  await page.waitForTimeout(1000);
}

test.describe("Modul Mengelola Data Jabatan", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await page.getByRole("button", { name: "Manager HR" }).click();
    await expect(page).toHaveURL(/dashboard/);
  });

  test("Menambah jabatan baru", async ({ page }) => {
    await navigateToJabatanList(page);

    // Klik tombol Tambah Posisi
    await page.getByRole("button", { name: "Tambah Posisi" }).click();
    await expect(page).toHaveURL(/tambah-jabatan/);

    // Verifikasi halaman tambah terbuka
    await expect(page.getByRole("heading", { name: "Tambah Jabatan Baru" })).toBeVisible();

    // Isi Nama Jabatan
    const randomSuffix = Math.floor(Math.random() * 10000);
    await page.locator("#name").fill(`Jabatan Test ${randomSuffix}`);

    // Jika departemen tidak terkunci (tidak ada query ?departmentId), pilih dari dropdown
    const isDeptLocked = await page.getByText("Terkunci").isVisible();
    if (!isDeptLocked) {
      await page.getByRole("combobox").filter({ hasText: /Pilih Departemen/ }).click();
      await page.getByRole("option").first().click();
    }

    // Pilih Level Jabatan
    await page.locator("#level").click();
    await page.getByRole("option", { name: /Level 1/ }).click();

    // Isi Deskripsi
    await page.locator("#description").fill("Deskripsi jabatan untuk keperluan pengujian otomatis.");

    // Simpan jabatan
    await page.getByRole("button", { name: "Simpan Jabatan" }).click();

    // Verifikasi kembali ke halaman daftar jabatan
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jabatan\?departmentId=/);
    await expect(page.getByText("Jabatan berhasil ditambahkan").first()).toBeVisible();
  });

  test("Mengubah jabatan", async ({ page }) => {
    await navigateToJabatanList(page);

    // Klik tombol "Edit" pada jabatan pertama di tabel
    await page.getByRole("button", { name: "Edit" }).first().click();

    // Pastikan halaman edit terbuka (URL mengandung ?edit=)
    await expect(page).toHaveURL(/tambah-jabatan\?edit=/);
    await expect(page.getByRole("heading", { name: "Edit Jabatan" })).toBeVisible();

    // Ubah deskripsi jabatan
    await page.locator("#description").fill("Deskripsi jabatan diperbarui melalui pengujian otomatis.");

    // Simpan perubahan
    await page.getByRole("button", { name: "Simpan Perubahan" }).click();

    // Verifikasi berhasil (router.back() → kembali ke daftar jabatan)
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jabatan\?departmentId=/);
    await expect(page.getByText("Jabatan berhasil diperbarui").first()).toBeVisible();
  });

  test("Menonaktifkan jabatan", async ({ page }) => {
    // Intercept dialog browser (konfirmasi window.confirm)
    page.on("dialog", async (dialog) => {
      // Otomatis klik OK pada semua dialog konfirmasi
      await dialog.accept();
    });

    await navigateToJabatanList(page);

    // Klik tombol "Nonaktifkan" pada jabatan pertama yang berstatus Aktif
    const nonaktifkanButton = page.getByRole("button", { name: "Nonaktifkan" }).first();
    await nonaktifkanButton.click();

    // Verifikasi pesan sukses muncul
    await expect(page.getByText("Jabatan berhasil dinonaktifkan").first()).toBeVisible();
  });
});
