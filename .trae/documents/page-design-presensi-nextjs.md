# Spesifikasi Desain Halaman: Presensi (Desktop-first)

## Global Styles
- Layout grid: container max-width 1200–1440px, padding kiri/kanan 24px, gap antar section 16–24px.
- Warna:
  - Background: #F7F8FA (page), Surface: #FFFFFF (card).
  - Border: #E6E8EC.
  - Primary: #1F6FEB (button/link), Hover: #1A5FD1.
  - Text utama: #111827, teks sekunder: #6B7280.
  - Status badge:
    - Hadir: hijau (bg lembut + teks gelap)
    - Terlambat: kuning/oranye
    - Izin: biru
    - Alpha: merah
- Tipografi: base 14–16px, heading 20–24px, angka KPI 24–32px.
- Komponen tombol: tinggi 36–40px, radius 10–12px, fokus terlihat (outline).
- Tabel: header sticky (opsional), row hover, zebra optional, alignment angka kanan.

## Meta Information
- Title: "Presensi | HRIS"
- Description: "Monitor presensi karyawan dengan filter periode, ringkasan, dan daftar presensi."
- Open Graph:
  - og:title: "Presensi"
  - og:description: sama dengan description
  - og:type: "website"

---

## Halaman: Presensi

### Layout
- Menggunakan CSS Grid untuk struktur halaman utama:
  - Baris 1: Header halaman.
  - Baris 2: KPI Cards (grid 4 kolom di desktop).
  - Baris 3: Filter (card full width).
  - Baris 4: Tabel Ringkasan (card full width).
  - Baris 5: Tabel Presensi + Pagination (card full width).
- Responsif:
  - ≥1200px: KPI 4–6 kolom (sesuai jumlah kartu).
  - 768–1199px: KPI 2 kolom.
  - <768px: KPI 1 kolom (meski fokus desktop-first, tetap tidak pecah layout).

### Page Structure
1. **Header Halaman** (stacked)
2. **Kartu KPI** (grid cards)
3. **Filter Presensi** (toolbar dalam card)
4. **Tabel Ringkasan** (tabel ringkas dalam card)
5. **Tabel Presensi** (tabel utama) + **Pagination** (footer area)

### Sections & Components

#### 1) Header Halaman
- Elemen:
  - Judul: "Presensi"
  - Subjudul kecil: periode default (mis. "Bulan ini") atau teks helper.
  - Breadcrumb (opsional jika ada sistem navigasi global): "HRIS / Presensi".
- Interaksi:
  - Tidak ada aksi utama di header (karena fokus halaman adalah filter & tabel).

#### 2) Kartu KPI (KPI Cards)
- Tujuan: memberi snapshot cepat hasil filter aktif.
- Struktur tiap kartu:
  - Label (mis. "Hadir")
  - Angka besar (count)
  - Delta/teks kecil (opsional) hanya jika tersedia dari data, jika tidak maka dihilangkan.
- Tata letak:
  - Grid, tinggi konsisten, padding 16–20px, border halus.

#### 3) Filter Presensi
- Wadah: Card dengan header kecil "Filter".
- Komponen filter (urut kiri ke kanan, meniru layout toolbar):
  - Date range picker: "Dari" dan "Sampai" (atau komponen range tunggal).
  - Dropdown Unit/Departemen.
  - Dropdown/segmented untuk Status.
  - Input pencarian (placeholder: "Cari nama/NIP...") dengan ikon.
  - Tombol aksi:
    - Primary: "Terapkan" (memuat ulang KPI + tabel)
    - Secondary: "Reset" (kembali ke default)
- Perilaku:
  - Saat loading, tombol disabled dan tampilkan indikator loading di area tabel.
  - Validasi ringan: tanggal mulai ≤ tanggal akhir.

#### 4) Tabel Ringkasan
- Wadah: Card.
- Isi: tabel 2–3 kolom, contoh:
  - Kolom "Kategori" (status/departemen)
  - Kolom "Jumlah"
- Perilaku:
  - Mengikuti filter aktif.
  - Empty state: tampilkan teks "Tidak ada ringkasan untuk filter ini".

#### 5) Tabel Presensi (Tabel Utama)
- Wadah: Card.
- Header card:
  - Judul kecil "Daftar Presensi"
  - Teks jumlah data (mis. "Total 1.240") jika tersedia.
- Kolom inti (disarankan):
  - Tanggal
  - Nama
  - NIP
  - Unit/Departemen
  - Status (badge berwarna)
  - Jam Masuk
  - Jam Keluar
- Interaksi:
  - Sorting pada 2–3 kolom utama (Tanggal, Nama, Status) jika diperlukan.
  - Row hover untuk keterbacaan.
- State:
  - Loading: skeleton rows.
  - Error: panel pesan + tombol "Coba lagi".
  - Empty: "Tidak ada data presensi".

#### 6) Pagination
- Letak: footer tabel (rata kanan), sejajar dengan info halaman.
- Komponen:
  - Tombol Prev/Next
  - Indikator halaman (mis. "Halaman 2 dari 25")
  - (Opsional) page size selector (10/20/50) jika ingin meniru kontrol lengkap.
- Aksesibilitas:
  - Tombol disabled saat di halaman pertama/terakhir.

### Interaction & Motion (ringkas)
- Hover state pada row dan tombol (transisi 150–200ms).
- Fokus keyboard jelas pada input, dropdown, dan pagination.
