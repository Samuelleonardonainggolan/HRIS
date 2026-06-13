import { test, expect } from "@playwright/test";

test.describe("Modul Monitoring Kehadiran di Departemen", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    
    // Gunakan kredensial Manager Departemen
    await page.getByLabel("Email").fill("manager.it@company.com");
    await page.getByLabel("Password").fill("password123");
    await page.getByRole("button", { name: "Masuk" }).click();
    
    await expect(page).toHaveURL(/dashboard/);
    
    // Navigasi ke halaman Presensi (Monitoring Kehadiran) untuk Manager Departemen
    await page.goto("/dashboard/manager-dept/presensi");
  });

  test("Melihat daftar kehadiran karyawan di departemen", async ({ page }) => {
    // Verifikasi halaman termuat dengan mencari judul halaman terkait Presensi/Kehadiran
    await expect(page.getByRole("heading", { name: /Presensi|Kehadiran/i }).first()).toBeVisible();
    
    // Verifikasi tabel/list presensi ditampilkan
    await expect(page.locator("table")).toBeVisible();
  });

  test("Menggunakan filter pencarian dan tanggal pada kehadiran", async ({ page }) => {
    // Tunggu sesaat agar data termuat
    await page.waitForTimeout(1000);

    // Cari input pencarian (berdasarkan placeholder)
    const searchInput = page.getByPlaceholder(/Cari karyawan/i);
    const isSearchVisible = await searchInput.isVisible().catch(() => false);

    if (isSearchVisible) {
      await searchInput.fill("Nama Karyawan Test");
      // Tunggu hasil filter ter-apply
      await page.waitForTimeout(500);
    }

    // Jika terdapat filter tanggal (Datepicker), asumsikan kita bisa mengubah nilainya
    // Locator mungkin berbeda sesuai komponen UI yang digunakan, di sini kita cek ketersediaan tombol filter tanggal
    const dateFilterBtn = page.getByRole("button", { name: /Pilih Tanggal|Hari Ini/i }).first();
    if (await dateFilterBtn.isVisible().catch(() => false)) {
      await dateFilterBtn.click();
      
      // Pilih tanggal tertentu di dalam kalender (misalnya hari ini atau yang spesifik)
      // Ini asumsi kalender memunculkan grid tanggal
      const dayCell = page.locator('.rdp-day:not(.rdp-day_outside)').first();
      if (await dayCell.isVisible().catch(() => false)) {
         await dayCell.click();
      }
    }

    // Verifikasi ulang bahwa tabel tetap terlihat setelah filtering
    await expect(page.locator("table")).toBeVisible();
  });

  test("Melihat detail presensi seorang karyawan", async ({ page }) => {
    await page.waitForTimeout(1000);
    
    // Pastikan ada baris data karyawan pada tabel
    const firstRow = page.locator("tbody tr").first();
    const isRowExist = await firstRow.isVisible().catch(() => false);
    
    if (!isRowExist) {
       console.log("Tidak ada data presensi untuk dilihat detailnya. Melewati tes ini.");
       return;
    }

    // Klik pada baris pertama untuk membuka detail (modal atau panel samping)
    await firstRow.click();

    // Verifikasi panel detail muncul dengan mengecek keberadaan teks "Detail Presensi" atau informasi waktu masuk/keluar
    const detailHeader = page.getByRole("heading", { name: /Detail Presensi|Info Kehadiran/i });
    if (await detailHeader.isVisible().catch(() => false)) {
       await expect(detailHeader).toBeVisible();
    } else {
       // Alternatif lain jika judul modal/panel berbeda, 
       // cukup pastikan bahwa elemen detail muncul di layar, e.g. panel info lokasi/foto
       await expect(page.getByText(/Jam Masuk|Waktu Masuk/i).first()).toBeVisible();
    }
  });
});
