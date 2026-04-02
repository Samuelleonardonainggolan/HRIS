"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Download,
  Calendar,
  MapPin,
  Search,
  ChevronLeft,
  ChevronRight,
  MoreVertical,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { format } from "date-fns";
import { id } from "date-fns/locale";
import type { DateRange } from "react-day-picker";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as UiCalendar } from "@/components/ui/calender"; // ✅

type AttendanceStatus = "HADIR" | "TELAT" | "IZIN" | "ALFA";

interface AttendanceRow {
  id: string;
  name: string;
  email: string;

  position: string; // ✅ Jabatan (pengganti ID/Dept)

  date: Date;
  dateLabel: string;

  clockIn: string;
  clockOut: string;
  status: AttendanceStatus;
  location: string;
}

function statusBadgeVariant(status: AttendanceStatus) {
  switch (status) {
    case "HADIR":
      return "success" as any;
    case "TELAT":
      return "warning" as any;
    case "IZIN":
      return "secondary" as any;
    case "ALFA":
      return "destructive" as any;
    default:
      return "secondary" as any;
  }
}

export default function PresensiKaryawanManagerDepartemenPage() {
  const [isLoading, setIsLoading] = useState(true);

  const defaultRange: DateRange = {
    from: new Date(2023, 9, 1),
    to: new Date(2023, 9, 31),
  };
  const [dateRange, setDateRange] = useState<DateRange | undefined>(defaultRange);

  const dateRangeLabel = useMemo(() => {
    if (!dateRange?.from) return "Pilih rentang tanggal";
    const from = format(dateRange.from, "dd MMM yyyy", { locale: id });
    const to = dateRange.to ? format(dateRange.to, "dd MMM yyyy", { locale: id }) : "-";
    return `${from} - ${to}`;
  }, [dateRange]);

  // ✅ filter jabatan
  const [positionFilter, setPositionFilter] = useState("all");

  const [searchEmployee, setSearchEmployee] = useState("");

  const [page, setPage] = useState(1);
  const pageSize = 4;

  const [rows, setRows] = useState<AttendanceRow[]>([]);

  useEffect(() => {
    const t = setTimeout(() => {
      setRows([
        {
          id: "1",
          name: "Budi Santoso",
          email: "budi.s@sapphire.com",
          position: "Software Developer",
          date: new Date(2023, 9, 25),
          dateLabel: "25 Okt 2023",
          clockIn: "08:00",
          clockOut: "17:05",
          status: "HADIR",
          location: "HQ Office",
        },
        {
          id: "2",
          name: "Siti Aminah",
          email: "siti.a@sapphire.com",
          position: "Accountant",
          date: new Date(2023, 9, 25),
          dateLabel: "25 Okt 2023",
          clockIn: "08:45",
          clockOut: "17:15",
          status: "TELAT",
          location: "Remote - Home",
        },
        {
          id: "3",
          name: "Rian Wibawa",
          email: "rian.w@sapphire.com",
          position: "Sales Executive",
          date: new Date(2023, 9, 25),
          dateLabel: "25 Okt 2023",
          clockIn: "--:--",
          clockOut: "--:--",
          status: "IZIN",
          location: "Unrecorded",
        },
        {
          id: "4",
          name: "Diana Putri",
          email: "diana.p@sapphire.com",
          position: "Software Developer",
          date: new Date(2023, 9, 10),
          dateLabel: "10 Okt 2023",
          clockIn: "07:55",
          clockOut: "17:30",
          status: "HADIR",
          location: "HQ Office",
        },
      ]);
      setIsLoading(false);
    }, 500);

    return () => clearTimeout(t);
  }, []);

  // ✅ option jabatan dari data (atau nanti dari API)
  const positionOptions = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.position).filter(Boolean))).sort();
    return uniq;
  }, [rows]);

  const filteredRows = useMemo(() => {
    const q = searchEmployee.toLowerCase().trim();

    const from = dateRange?.from
      ? new Date(dateRange.from.getFullYear(), dateRange.from.getMonth(), dateRange.from.getDate(), 0, 0, 0, 0)
      : null;

    const to = dateRange?.to
      ? new Date(dateRange.to.getFullYear(), dateRange.to.getMonth(), dateRange.to.getDate(), 23, 59, 59, 999)
      : null;

    return rows.filter((r) => {
      const matchName = r.name.toLowerCase().includes(q) || r.email.toLowerCase().includes(q);
      const matchPosition = positionFilter === "all" ? true : r.position === positionFilter;
      const matchDate = !from || !to ? true : r.date >= from && r.date <= to;

      return matchName && matchPosition && matchDate;
    });
  }, [rows, searchEmployee, positionFilter, dateRange]);

  const totalItems = filteredRows.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));

  const pagedRows = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredRows.slice(start, start + pageSize);
  }, [filteredRows, page]);

  useEffect(() => {
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  const fromText = totalItems === 0 ? 0 : (page - 1) * pageSize + 1;
  const toText = Math.min(page * pageSize, totalItems);

  const handleReset = () => {
    setDateRange(defaultRange);
    setPositionFilter("all");
    setSearchEmployee("");
    setPage(1);
  };

  const handleApplyFilter = () => {
    setPage(1);
  };

  // summary (dummy)
  const totalKehadiranPct = 94;
  const tepatWaktu = 42;
  const terlambat = 4;
  const izinSakit = 2;

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold text-gray-900">Presensi Karyawan</h1>
          <p className="text-sm text-gray-600">
            Kelola data kehadiran harian karyawan dalam departemen Anda
          </p>
        </div>

        <Button className="bg-blue-600 hover:bg-blue-700 text-white gap-2">
          <Download className="h-4 w-4" />
          Export Laporan
        </Button>
      </div>

      {/* Filter Card */}
      <Card className="rounded-2xl">
        <CardContent className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
            {/* Filter Tanggal */}
            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">
                Filter Tanggal
              </div>

              <Popover>
                <PopoverTrigger asChild>
                  <button
                    type="button"
                    className="w-full flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-sm
                               hover:bg-gray-50 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    <Calendar className="h-4 w-4 text-gray-400" />
                    <span className={dateRange?.from ? "text-gray-900" : "text-gray-400"}>
                      {dateRangeLabel}
                    </span>
                  </button>
                </PopoverTrigger>

                <PopoverContent className="w-auto p-0" align="start">
                  <UiCalendar
                    mode="range"
                    numberOfMonths={2}
                    selected={dateRange}
                    onSelect={setDateRange}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* ✅ Filter Jabatan */}
            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">
                Filter Jabatan
              </div>
              <Select value={positionFilter} onValueChange={setPositionFilter}>
                <SelectTrigger className="rounded-xl">
                  <SelectValue placeholder="Semua Jabatan" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Jabatan</SelectItem>
                  {positionOptions.map((p) => (
                    <SelectItem key={p} value={p}>
                      {p}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Actions */}
            <div className="flex md:justify-end gap-3">
              <Button variant="outline" className="rounded-xl" onClick={handleReset}>
                Reset
              </Button>
              <Button
                className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl"
                onClick={handleApplyFilter}
                disabled={!dateRange?.from || !dateRange?.to}
                title={!dateRange?.to ? "Pilih rentang tanggal lengkap (mulai & selesai)" : undefined}
              >
                Terapkan Filter
              </Button>
            </div>
          </div>

          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              value={searchEmployee}
              onChange={(e) => setSearchEmployee(e.target.value)}
              placeholder="Cari karyawan..."
              className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                         focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="rounded-2xl">
        <CardContent className="p-0">
          <div className="px-6 pt-6 pb-3">
            {isLoading && <div className="text-sm text-gray-600">Memuat presensi...</div>}
          </div>

          <div className="overflow-hidden rounded-2xl">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-y border-gray-100">
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Nama Karyawan
                    </th>
                    {/* ✅ kolom jabatan */}
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Jabatan
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Tanggal
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Clock In
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Clock Out
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Status
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Lokasi
                    </th>
                    <th className="px-6 py-4 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                      &nbsp;
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-gray-100 bg-white">
                  {pagedRows.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center text-sm font-semibold text-gray-700">
                            {r.name
                              .split(/\s+/)
                              .filter(Boolean)
                              .map((p) => p[0])
                              .join("")
                              .slice(0, 2)
                              .toUpperCase()}
                          </div>
                          <div>
                            <div className="font-semibold text-gray-900">{r.name}</div>
                            <div className="text-xs text-gray-500">{r.email}</div>
                          </div>
                        </div>
                      </td>

                      {/* ✅ jabatan */}
                      <td className="px-6 py-4">
                        <div className="text-sm text-gray-700">{r.position}</div>
                      </td>

                      <td className="px-6 py-4 text-sm text-gray-700">{r.dateLabel}</td>

                      <td className="px-6 py-4">
                        <span
                          className={[
                            "text-sm font-semibold",
                            r.status === "TELAT" ? "text-red-600" : "text-gray-900",
                          ].join(" ")}
                        >
                          {r.clockIn}
                        </span>
                      </td>

                      <td className="px-6 py-4">
                        <span className="text-sm font-semibold text-gray-900">{r.clockOut}</span>
                      </td>

                      <td className="px-6 py-4">
                        <Badge variant={statusBadgeVariant(r.status)}>{r.status}</Badge>
                      </td>

                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2 text-sm text-gray-700">
                          <MapPin className="h-4 w-4 text-gray-400" />
                          <span className={r.location === "Unrecorded" ? "italic text-gray-400" : ""}>
                            {r.location}
                          </span>
                        </div>
                      </td>

                      <td className="px-6 py-4 text-right">
                        <button className="p-2 rounded-lg hover:bg-gray-100">
                          <MoreVertical className="h-4 w-4 text-gray-500" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {!isLoading && filteredRows.length === 0 && (
                <div className="py-12 text-center text-sm text-gray-500">
                  Tidak ada data presensi ditemukan.
                </div>
              )}
            </div>

            <div className="px-6 py-4 flex items-center justify-between text-sm text-gray-600">
              <div>
                Menampilkan {fromText} - {toText} dari {totalItems} karyawan
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                             disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
                >
                  <ChevronLeft className="h-4 w-4" />
                </button>

                {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={[
                      "h-9 w-9 rounded-lg border text-sm font-medium",
                      p === page
                        ? "border-blue-600 text-blue-600 bg-blue-50"
                        : "border-gray-200 text-gray-700 bg-white hover:bg-gray-50",
                    ].join(" ")}
                  >
                    {p}
                  </button>
                ))}

                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                             disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
                >
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Summary cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="rounded-2xl bg-blue-600 text-white">
          <CardContent className="p-5">
            <div className="text-xs font-semibold uppercase opacity-90">Total Kehadiran</div>
            <div className="mt-2 text-3xl font-bold">{totalKehadiranPct}%</div>
            <div className="mt-2 text-xs opacity-90">+2.4% dari bulan lalu</div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-5">
            <div className="text-xs font-semibold uppercase text-gray-500">Tepat Waktu</div>
            <div className="mt-2 text-3xl font-bold text-green-600">{tepatWaktu}</div>
            <div className="mt-2 text-xs text-gray-500">Hadir sesuai jadwal</div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-5">
            <div className="text-xs font-semibold uppercase text-gray-500">Terlambat</div>
            <div className="mt-2 text-3xl font-bold text-orange-600">
              {String(terlambat).padStart(2, "0")}
            </div>
            <div className="mt-2 text-xs text-gray-500">Memerlukan review</div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-5">
            <div className="text-xs font-semibold uppercase text-gray-500">Izin/Sakit</div>
            <div className="mt-2 text-3xl font-bold text-blue-600">
              {String(izinSakit).padStart(2, "0")}
            </div>
            <div className="mt-2 text-xs text-gray-500">Telah disetujui HR</div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}