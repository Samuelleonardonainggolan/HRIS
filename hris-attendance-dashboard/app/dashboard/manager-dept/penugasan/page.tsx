"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CalendarDays, Eye, Plus } from "lucide-react";

type AssignmentStatus = "draft" | "submitted" | "published" | "cancelled";

type AssignmentRow = {
  id: string;
  date: string; // YYYY-MM-DD
  department: string;
  createdBy: string;
  createdByInitial: string;
  createdByColor: string;
  employeesCount: number;
  shift: string; // "08:00 - 16:00"
  status: AssignmentStatus;
};

const MOCK_ASSIGNMENTS: AssignmentRow[] = [
  {
    id: "a1",
    date: "2026-05-07",
    department: "IT Engineering",
    createdBy: "Ananda Dwi",
    createdByInitial: "AD",
    createdByColor: "bg-blue-100 text-blue-700",
    employeesCount: 3,
    shift: "08:00 - 16:00",
    status: "submitted",
  },
  {
    id: "a2",
    date: "2026-05-08",
    department: "IT Engineering",
    createdBy: "Ananda Dwi",
    createdByInitial: "AD",
    createdByColor: "bg-blue-100 text-blue-700",
    employeesCount: 2,
    shift: "15:00 - 23:00",
    status: "draft",
  },
  {
    id: "a3",
    date: "2026-05-05",
    department: "IT Engineering",
    createdBy: "Ananda Dwi",
    createdByInitial: "AD",
    createdByColor: "bg-blue-100 text-blue-700",
    employeesCount: 5,
    shift: "08:00 - 16:00",
    status: "published",
  },
];

function formatDate(dateStr: string) {
  const d = new Date(dateStr);
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" }).replace(/\./g, "");
}

function statusBadge(status: AssignmentStatus) {
  switch (status) {
    case "draft":
      return <Badge className="bg-zinc-100 text-zinc-800 border-zinc-300">Draft</Badge>;
    case "submitted":
      return <Badge className="bg-blue-50 text-blue-800 border-blue-200">Submitted</Badge>;
    case "published":
      return <Badge className="bg-purple-50 text-purple-800 border-purple-200">Published</Badge>;
    case "cancelled":
      return <Badge className="bg-red-50 text-red-800 border-red-200">Cancelled</Badge>;
    default:
      return <Badge>{status}</Badge>;
  }
}

export default function DaftarPenugasanKadep() {
  const [status, setStatus] = useState<string>("all");
  const [date, setDate] = useState<string>("");
  const [q, setQ] = useState<string>("");

  const filtered = useMemo(() => {
    return MOCK_ASSIGNMENTS.filter((row) => {
      const matchStatus = status === "all" || row.status === status;
      const matchDate = !date || row.date === date;
      const matchQ =
        !q ||
        row.department.toLowerCase().includes(q.toLowerCase()) ||
        row.createdBy.toLowerCase().includes(q.toLowerCase()) ||
        row.shift.toLowerCase().includes(q.toLowerCase());
      return matchStatus && matchDate && matchQ;
    });
  }, [status, date, q]);

  return (
    <div className="p-8 max-w-[1300px] mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Daftar Penugasan</h1>
          <p className="text-gray-500 mt-2">
            Kelola penugasan karyawan untuk kebutuhan tambahan operasional hotel.
          </p>
        </div>
        <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-5 py-2 font-semibold shadow">
          <Plus className="h-4 w-4 mr-2" />
          Buat Penugasan
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center mb-5">
        <div className="flex flex-wrap items-center gap-3">
          <span className="font-medium text-gray-700">Filters:</span>

          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="w-[180px] rounded-lg">
              <SelectValue placeholder="Status: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Status: All</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="submitted">Submitted</SelectItem>
              <SelectItem value="published">Published</SelectItem>
              <SelectItem value="cancelled">Cancelled</SelectItem>
            </SelectContent>
          </Select>

          <div className="relative">
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} className="pl-8 w-[180px] rounded-lg" />
            <CalendarDays className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
          </div>

          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Cari departemen, pembuat, shift..."
            className="w-[260px] rounded-lg"
          />
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
              <th className="px-4 py-3 text-center">Shift</th>
              <th className="px-4 py-3 text-center">Status</th>
              <th className="px-4 py-3 text-center">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y text-gray-700">
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-6 text-center text-gray-400">
                  Belum ada penugasan.
                </td>
              </tr>
            ) : (
              filtered.map((row) => (
                <tr key={row.id}>
                  <td className="px-4 py-4">{formatDate(row.date)}</td>
                  <td className="px-4 py-4">{row.department}</td>
                  <td className="px-4 py-4">
                    <div className="flex items-center gap-2">
                      <span className={`h-9 w-9 flex items-center justify-center rounded-full font-bold text-xs ${row.createdByColor}`}>
                        {row.createdByInitial}
                      </span>
                      <span>{row.createdBy}</span>
                    </div>
                  </td>
                  <td className="px-4 py-4 text-center">{row.employeesCount}</td>
                  <td className="px-4 py-4 text-center font-semibold">{row.shift}</td>
                  <td className="px-4 py-4 text-center">{statusBadge(row.status)}</td>
                  <td className="px-4 py-4 text-center">
                    <Button variant="ghost" size="icon" className="hover:bg-zinc-100">
                      <Eye className="h-5 w-5" />
                    </Button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        <div className="flex items-center justify-between border-t px-4 py-3 text-sm bg-white">
          <div>Showing 1 to {filtered.length} of {MOCK_ASSIGNMENTS.length} requests</div>
          <div className="flex items-center gap-1">
            <Button size="sm" variant="outline" disabled>
              Previous
            </Button>
            <Button size="sm" className="bg-blue-600 text-white">
              1
            </Button>
            <Button size="sm" variant="outline">
              2
            </Button>
            <Button size="sm" variant="outline">
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}