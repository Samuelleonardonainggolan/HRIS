import { test, expect } from "@playwright/test";

test.describe("Modul Mengelola Akun Pengguna", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    // Asumsi menggunakan Quick Login sebagai Manager HR agar memiliki akses penuh ke fitur Manajemen Karyawan
    await page.getByRole("button", { name: "Manager HR" }).click();
    await expect(page).toHaveURL(/dashboard/);
    
    // Navigasi ke halaman Manajemen Pegawai
    await page.goto("/dashboard/manager-hr/karyawan");
  });

  test("Menambah akun pengguna baru", async ({ page }) => {
    // Klik tombol Tambah Pegawai
    await page.getByRole("button", { name: /\+ Tambah Pegawai/i }).click();
    await expect(page).toHaveURL(/tambah-pegawai/);

    // Buat data unik agar tidak duplikat (terutama untuk nama dan email)
    const randomSuffix = Math.floor(Math.random() * 100000);
    const testName = `Karyawan Test ${randomSuffix}`;
    const testEmail = `karyawan${randomSuffix}@company.com`;

    // Tunggu sesaat agar nomor payroll tergenerate otomatis dari backend
    await page.waitForTimeout(1000); 

    // Isi Form Identitas
    await page.locator("#full_name").fill(testName);
    await page.locator("#birth_date").fill("1995-01-01");

    // Select Agama
    await page.getByRole("combobox").filter({ hasText: "Pilih Agama" }).click();
    await page.getByRole("option", { name: "Islam" }).click();

    // Select Pendidikan Terakhir
    await page.getByRole("combobox").filter({ hasText: "Pilih Pendidikan" }).click();
    await page.getByRole("option", { name: "S1" }).click();

    // Isi Tanggal Masuk
    await page.locator("#year_enrolled").fill("2023-01-01");

    // Select Status Kepegawaian
    await page.getByRole("combobox").filter({ hasText: "Pilih Status" }).click();
    await page.getByRole("option", { name: "Tetap" }).click();

    // Select Departemen
    await page.getByRole("combobox").filter({ hasText: /^Pilih Departemen$/ }).click();
    await page.getByRole("option").first().click(); // Pilih departemen pertama

    // Tunggu sesaat agar data jabatan (positions) ter-load berdasarkan departemen
    await page.waitForTimeout(500);

    // Select Jabatan
    await page.getByRole("combobox").filter({ hasText: "Pilih Jabatan" }).click();
    await page.getByRole("option").first().click(); // Pilih jabatan pertama

    // Isi Kontak
    await page.locator("#email").fill(testEmail);
    await page.locator("#phone").fill("081234567890");

    // Select Role
    // Karena nilai defaultnya sudah "staf", teks di dalam combobox adalah "Staf", bukan "Pilih Role"
    await page.getByRole("combobox").filter({ hasText: /^Staf$/ }).click();
    await page.getByRole("option", { name: "Staf", exact: true }).click();

    // Isi Alamat
    await page.locator("#address").fill("Jl. Contoh Alamat Testing No. 123");

    // Submit
    await page.getByRole("button", { name: "Simpan Pegawai" }).click();

    // Verifikasi kembali ke halaman list dan ada notifikasi berhasil
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/karyawan/);
    await expect(page.getByText("Pegawai berhasil dibuat").first()).toBeVisible();
  });

  test("Mengubah akun pengguna", async ({ page }) => {
    // Cari pegawai yang akan diubah
    const searchInput = page.getByPlaceholder("Cari pegawai...");
    await searchInput.fill("Karyawan Test");
    
    // Tunggu hasil pencarian termuat
    await page.waitForTimeout(1000);

    // Pilih baris pertama dari hasil pencarian di tabel untuk memunculkan panel detail
    await page.locator("tbody tr").first().click();

    // Di dalam panel detail, klik tombol Edit
    // Selector disesuaikan dengan teks tombol "Edit Data" di EmployeeDetailPanel
    const editButton = page.getByRole("button", { name: "Edit Data" }); 
    await editButton.click();

    // Pastikan berada di halaman edit
    await expect(page).toHaveURL(/edit-pegawai/);

    // Ubah data (contoh mengubah nomor telepon dan alamat)
    await page.locator("#phone").fill("089999999999");
    await page.locator("#address").fill("Jl. Alamat Baru No. 456");

    // Submit perubahan
    await page.getByRole("button", { name: /Simpan/i }).click();

    // Verifikasi diarahkan kembali ke halaman list karyawan
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/karyawan/);
  });

  test("Menonaktifkan akun pengguna", async ({ page }) => {
    // Cari pegawai
    const searchInput = page.getByPlaceholder("Cari pegawai...");
    await searchInput.fill("Karyawan Test");
    
    await page.waitForTimeout(1000);

    // Temukan baris pertama, dan klik tombol opsi dropdown (icon titik tiga/MoreVertical)
    const rowDropdown = page.locator("tbody tr").first().locator('button[aria-haspopup="menu"]');
    await rowDropdown.click();

    // Klik opsi pada dropdown menu (bisa Nonaktifkan atau Aktifkan tergantung status saat ini)
    await page.getByRole("menuitem").first().click();

    // Tunggu pesan toast yang menunjukkan status berhasil diubah
    await expect(page.getByText(/berhasil di(non)?aktifkan/i).first()).toBeVisible();
  });
});
