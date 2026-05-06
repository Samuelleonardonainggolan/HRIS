"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CalendarDays, Eye, Edit2 } from "lucide-react";
import Link from "next/link";

type OvertimeStatus = "Approved" | "Submitted" | "Rejected" | "Draft" | "Published";
type OvertimeRow = {
  id: string;
  date: string; // "2023-10-12"
  department: string;
  requestedBy: string;
  requestedByInitial: string;
  requestedByColor: string; // bg color for avatar
  jumlahKaryawan: number;
  jamLembur: string; // "42.5 Jam"
  status: OvertimeStatus;
};

const MOCK_DATA: OvertimeRow[] = [
  {
    id: "1",
    date: "2023-10-12",
    department: "IT Engineering",
    requestedBy: "Ananda Dwi",
    requestedByInitial: "AD",
    requestedByColor: "bg-blue-100 text-blue-700",
    jumlahKaryawan: 12,
    jamLembur: "42.5 Jam",
    status: "Approved",
  },
  {
    id: "2",
    date: "2023-10-14",
    department: "Operations",
    requestedBy: "Rendi Ramadhan",
    requestedByInitial: "RR",
    requestedByColor: "bg-teal-100 text-teal-700",
    jumlahKaryawan: 8,
    jamLembur: "16.0 Jam",
    status: "Submitted",
  },
  {
    id: "3",
    date: "2023-10-15",
    department: "Marketing",
    requestedBy: "Sarah Hutapea",
    requestedByInitial: "SH",
    requestedByColor: "bg-orange-100 text-orange-700",
    jumlahKaryawan: 5,
    jamLembur: "10.5 Jam",
    status: "Rejected",
  },
  {
    id: "4",
    date: "2023-10-16",
    department: "Human Resources",
    requestedBy: "Budi Kurniawan",
    requestedByInitial: "BK",
    requestedByColor: "bg-purple-100 text-purple-700",
    jumlahKaryawan: 2,
    jamLembur: "4.0 Jam",
    status: "Draft",
  },
  {
    id: "5",
    date: "2023-10-08",
    department: "Logistics",
    requestedBy: "M. Fadli",
    requestedByInitial: "MF",
    requestedByColor: "bg-green-100 text-green-700",
    jumlahKaryawan: 22,
    jamLembur: "88.0 Jam",
    status: "Published",
  },
];

function statusBadge(status: OvertimeStatus) {
  switch (status) {
    case "Approved":
      return <Badge className="bg-green-50 text-green-800 border-green-200">Approved</Badge>;
    case "Submitted":
      return <Badge className="bg-blue-50 text-blue-800 border-blue-200">Submitted</Badge>;
    case "Rejected":
      return <Badge className="bg-red-50 text-red-800 border-red-200">Rejected</Badge>;
    case "Draft":
      return <Badge className="bg-zinc-100 text-zinc-800 border-zinc-300">Draft</Badge>;
    case "Published":
      return <Badge className="bg-purple-50 text-purple-800 border-purple-200">Published</Badge>;
    default:
      return <Badge>{status}</Badge>;
  }
}

function formatDate(dateStr: string) {
  // 2023-10-12 → 12 Okt 2023
  const date = new Date(dateStr);
  return date
    .toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    })
    .replace(/\./g, "");
}

