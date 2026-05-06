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
  Send,
  FileText,
  Building2
} from "lucide-react";
import { overtimeRequestsApi } from "@/lib/api/overtime-requests";
import { useRouter } from "next/navigation";
import { toast } from "react-hot-toast";
import Link from "next/link";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

function statusBadge(status: string) {
  const s = status?.toLowerCase();
  switch (s) {
    case "published":
      return <Badge className="bg-purple-50 text-purple-700 border-purple-200 uppercase text-[10px]">Published</Badge>;
    case "submitted":
      return <Badge className="bg-blue-50 text-blue-700 border-blue-200 uppercase text-[10px]">Submitted</Badge>;
    case "approved":
      return <Badge className="bg-green-50 text-green-700 border-green-200 uppercase text-[10px]">Approved</Badge>;
    case "rejected":
      return <Badge className="bg-red-50 text-red-700 border-red-200 uppercase text-[10px]">Rejected</Badge>;
    default:
      return <Badge className="bg-zinc-100 text-zinc-600 border-zinc-200 uppercase text-[10px]">{status || "-"}</Badge>;
  }
}

function empStatusBadge(status: string) {
  const s = status?.toLowerCase();
  switch (s) {
    case "agreed":
      return <Badge className="bg-green-50 text-green-700 border-green-200">Setuju</Badge>;
    case "rejected":
      return <Badge className="bg-red-50 text-red-700 border-red-200">Ditolak</Badge>;
    case "pending":
      return <Badge className="bg-amber-50 text-amber-700 border-amber-200">Menunggu</Badge>;
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

export default function DetailLemburHR({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = use(params);
  
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isPublishing, setIsPublishing] = useState(false);
  const [showPublishModal, setShowPublishModal] = useState(false);
  const [publishPayload, setPublishPayload] = useState({ letter_url: "", notes: "" });

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const res = await overtimeRequestsApi.getForManagerHR(id);
      setData((res as any).overtime || res);
    } catch (error) {
      console.error(error);
      toast.error("Gagal mengambil detail pengajuan");
      router.push("/dashboard/manager-hr/lembur");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id, router]);

  const handlePublish = async () => {
    if (!publishPayload.letter_url) {
      toast.error("Mohon isi URL SPKL atau lampiran");
      return;
    }
    
    setIsPublishing(true);
    try {
      // In manager-hr, we use overtimeRequestsApi
      await (overtimeRequestsApi as any).publish(id, publishPayload);
      toast.success("SPKL berhasil dipublikasikan");
      setShowPublishModal(false);
      fetchData();
    } catch (error: any) {
      toast.error(error.message || "Gagal mempublikasikan SPKL");
    } finally {
      setIsPublishing(false);
    }
  };

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
      <Link href="/dashboard/manager-hr/lembur" className="flex items-center gap-2 text-gray-500 hover:text-blue-600 transition-colors mb-6 w-fit">
        <ArrowLeft className="h-4 w-4" />
        <span className="text-sm font-medium">Kembali ke Daftar</span>
      </Link>

      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <h2 className="text-2xl font-bold text-gray-900">Detail Monitoring Lembur</h2>
            {statusBadge(data?.status)}
          </div>
          <p className="text-gray-500 text-sm flex items-center gap-2">
            <Building2 className="h-4 w-4" /> {data?.department_name} • ID: {id}
          </p>
        </div>
        
        {data?.status === "submitted" && (
          <Button 
            className="bg-purple-600 hover:bg-purple-700 text-white shadow-sm flex items-center gap-2"
            onClick={() => setShowPublishModal(true)}
          >
            <Send className="h-4 w-4" />
            Publish SPKL (Final)
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-white rounded-xl border p-6 shadow-sm">
            <h3 className="font-bold text-gray-900 mb-4 flex items-center gap-2">
              <FileText className="h-5 w-5 text-purple-600" /> Informasi Pengajuan
            </h3>
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Tanggal</label>
                <div className="flex items-center gap-2 text-gray-700">
                  <CalendarDays className="h-4 w-4 text-gray-400" />
                  {formatDate(data?.date)}
                </div>
              </div>
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Waktu</label>
                <div className="flex items-center gap-2 text-gray-700">
                  <Clock className="h-4 w-4 text-gray-400" />
                  {data?.start_time?.slice(0, 5)} - {data?.end_time?.slice(0, 5)}
                </div>
              </div>
              <div>
                <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Alasan</label>
                <p className="text-gray-700 text-sm leading-relaxed bg-gray-50 p-3 rounded-lg border italic">
                  "{data?.reason}"
                </p>
              </div>
              {data?.letter_url && (
                <div>
                  <label className="text-xs text-gray-400 font-semibold uppercase block mb-1">Dokumen SPKL</label>
                  <a 
                    href={data.letter_url} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline text-sm flex items-center gap-1 font-medium"
                  >
                    Lihat Lampiran SPKL
                  </a>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="lg:col-span-2">
          <div className="bg-white rounded-xl border shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b bg-gray-50/50 flex items-center justify-between">
              <h3 className="font-bold text-gray-900">Daftar Karyawan Terlibat</h3>
              <span className="text-xs font-medium text-gray-500 bg-white px-2 py-1 rounded border">
                {employees.length} Karyawan
              </span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-gray-400 font-semibold text-[10px] uppercase tracking-wider border-b">
                    <th className="px-6 py-3 text-left">Karyawan</th>
                    <th className="px-6 py-3 text-center">Status Mandiri</th>
                    <th className="px-6 py-3 text-right">Catatan</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {employees.map((emp: any) => (
                    <tr key={emp.user_id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center text-purple-700 font-bold">
                            {emp.full_name?.substring(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <p className="font-bold text-gray-900">{emp.full_name}</p>
                            <p className="text-[10px] text-gray-500 uppercase">{emp.payroll_number} • {emp.position_name}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-center">
                        {empStatusBadge(emp.employee_status)}
                      </td>
                      <td className="px-6 py-4 text-right">
                        {emp.employee_status === "rejected" ? (
                          <span className="text-xs text-red-500 italic">{emp.rejection_note}</span>
                        ) : (
                          <span className="text-xs text-gray-400">Tersetujui oleh Sistem</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <Dialog open={showPublishModal} onOpenChange={setShowPublishModal}>
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle>Publikasi SPKL (Role HR)</DialogTitle>
            <DialogDescription>
              Sebagai HR, Anda akan memfinalisasi lembur ini dengan melampirkan surat perintah resmi (SPKL).
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-700">Link Dokumen SPKL</label>
              <Input 
                placeholder="https://cloud.hris.com/spkl-123.pdf"
                value={publishPayload.letter_url}
                onChange={e => setPublishPayload(prev => ({ ...prev, letter_url: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-700">Catatan HR (Opsional)</label>
              <Textarea 
                placeholder="Informasi tambahan untuk departemen terkait..."
                value={publishPayload.notes}
                onChange={e => setPublishPayload(prev => ({ ...prev, notes: e.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowPublishModal(false)}>Batal</Button>
            <Button 
              className="bg-purple-600 hover:bg-purple-700 text-white" 
              onClick={handlePublish}
              disabled={isPublishing}
            >
              {isPublishing ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Send className="h-4 w-4 mr-2" />}
              Publish Lembur
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
