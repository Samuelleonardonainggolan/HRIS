"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectValue, SelectTrigger } from "@/components/ui/select";
import { CalendarDays, Eye, Loader2, Search } from "lucide-react";
import Link from "next/link";
import { overtimeRequestsApi } from "@/lib/api/overtime-requests";
import { toast } from "react-hot-toast";
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
    case "published":
      return <Badge className="bg-purple-50 text-purple-700 border-purple-200">Published</Badge>;
    case "submitted":
      return <Badge className="bg-blue-50 text-blue-700 border-blue-200">Submitted</Badge>;
    case "approved":
      return <Badge className="bg-green-50 text-green-700 border-green-200">Approved</Badge>;
    case "rejected":
      return <Badge className="bg-red-50 text-red-700 border-red-200">Rejected</Badge>;
    case "draft":
      return <Badge className="bg-zinc-100 text-zinc-600 border-zinc-200">Draft</Badge>;
    default:
      return <Badge className="bg-zinc-50 text-zinc-500 border-zinc-200">{status || "Unknown"}</Badge>;
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

export default function DaftarPengajuanLemburHR() {
  const [data, setData] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const res = await overtimeRequestsApi.listForManagerHR({
        status: statusFilter === "all" ? "ALL" : (statusFilter.toUpperCase() as any),
        search: search || undefined
      });
      // Extract overtime data and employee info
      const mapped = (res as any[]).map(item => {
        const ov = item.overtime || item;
        const emp = item.employee || {};
        return {
          id: ov.id,
          date: ov.date,
          department: emp.department_name || ov.department_name || "-",
          requestedBy: emp.full_name || ov.requested_by_name || "System",
          jumlahKaryawan: ov.employees?.length || 0,
          employees: ov.employees || [],
          totalHours: ov.total || "0",
          status: ov.status || ov.final_status || "Unknown",
        };
      });
      setData(mapped);
    } catch (error) {
      console.error(error);
      toast.error("Gagal memuat data lembur");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchData();
    }, 300);
    return () => clearTimeout(timer);
  }, [search, statusFilter]);

  return (
    <div className="p-8 max-w-[1400px] mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-8 gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Monitoring Lembur Perusahaan</h1>
          <p className="text-gray-500 mt-1">
            Pantau semua pengajuan lembur dari seluruh departemen.
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white p-4 rounded-xl border shadow-sm mb-6 flex flex-wrap items-center gap-4">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input 
            placeholder="Cari nama pemohon atau departemen..." 
            className="pl-9 h-10 rounded-lg"
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[180px] h-10 rounded-lg">
            <SelectValue placeholder="Status: Semua" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Status</SelectItem>
            <SelectItem value="submitted">Submitted</SelectItem>
            <SelectItem value="approved">Approved</SelectItem>
            <SelectItem value="published">Published</SelectItem>
            <SelectItem value="rejected">Rejected</SelectItem>
          </SelectContent>
        </Select>

        <Button variant="outline" className="h-10 rounded-lg" onClick={() => fetchData()}>
          Refresh
        </Button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl border shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 text-gray-500 font-semibold text-[10px] uppercase tracking-wider border-b">
                <th className="px-6 py-4 text-left">Tanggal</th>
                <th className="px-6 py-4 text-left">Departemen</th>
                <th className="px-6 py-4 text-left">Diajukan Oleh</th>
                <th className="px-6 py-4 text-center">Jml. Karyawan</th>
                <th className="px-6 py-4 text-center">Total Jam</th>
                <th className="px-6 py-4 text-center">Status</th>
                <th className="px-6 py-4 text-center">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y text-gray-700">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="py-20 text-center">
                    <Loader2 className="h-8 w-8 animate-spin text-blue-600 mx-auto mb-2" />
                    <p className="text-gray-400">Memuat data...</p>
                  </td>
                </tr>
              ) : data.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-20 text-center text-gray-400">
                    Tidak ada pengajuan lembur yang ditemukan.
                  </td>
                </tr>
              ) : (
                data.map((row) => (
                  <tr key={row.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 font-medium">{formatDate(row.date)}</td>
                    <td className="px-6 py-4">
                      <span className="bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs font-semibold">
                        {row.department}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 font-bold text-xs">
                          {row.requestedBy.substring(0, 2).toUpperCase()}
                        </div>
                        <span className="font-medium">{row.requestedBy}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <Dialog>
                        <DialogTrigger asChild>
                          <div className="cursor-pointer hover:bg-gray-100 py-1 rounded-lg transition-all active:scale-95 group">
                            <span className="font-bold text-blue-600 group-hover:underline">
                              {row.jumlahKaryawan}
                            </span>{" "}
                            <span className="text-gray-500">Orang</span>
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
                                      {name !== "?" ? name.substring(0, 2).toUpperCase() : "?"}
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
                    <td className="px-6 py-4 text-center font-bold text-blue-600">{row.totalHours} Jam</td>
                    <td className="px-6 py-4 text-center">
                      {statusBadge(row.status)}
                    </td>
                    <td className="px-6 py-4 text-center">
                      <div className="flex justify-center items-center gap-2">
                        <Link href={`/dashboard/manager-hr/lembur/detail/${row.id}`}>
                          <Button variant="ghost" size="icon" className="hover:bg-blue-50 text-blue-600">
                            <Eye className="h-5 w-5" />
                          </Button>
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

