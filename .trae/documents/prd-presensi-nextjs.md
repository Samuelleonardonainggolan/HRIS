## 1. Product Overview

Halaman **Presensi (Attendance)** untuk meninjau data kehadiran karyawan dalam bentuk KPI, filter, tabel ringkasan, tabel data, dan pagination.
Fokus utamanya adalah mempercepat monitoring dan pencarian data presensi berdasarkan periode dan atribut karyawan.

## 2. Core Features

### 2.1 User Roles

| Role     | Registration Method                  | Core Permissions                                                           |
| -------- | ------------------------------------ | -------------------------------------------------------------------------- |
| HR/Admin | Sudah terdaftar di sistem perusahaan | Melihat seluruh data presensi, melakukan filter, dan navigasi halaman data |
| Karyawan | Sudah terdaftar di sistem perusahaan | Melihat data presensi dirinya (tampilan dan filter tetap sama)             |

### 2.2 Feature Module

Kebutuhan halaman terdiri dari:

1. **Halaman Presensi**: kartu KPI, area filter, tabel ringkasan, tabel presensi, pagination.

### 2.3 Page Details

| Page Name | Module Name              | Feature description                                                                                                                            |
| --------- | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Presensi  | Header & konteks halaman | Menampilkan judul halaman, breadcrumb (jika ada), dan informasi periode aktif (ringkas).                                                       |
| Presensi  | Kartu KPI                | Menampilkan 3–6 metrik ringkas (mis. Hadir, Terlambat, Izin, Alpha) untuk periode/filter aktif.                                                |
| Presensi  | Filter                   | Memfilter data dengan: rentang tanggal, unit/departemen, status presensi, dan pencarian teks (nama/NIP). Menyediakan aksi Terapkan dan Reset.  |
| Presensi  | Tabel ringkasan          | Menampilkan ringkasan agregat per kategori (mis. per status atau per unit) sesuai filter aktif.                                                |
| Presensi  | Tabel presensi           | Menampilkan daftar presensi (kolom inti seperti tanggal, nama, NIP, unit, status, jam masuk/keluar). Mendukung sorting dasar pada kolom utama. |
| Presensi  | Pagination               | Mengganti halaman data, memilih ukuran halaman (opsional), dan menampilkan total item (jika tersedia).                                         |
| Presensi  | State UI                 | Menampilkan loading, empty state (tidak ada data), dan error state (gagal memuat) secara konsisten pada KPI/tabel.                             |

## 3. Core Process

Alur pengguna (umum):

1. Kamu membuka halaman Presensi.
2. Sistem menampilkan KPI, tabel ringkasan, dan tabel presensi untuk periode default (mis. hari ini/bulan berjalan).
3. Kamu mengatur filter (rentang tanggal, unit, status, pencarian), lalu menekan Terapkan.
4. Sistem memuat ulang KPI, tabel ringkasan, dan tabel presensi berdasarkan filter.
5. Kamu menavigasi data menggunakan pagination untuk melihat halaman berikutnya/sebelumnya.

```mermaid
graph TD
  A["App Shell / Navigasi"] --> B["Halaman Presensi"]
  B --> C["Presensi (Filter Diterapkan)"]
  C --> B
  B --> D["Presensi (Hal
```