export default function DaftarPengajuanLembur() {
  const [department, setDepartment] = useState("all");
  const [status, setStatus] = useState("all");
  const [date, setDate] = useState("");

  // Filter dummy
  const filtered = MOCK_DATA.filter((row) =>
    (department === "all" || row.department === department) &&
    (status === "all" || row.status === status) &&
    (!date || row.date === date)
  );

  return (
    <div className="p-8">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-8 gap-3">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Daftar Pengajuan Lembur</h1>
          <p className="text-gray-500 mt-2">
            Kelola dan tinjau semua permintaan lembur departemen Anda.
          </p>
        </div>
        <Link href="/dashboard/manager-dept/lembur/tambah-lembur">
          <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-5 py-2 font-semibold shadow" size="lg">
            + Buat Pengajuan Lembur
          </Button>
        </Link>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center mb-5">
        <div className="flex items-center gap-3">
          <span className="font-medium text-gray-700">Filters:</span>
          {/* Department */}
          <Select value={department} onValueChange={setDepartment}>
            <SelectTrigger className="w-[160px] rounded-lg">
              <SelectValue placeholder="Department: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Department: All</SelectItem>
              <SelectItem value="IT Engineering">IT Engineering</SelectItem>
              <SelectItem value="Operations">Operations</SelectItem>
              <SelectItem value="Marketing">Marketing</SelectItem>
              <SelectItem value="Human Resources">Human Resources</SelectItem>
              <SelectItem value="Logistics">Logistics</SelectItem>
            </SelectContent>
          </Select>

          {/* Status */}
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="w-[130px] rounded-lg">
              <SelectValue placeholder="Status: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Status: All Status</SelectItem>
              <SelectItem value="Approved">Approved</SelectItem>
              <SelectItem value="Submitted">Submitted</SelectItem>
              <SelectItem value="Rejected">Rejected</SelectItem>
              <SelectItem value="Draft">Draft</SelectItem>
              <SelectItem value="Published">Published</SelectItem>
            </SelectContent>
          </Select>

          {/* Date */}
          <div className="relative flex items-center gap-2">
            <Input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="pl-8 w-[170px] rounded-lg"
              placeholder="Pilih tanggal"
            />
            <CalendarDays className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-auto rounded-xl border bg-white">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-50 text-gray-500 font-semibold text-xs uppercase">
              <th className="px-4 py-3 text-left">Tanggal</th>
              <th className="px-4 py-3 text-left">Departemen</th>
              <th className="px-4 py-3 text-left">Diajukan Oleh</th>
              <th className="px-4 py-3 text-center">Jumlah Karyawan</th>
              <th className="px-4 py-3 text-center">Jam Lembur</th>
              <th className="px-4 py-3 text-center">Status</th>
              <th className="px-4 py-3 text-center">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y text-gray-700">
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-6 text-center text-gray-400">Belum ada pengajuan lembur.</td>
              </tr>
            ) : (
              filtered.map((row) => (
                <tr key={row.id}>
                  <td className="px-4 py-4">{formatDate(row.date)}</td>
                  <td className="px-4 py-4">{row.department}</td>
                  <td className="px-4 py-4">
                    <div className="flex items-center gap-2">
                      <span className={`h-9 w-9 flex items-center justify-center rounded-full font-bold text-xs ${row.requestedByColor}`}>
                        {row.requestedByInitial}
                      </span>
                      <span>{row.requestedBy}</span>
                    </div>
                  </td>
                  <td className="px-4 py-4 text-center">{row.jumlahKaryawan}</td>
                  <td className="px-4 py-4 text-center font-semibold">{row.jamLembur}</td>
                  <td className="px-4 py-4 text-center">
                    {statusBadge(row.status)}
                  </td>
                  <td className="px-4 py-4 text-center">
                    <div className="flex justify-center items-center gap-2">
                      <Button variant="ghost" size="icon" className="hover:bg-zinc-100">
                        <Eye className="h-5 w-5" />
                      </Button>
                      <Button variant="ghost" size="icon" className="hover:bg-zinc-100">
                        <Edit2 className="h-5 w-5" />
                      </Button>
                      {/* More actions, e.g. dropdown, can be added here */}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        {/* Pagination */}
        <div className="flex items-center justify-between border-t px-4 py-3 text-sm bg-white">
          <div>
            Showing 1 to {filtered.length} of {MOCK_DATA.length} requests
          </div>
          <div className="flex items-center gap-1">
            <Button size="icon" variant="outline" disabled>
              Previous
            </Button>
            <Button size="icon" className="bg-blue-600 text-white" variant="default">
              1
            </Button>
            <Button size="icon" variant="outline">
              2
            </Button>
            <Button size="icon" variant="outline">
              3
            </Button>
            <Button size="icon" variant="outline">
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}