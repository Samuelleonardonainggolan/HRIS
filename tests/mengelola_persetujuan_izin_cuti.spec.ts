import { test, expect } from "@playwright/test";

test.describe("Modul Persetujuan Izin dan Cuti", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    // Login menggunakan kredensial asli
    await page.getByLabel("Email").fill("manager.hr@company.com");
    await page.getByLabel("Password").fill("password123");
    await page.getByRole("button", { name: "Masuk" }).click();
    await expect(page).toHaveURL(/dashboard/);
    
    // Navigasi ke halaman Persetujuan Izin dan Cuti
    await page.goto("/dashboard/manager-hr/persetujuan-izin-cuti");
  });

  test("Melihat daftar pengajuan izin dan cuti", async ({ page }) => {
    // Verifikasi judul halaman termuat
    await expect(page.getByRole("heading", { name: /Manajemen Izin & Cuti/i })).toBeVisible();
    
    // Verifikasi tabel pengajuan ditampilkan
    await expect(page.locator("table")).toBeVisible();
  });

  test("Menyetujui pengajuan izin/cuti", async ({ page }) => {
    await page.waitForTimeout(1000); // Tunggu data termuat
    
    // Cek apakah ada data yang masih Pending
    const pendingRow = page.locator("tbody tr").filter({ hasText: "Pending" }).first();
    const isPendingExist = await pendingRow.isVisible().catch(() => false);
    
    if (!isPendingExist) {
      console.log("Tidak ada pengajuan berstatus Pending. Melewati test persetujuan.");
      return; // Skip test jika tidak ada data untuk disetujui
    }

    // Klik baris yang Pending
    await pendingRow.click();

    // Di dalam panel detail, klik tombol Setuju
    const setujuButton = page.getByRole("button", { name: "Setuju" }); 
    await setujuButton.click();

    // Verifikasi ada notifikasi berhasil
    await expect(page.getByText(/Pengajuan disetujui/i).first()).toBeVisible();
  });

  test("Menolak pengajuan izin/cuti dengan alasan", async ({ page }) => {
    await page.waitForTimeout(1000); // Tunggu data termuat

    // Cek apakah ada data yang masih Pending
    const pendingRow = page.locator("tbody tr").filter({ hasText: "Pending" }).first();
    const isPendingExist = await pendingRow.isVisible().catch(() => false);
    
    if (!isPendingExist) {
      console.log("Tidak ada pengajuan berstatus Pending. Melewati test penolakan.");
      return; // Skip test jika tidak ada data
    }

    // Klik baris yang Pending
    await pendingRow.click();

    // Klik tombol Tolak
    const tolakButton = page.getByRole("button", { name: "Tolak" }); 
    await tolakButton.click();

    // Modal penolakan muncul, verifikasi elemen modal
    await expect(page.getByRole("heading", { name: /Tolak Pengajuan/i })).toBeVisible();

    // Isi alasan penolakan
    const reasonInput = page.getByPlaceholder(/Tuliskan alasan penolakan/i);
    await reasonInput.fill("Dokumen pendukung kurang lengkap (Test Playwright).");

    // Submit penolakan
    const submitRejectButton = page.getByRole("button", { name: "Tolak Pengajuan" });
    await submitRejectButton.click();

    // Verifikasi ada notifikasi berhasil
    await expect(page.getByText(/Pengajuan ditolak/i).first()).toBeVisible();
  });
});
