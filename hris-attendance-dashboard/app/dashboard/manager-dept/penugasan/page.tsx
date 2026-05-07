"use client";

import { useMemo, useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CalendarDays, Clock, Eye, Plus, Users, Loader2 } from "lucide-react";
import { assignmentsApi } from "@/lib/api/assignments";
import { useAuth } from "@/contexts/AuthContext";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

type AssignmentStatus = "draft" | "submitted" | "published" | "cancelled";

function statusBadge(status: string) {
  const s = status?.toLowerCase();
  switch (s) {
    case "published":
      return <Badge className="bg-green-50 text-green-700 border-green-100">Dipublikasi</Badge>;
    case "submitted":
      return <Badge className="bg-blue-50 text-blue-700 border-blue-100">Diajukan</Badge>;
    case "cancelled":
      return <Badge className="bg-red-50 text-red-700 border-red-100">Dibatalkan</Badge>;
    case "draft":
    default:
      return <Badge variant="outline" className="text-gray-500">Draft</Badge>;
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return "-";
  const date = new Date(dateStr);
  return date.toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });
}

export default function DaftarPenugasanKadep() {
  const { user } = useAuth();
  const [data, setData] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [status, setStatus] = useState<string>("all");
  const [date, setDate] = useState<string>("");
  const [q, setQ] = useState<string>("");

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        // Menggunakan department_id dari user login
        const res = await assignmentsApi.list(user?.department_id);
        setData(res || []);
      } catch (error) {
        console.error("Failed to fetch assignments", error);
      } finally {
        setIsLoading(false);
      }
    };
    if (user?.department_id) {
      fetchData();
    }
  }, [user?.department_id]);

  const filtered = useMemo(() => {
    return data.filter((row) => {
      const matchStatus = status === "all" || row.status === status;
      const matchDate = !date || row.date.substring(0, 10) === date;
      const matchSearch = !q || 
        row.reason.toLowerCase().includes(q.toLowerCase()) ||
        row.employees.some((e: any) => e.full_name.toLowerCase().includes(q.toLowerCase()));
      
      return matchStatus && matchDate && matchSearch;
    });
  }, [data, status, date, q]);

  return (
    <div className="p-8 max-w-[1300px] mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Daftar Penugasan</h1>
          <p className="text-gray-500 mt-2">
            Kelola penugasan karyawan untuk kebutuhan tambahan operasional hotel.
          </p>
        </div>
        <Link href="/dashboard/manager-dept/penugasan/tambah-penugasan">
          <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-5 py-2 font-semibold shadow">
            <Plus className="h-4 w-4 mr-2" />
            Buat Penugasan
          </Button>
        </Link>
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
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-gray-100">
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Tanggal</th>
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Alasan Penugasan</th>
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Karyawan</th>
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider text-center">Shift Target</th>
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider text-center">Status</th>
                <th className="px-4 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider text-right">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-10 text-center text-gray-500">
                    <div className="flex flex-col items-center gap-2">
                      <Loader2 className="h-6 w-6 animate-spin text-blue-600" />
                      <span>Memuat data...</span>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-10 text-center text-gray-500">
                    Tidak ada data penugasan.
                  </td>
                </tr>
              ) : filtered.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50/50 transition-colors group">
                  <td className="px-4 py-4">
                    <div className="flex items-center gap-2 text-gray-900 font-medium">
                      <CalendarDays className="h-4 w-4 text-gray-400" />
                      {formatDate(row.date)}
                    </div>
                  </td>
                  <td className="px-4 py-4">
                    <div className="font-medium text-gray-900">{row.reason}</div>
                    <div className="text-xs text-gray-400">{row.department_name}</div>
                  </td>
                  <td className="px-4 py-4">
                    <Dialog>
                      <DialogTrigger asChild>
                        <div className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded-lg transition-colors w-fit">
                          <div className="flex -space-x-2">
                            {row.employees?.slice(0, 3).map((emp: any, i: number) => (
                              <span key={i} className="h-8 w-8 rounded-full border-2 border-white bg-blue-100 flex items-center justify-center text-xs font-bold text-blue-700" title={emp.full_name}>
                                {emp.full_name.substring(0, 2).toUpperCase()}
                              </span>
                            ))}
                          </div>
                          <span className="text-sm font-medium text-gray-700">
                            {row.employees.length} Orang
                          </span>
                        </div>
                      </DialogTrigger>
                      <DialogContent className="sm:max-w-[425px]">
                        <DialogHeader>
                          <DialogTitle>Daftar Karyawan Ditugaskan</DialogTitle>
                          <DialogDescription>
                            Berikut adalah daftar karyawan yang diajukan untuk penugasan pada tanggal {formatDate(row.date)}.
                          </DialogDescription>
                        </DialogHeader>
                        <div className="mt-4 space-y-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
                          {row.employees.map((emp: any, i: number) => (
                            <div key={i} className="flex items-center justify-between p-3 rounded-xl border bg-white shadow-sm hover:shadow-md transition-shadow">
                              <div className="flex items-center gap-3">
                                <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-sm font-bold text-white shadow-sm">
                                  {emp.full_name.substring(0, 2).toUpperCase()}
                                </div>
                                <div>
                                  <p className="text-sm font-bold text-gray-900">{emp.full_name}</p>
                                  <p className="text-xs text-gray-500 font-medium">
                                    {emp.payroll_number}
                                  </p>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </DialogContent>
                    </Dialog>
                  </td>
                  <td className="px-4 py-4 text-center">
                    <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-blue-50 text-blue-700 border border-blue-100 text-xs font-bold">
                      <Clock className="h-3 w-3" />
                      {row.shift_start} - {row.shift_end}
                    </div>
                  </td>
                  <td className="px-4 py-4 text-center">
                    {statusBadge(row.status)}
                  </td>
                  <td className="px-4 py-4 text-right">
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0 rounded-full">
                      <Eye className="h-4 w-4 text-gray-500" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="flex items-center justify-between border-t px-4 py-3 text-sm bg-white">
          <div>Showing 1 to {filtered.length} of {data.length} requests</div>
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