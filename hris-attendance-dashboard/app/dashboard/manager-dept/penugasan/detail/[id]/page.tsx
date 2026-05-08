"use client";

import { useState, useEffect, use } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { 
  CalendarDays, 
  Clock, 
  ArrowLeft, 
  Loader2, 
  User as UserIcon, 
  CheckCircle2, 
  XCircle, 
  Clock3,
  FileText,
  Users,
  Award,
  CircleAlert
} from "lucide-react";
import { assignmentsApi } from "@/lib/api/assignments";
import { useRouter, useParams } from "next/navigation";
import { toast } from "react-hot-toast";
import Link from "next/link";

function statusBadge(status: string) {
  const s = status?.toLowerCase();
  switch (s) {
    case "agreed":
      return (
        <Badge className="bg-green-50 text-green-700 border-green-200 flex items-center gap-1.5 py-1 px-2.5">
          <CheckCircle2 className="h-3 w-3" /> Setuju
        </Badge>
      );
    case "rejected":
      return (
        <Badge className="bg-red-50 text-red-700 border-red-200 flex items-center gap-1.5 py-1 px-2.5">
          <XCircle className="h-3 w-3" /> Ditolak
        </Badge>
      );
    case "pending":
      return (
        <Badge className="bg-amber-50 text-amber-700 border-amber-200 flex items-center gap-1.5 py-1 px-2.5">
          <Clock3 className="h-3 w-3" /> Menunggu
        </Badge>
      );
    default:
      return <Badge className="bg-zinc-100 text-zinc-600 border-zinc-200 uppercase text-[10px]">{status || "-"}</Badge>;
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

export default function DetailPenugasanKadep() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;
  
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const res = await assignmentsApi.getById(id);
      setData(res);
    } catch (error) {
      console.error(error);
      toast.error("Gagal mengambil detail penugasan");
      router.push("/dashboard/manager-dept/penugasan");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (id) fetchData();
  }, [id]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  const employees = data?.employees || [];

  return (
    <div className="px-8 py-8 max-w-[1200px] mx-auto">
      <Link href="/dashboard/manager-dept/penugasan" className="flex items-center gap-2 text-gray-500 hover:text-blue-600 transition-colors mb-6 w-fit">
        <ArrowLeft className="h-4 w-4" />
        <span className="text-sm font-medium">Kembali ke Daftar</span>
      </Link>

      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <h2 className="text-2xl font-bold text-gray-900">Detail Penugasan</h2>
            <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200 uppercase text-[10px] font-bold">
              {data?.status}
            </Badge>
          </div>
          <p className="text-gray-500 text-sm">ID Penugasan: {id}</p>
        </div>
        
        {data?.status === "draft" && (
           <Link href={`/dashboard/manager-dept/penugasan/edit-penugasan/${id}`}>
              <Button className="bg-blue-600 hover:bg-blue-700 text-white shadow-sm">
                Edit Penugasan
              </Button>
           </Link>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column: Info */}
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-white rounded-xl border p-6 shadow-sm">
            <h3 className="font-bold text-gray-900 mb-4 flex items-center gap-2">
              <FileText className="h-5 w-5 text-blue-600" /> Informasi Penugasan
            </h3>
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Tanggal</label>
                <div className="flex items-center gap-2 text-gray-700 font-medium">
                  <CalendarDays className="h-4 w-4 text-gray-400" />
                  {formatDate(data?.date)}
                </div>
              </div>
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Shift Baru</label>
                <div className="flex items-center gap-2 text-gray-700 font-medium">
                  <Clock className="h-4 w-4 text-gray-400" />
                  {data?.start_time} - {data?.end_time}
                </div>
              </div>
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Alasan / Event</label>
                <div className="text-gray-700 text-sm leading-relaxed bg-gray-50 p-3 rounded-lg border italic">
                  "{data?.reason}"
                </div>
              </div>
              {data?.notes && (
                <div>
                  <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Catatan Tambahan</label>
                  <p className="text-gray-600 text-sm">{data.notes}</p>
                </div>
              )}
            </div>
          </div>

          <div className="bg-blue-50/50 rounded-xl border border-blue-100 p-6 shadow-sm">
             <h3 className="font-bold text-blue-900 mb-4 flex items-center gap-2">
                <Award className="h-5 w-5 text-blue-600" /> Ringkasan Reward
             </h3>
             <div className="space-y-3">
                <div className="flex justify-between items-center text-sm">
                   <span className="text-gray-600">Total Ditugaskan</span>
                   <span className="font-bold text-gray-900">{employees.length} Orang</span>
                </div>
                <div className="flex justify-between items-center text-sm">
                   <span className="text-gray-600">Berhak Off Pengganti</span>
                   <span className="font-bold text-green-600">{employees.filter((e: any) => e.day_off_eligible).length} Orang</span>
                </div>
                <div className="pt-3 border-t border-blue-100">
                   <p className="text-[10px] text-blue-500 italic leading-tight">
                      *Karyawan yang berhak Off Pengganti adalah mereka yang seharusnya OFF/Libur pada tanggal tersebut namun ditugaskan masuk.
                   </p>
                </div>
             </div>
          </div>
        </div>

        {/* Right Column: Employees */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-xl border shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b bg-gray-50/50 flex items-center justify-between">
              <h3 className="font-bold text-gray-900 flex items-center gap-2">
                 <Users className="h-5 w-5 text-gray-400" /> Daftar Karyawan & Status
              </h3>
              <span className="text-xs font-medium text-gray-500 bg-white px-2 py-1 rounded border">
                {employees.length} Karyawan
              </span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-gray-400 font-semibold text-[10px] uppercase tracking-wider border-b bg-gray-50/30">
                    <th className="px-6 py-3 text-left">Karyawan</th>
                    <th className="px-6 py-3 text-left">Jadwal Asli</th>
                    <th className="px-6 py-3 text-center">Status Konfirmasi</th>
                    <th className="px-6 py-3 text-right">Catatan</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {employees.map((emp: any) => (
                    <tr key={emp.user_id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 font-bold">
                            {emp.full_name?.substring(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <p className="font-bold text-gray-900">{emp.full_name}</p>
                            <p className="text-[10px] text-gray-500 uppercase font-medium">{emp.payroll_number} • {emp.position_name}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                         <div className="flex flex-col">
                            <span className={`text-[11px] font-bold uppercase ${emp.original_shift_type === 'off' ? 'text-red-500' : 'text-gray-500'}`}>
                               {emp.original_shift_type === 'off' ? 'OFF / Libur' : 'Shift Kerja'}
                            </span>
                            {emp.original_shift_type === 'shift' && (
                               <span className="text-xs text-gray-700">{emp.original_start_time} - {emp.original_end_time}</span>
                            )}
                         </div>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <div className="flex flex-col items-center gap-1">
                          {statusBadge(emp.employee_status)}
                          {emp.confirmed_at && (
                            <span className="text-[10px] text-gray-400">
                              {new Date(emp.confirmed_at).toLocaleString("id-ID", { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-right">
                        {emp.employee_status === "rejected" ? (
                          <div className="text-xs text-red-600 bg-red-50 p-2 rounded-lg border border-red-100 max-w-[200px] ml-auto flex items-start gap-1.5 text-left">
                            <CircleAlert className="h-3.5 w-3.5 mt-0.5 shrink-0" />
                            <div>
                               <strong>Ditolak:</strong> {emp.rejection_note || "Tidak ada alasan spesifik"}
                            </div>
                          </div>
                        ) : emp.day_off_eligible && emp.employee_status === "agreed" ? (
                           <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded bg-green-50 border border-green-100 text-green-700 text-[10px] font-bold uppercase">
                              <Award className="h-3 w-3" /> Berhak Reward Off
                           </div>
                        ) : (
                          <span className="text-xs text-gray-400 italic">
                             {emp.employee_status === 'pending' ? 'Menunggu Konfirmasi' : '-'}
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {employees.length === 0 && (
               <div className="p-8 text-center text-gray-400 text-sm">
                  Tidak ada data karyawan.
               </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
