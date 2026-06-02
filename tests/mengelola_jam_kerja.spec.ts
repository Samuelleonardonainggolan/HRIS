import { test, expect } from "@playwright/test";

test.describe("Modul Mengelola Jam Kerja Karyawan", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    // Gunakan Quick Login sebagai Manager HR
    await page.getByRole("button", { name: "Manager HR" }).click();
    await expect(page).toHaveURL(/dashboard/);

    // Navigasi ke halaman Manajemen Jam Kerja
    await page.goto("/dashboard/manager-hr/jam-kerja");
  });

  test("Menambah jam kerja karyawan baru", async ({ page }) => {
    // Klik tombol Tambah
    await page.getByRole("button", { name: "Tambah" }).click();
    await expect(page).toHaveURL(/tambah-jam-kerja/);

    // Cari karyawan menggunakan input autocomplete
    const searchInput = page.getByPlaceholder(/Ketik minimal 2 huruf/i);
    await searchInput.fill("ka"); // Harus minimal 2 huruf sesuai instruksi komponen

    // Tunggu proses debounce dan API call selesai
    await page.waitForTimeout(1000);

    // Cek apakah ada pesan tidak ada karyawan yang tersedia
    const noResult = page.getByText(/Tidak ada karyawan yang cocok/i);
    if (await noResult.isVisible()) {
      test.skip(); // Lewati test jika semua karyawan sudah punya jam kerja
      return;
    }

    // Jika ada hasil, klik tombol saran karyawan pertama
    await page.locator('.absolute.z-20 button').first().click();

    // Pilih hari kerja: Senin sampai Jumat
    await page.getByRole("button", { name: "Senin" }).click();
    await page.getByRole("button", { name: "Selasa" }).click();
    await page.getByRole("button", { name: "Rabu" }).click();
    await page.getByRole("button", { name: "Kamis" }).click();
    await page.getByRole("button", { name: "Jumat" }).click();

    // Isi Waktu Mulai dan Waktu Selesai
    const timeInputs = page.locator('input[type="time"]');
    await timeInputs.nth(0).fill("08:00");
    await timeInputs.nth(1).fill("17:00");

    // Simpan Jam Kerja
    await page.getByRole("button", { name: "Simpan Jam Kerja" }).click();

    // Verifikasi kembali ke halaman list jam kerja
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jam-kerja/);
  });

  test("Mengubah jam kerja karyawan", async ({ page }) => {
    // Tunggu sebentar hingga data tabel dimuat
    await page.waitForTimeout(1000);
    
    // Klik tombol 'Atur Jam Kerja' pada karyawan pertama di tabel
    const aturButton = page.getByRole("button", { name: "Atur Jam Kerja" }).first();
    await aturButton.click();

    // Pastikan masuk ke halaman pengaturan individu
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jam-kerja\/.+/);

    // Ubah Waktu Mulai (misal menjadi 09:00) dan Selesai (18:00)
    const timeInputs = page.locator('input[type="time"]');
    await timeInputs.nth(0).fill("09:00");
    await timeInputs.nth(1).fill("18:00");

    // Simpan perubahan
    await page.getByRole("button", { name: "Simpan" }).click();

    // Verifikasi kembali ke halaman list
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jam-kerja/);
  });

  test("Menonaktifkan jam kerja karyawan", async ({ page }) => {
    // Tunggu sebentar hingga data tabel dimuat
    await page.waitForTimeout(1000);
    
    // Klik tombol 'Atur Jam Kerja' pada karyawan pertama
    const aturButton = page.getByRole("button", { name: "Atur Jam Kerja" }).first();
    await aturButton.click();

    // Pastikan masuk ke halaman pengaturan individu
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jam-kerja\/.+/);

    // Cari checkbox 'Aktif' dan hilangkan centangnya (menjadi nonaktif)
    const checkbox = page.getByRole("checkbox");
    const isChecked = await checkbox.isChecked();
    if (isChecked) {
      await checkbox.uncheck();
    }

    // Simpan perubahan
    await page.getByRole("button", { name: "Simpan" }).click();

    // Verifikasi kembali ke halaman list
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/jam-kerja/);
  });
});
