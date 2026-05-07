"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CalendarDays, Eye, Edit2, Loader2 } from "lucide-react";
import Link from "next/link";
import { deptOvertimeRequestsApi } from "@/lib/api/overtime-requests";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

function statusBadge(status: string) {
  const s = status?.toLowerCase();
  switch (s) {
    case "approved":
      return <Badge className="bg-green-50 text-green-800 border-green-200">Approved</Badge>;
    case "submitted":
      return <Badge className="bg-blue-50 text-blue-800 border-blue-200">Submitted</Badge>;
    case "rejected":
      return <Badge className="bg-red-50 text-red-800 border-red-200">Rejected</Badge>;
    case "draft":
      return <Badge className="bg-zinc-100 text-zinc-800 border-zinc-300">Draft</Badge>;
    case "published":
      return <Badge className="bg-purple-50 text-purple-800 border-purple-200">Published</Badge>;
    default:
      return <Badge>{status || "Unknown"}</Badge>;
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

function getDuration(start: string, end: string) {
  if (!start || !end) return "0 Jam";
  const [sH, sM] = start.split(":").map(Number);
  const [eH, eM] = end.split(":").map(Number);
  let jam = (eH + eM / 60) - (sH + sM / 60);
  if (jam < 0) jam += 24;
  return `${Math.max(parseFloat(jam.toFixed(1)), 0)} Jam`;
}

export default function DaftarPengajuanLembur() {
  const [department, setDepartment] = useState("all");
  const [status, setStatus] = useState("all");
  const [date, setDate] = useState("");
  
  const [data, setData] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const res = await deptOvertimeRequestsApi.list();
        setData(res);
      } catch (error) {
        console.error("Failed to fetch overtime data", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, []);

  const filtered = data.filter((row) => {
    const rowDept = row.department_name || "";
    const rowStatus = row.status || "";
    const rowDate = row.date ? row.date.substring(0, 10) : "";

    const matchDept = department === "all" || rowDept === department;
    const matchStatus = status === "all" || rowStatus.toLowerCase() === status.toLowerCase();
    const matchDate = !date || rowDate === date;

    return matchDept && matchStatus && matchDate;
  });

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
              {Array.from(new Set(data.map(d => d.department_name).filter(Boolean))).map(dept => (
                <SelectItem key={dept} value={dept}>{dept}</SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Status */}
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="w-[130px] rounded-lg">
              <SelectValue placeholder="Status: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Status: All Status</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="submitted">Submitted</SelectItem>
              <SelectItem value="published">Published</SelectItem>
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
              <th className="px-4 py-3 text-left">Waktu Lembur</th>
              <th className="px-4 py-3 text-left">Karyawan</th>
              <th className="px-4 py-3 text-left">Alasan</th>
              <th className="px-4 py-3 text-center">Status</th>
              <th className="px-4 py-3 text-center">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y text-gray-700">
            {isLoading ? (
              <tr>
                <td colSpan={6} className="py-12 text-center">
                  <div className="flex justify-center items-center">
                    <Loader2 className="h-6 w-6 animate-spin text-blue-600" />
                  </div>
                </td>
              </tr>
            ) : filtered.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-6 text-center text-gray-400">Belum ada pengajuan lembur.</td>
              </tr>
            ) : (
              filtered.map((row) => (
                <tr key={row.id}>
                  <td className="px-4 py-4">{formatDate(row.date)}</td>
                  <td className="px-4 py-4">
                    <div className="font-medium text-gray-900">{row.start_time?.slice(0, 5) || "-"} - {row.end_time?.slice(0, 5) || "-"}</div>
                    <div className="text-xs text-gray-500">{getDuration(row.start_time, row.end_time)}</div>
                  </td>
                  <td className="px-4 py-4">
                    <Dialog>
                      <DialogTrigger asChild>
                        <div className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded-lg transition-colors w-fit">
                          <div className="flex -space-x-2">
                            {row.employees?.slice(0, 3).map((emp: any, i: number) => {
                              const name = emp.user?.full_name || emp.name || emp.full_name || emp.employee_name || "?";
                              return (
                                <span key={i} className="h-8 w-8 rounded-full border-2 border-white bg-blue-100 flex items-center justify-center text-xs font-bold text-blue-700" title={name}>
                                  {name !== "?" ? name.substring(0, 2).toUpperCase() : "?"}
                                </span>
                              );
                            })}
                          </div>
                          <span className="text-sm font-medium text-gray-700">
                            {row.employees?.length || 0} Orang
                          </span>
                        </div>
                      </DialogTrigger>
                      <DialogContent className="sm:max-w-[425px]">
                        <DialogHeader>
                          <DialogTitle>Daftar Karyawan Lembur</DialogTitle>
                          <DialogDescription>
                            Berikut adalah daftar karyawan yang diajukan untuk lembur pada tanggal {formatDate(row.date)}.
                          </DialogDescription>
                        </DialogHeader>
                        <div className="mt-4 space-y-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
                          {row.employees?.map((emp: any, i: number) => {
                            const name = emp.full_name || emp.user?.full_name || emp.name || emp.employee_name || "-";
                            const payroll = emp.payroll_number || emp.user?.payroll_number || emp.nik || "N/A";
                            const position = emp.position_name || emp.user?.position_name || "-";
                            return (
                              <div key={i} className="flex items-center justify-between p-3 rounded-xl border bg-white shadow-sm hover:shadow-md transition-shadow">
                                <div className="flex items-center gap-3">
                                  <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-sm font-bold text-white shadow-sm">
                                    {name.substring(0, 2).toUpperCase()}
                                  </div>
                                  <div>
                                    <p className="text-sm font-bold text-gray-900">{name}</p>
                                    <p className="text-xs text-gray-500 font-medium">
                                      {payroll} • {position}
                                    </p>
                                  </div>
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </DialogContent>
                    </Dialog>
                  </td>
                  <td className="px-4 py-4 max-w-[200px] truncate" title={row.reason}>
                    {row.reason || "-"}
                  </td>
                  <td className="px-4 py-4 text-center">
                    {statusBadge(row.status)}
                  </td>
                  <td className="px-4 py-4 text-center">
                    <div className="flex justify-center items-center gap-2">
                      <Link href={`/dashboard/manager-dept/lembur/detail/${row.id}`}>
                        <Button variant="ghost" size="icon" className="hover:bg-zinc-100">
                          <Eye className="h-5 w-5" />
                        </Button>
                      </Link>
                      <Link href={`/dashboard/manager-dept/lembur/edit-lembur/${row.id}`}>
                        <Button variant="ghost" size="icon" className="hover:bg-zinc-100">
                          <Edit2 className="h-5 w-5" />
                        </Button>
                      </Link>
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
            Showing {filtered.length === 0 ? 0 : 1} to {filtered.length} of {data.length} requests
          </div>
          <div className="flex items-center gap-1">
            <Button size="icon" variant="outline" disabled>
              Previous
            </Button>
            <Button size="icon" className="bg-blue-600 text-white" variant="default">
              1
            </Button>
            <Button size="icon" variant="outline" disabled>
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}