"use client";

import { useState, useEffect, use } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { CalendarDays, Clock, User as UserIcon, Loader2, ArrowLeft } from "lucide-react";
import { employeeService, Employee } from "@/lib/api/employee";
import { deptOvertimeRequestsApi } from "@/lib/api/overtime-requests";
import { authService } from "@/lib/api/auth";
import { useRouter } from "next/navigation";
import { toast } from "react-hot-toast";
import Link from "next/link";
import { Calendar } from "@/components/ui/calender";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { cn } from "@/lib/utils";

function formatDate(dateStr: string) {
  if (!dateStr) return "-";
  const date = new Date(dateStr);
  return date.toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });
}

export default function EditPengajuanLembur({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = use(params);
  const currentUser = authService.getUser();

  // Form state
  const [date, setDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [searchEmp, setSearchEmp] = useState("");
  const [searchResults, setSearchResults] = useState<Employee[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [pickedEmployees, setPickedEmployees] = useState<Employee[]>([]);
  const [reason, setReason] = useState("");
  const [status, setStatus] = useState("draft");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch initial data
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const data = await deptOvertimeRequestsApi.get(id);
        const overtime = (data as any).overtime || data;
        
        // Cek employees di level atas data atau di dalam objek overtime
        const employeesList = (data as any).employees || (overtime as any).employees || [];
        
        setDate(overtime.date ? overtime.date.substring(0, 10) : "");
        setStartTime(overtime.start_time || "");
        setEndTime(overtime.end_time || "");
        setReason(overtime.reason || "");
        setStatus(overtime.status || "draft");
        
        if (employeesList && Array.isArray(employeesList)) {
          const mappedEmps = employeesList.map((e: any) => ({
            id: e.user?.id || e.user_id || e.id,
            name: e.user?.full_name || e.full_name || e.name || e.employee_name || e.user?.name || "Unknown",
            nik: e.user?.payroll_number || e.payroll_number || e.nik || e.user?.nik || "N/A",
            position: e.user?.position_name || e.position_name || "-"
          }));
          setPickedEmployees(mappedEmps);
        }
      } catch (error) {
        console.error(error);
        toast.error("Gagal mengambil data pengajuan lembur");
        router.push("/dashboard/manager-dept/lembur");
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [id, router]);

  // Estimasi jam lembur
  const estHours = (() => {
    if (!startTime || !endTime) return 0;
    const [sH, sM] = startTime.split(":").map(Number);
    const [eH, eM] = endTime.split(":").map(Number);
    let jam = (eH + eM / 60) - (sH + sM / 60);
    if (jam < 0) jam += 24; 
    return Math.max(parseFloat(jam.toFixed(1)), 0);
  })();

  // Debounced Search
  useEffect(() => {
    const timer = setTimeout(async () => {
      if (searchEmp.trim().length >= 1) {
        setIsSearching(true);
        try {
          const exclude = pickedEmployees.map(e => e.id);
          const results = await employeeService.searchEmployees(searchEmp, exclude, currentUser?.department_id);
          const mapped = (results as any[]).map(u => ({
            id: u.id,
            name: u.full_name,
            nik: u.payroll_number || u.nik || "N/A"
          }));
          setSearchResults(mapped);
        } catch (error) {
          console.error(error);
        } finally {
          setIsSearching(false);
        }
      } else {
        setSearchResults([]);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [searchEmp, pickedEmployees, currentUser?.department_id]);

  const isPicked = (id: string) => pickedEmployees.some((e) => e.id === id);
  
  const addEmployee = (emp: Employee) => {
    if (!isPicked(emp.id)) setPickedEmployees((prev) => [...prev, emp]);
    setSearchEmp("");
    setSearchResults([]);
  };

  const removeEmployee = (id: string) => setPickedEmployees((prev) => prev.filter((e) => e.id !== id));

  const validate = () => {
    if (!date || !startTime || !endTime || !reason || pickedEmployees.length === 0) {
      toast.error("Mohon lengkapi semua data dan pilih minimal satu karyawan.");
      return false;
    }
    return true;
  };

  const handleUpdate = async (newStatus: string) => {
    if (!validate()) return;
    
    setIsSubmitting(true);
    try {
      const payload = {
        date,
        start_time: startTime,
        end_time: endTime,
        reason,
        status: newStatus,
        employees: pickedEmployees.map(e => ({ user_id: e.id }))
      };

      await deptOvertimeRequestsApi.update(id, payload);
      toast.success(newStatus === "draft" ? "Draft berhasil diperbarui" : "Pengajuan berhasil dikirim");
      router.push("/dashboard/manager-dept/lembur");
    } catch (error: any) {
      toast.error(error.message || "Gagal memperbarui pengajuan");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="px-8 py-8 max-w-[1300px] mx-auto">
      <Link href="/dashboard/manager-dept/lembur" className="flex items-center gap-2 text-gray-500 hover:text-blue-600 transition-colors mb-6 w-fit">
        <ArrowLeft className="h-4 w-4" />
        <span className="text-sm font-medium">Kembali ke Daftar</span>
      </Link>

      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Edit Pengajuan Lembur</h2>
        <p className="text-gray-500">
          Perbarui data pengajuan lembur yang telah dibuat sebelumnya.
        </p>
      </div>

      <div className="flex flex-col md:flex-row md:items-start gap-8">
        <div className="w-full md:w-2/3 bg-white rounded-xl shadow-sm border p-10 space-y-6 min-w-[350px]">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <CalendarDays className="h-4 w-4" /> Tanggal Lembur
              </label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant={"outline"}
                    className={cn(
                      "w-full mt-1 justify-start text-left font-normal h-10 rounded-lg border-gray-200",
                      !date && "text-gray-400"
                    )}
                  >
                    <CalendarDays className="mr-2 h-4 w-4 text-gray-400" />
                    {date ? format(new Date(date), "dd MMMM yyyy", { locale: localeId }) : <span>Pilih tanggal</span>}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0 bg-white shadow-xl border rounded-xl" align="start">
                  <Calendar
                    mode="single"
                    selected={date ? new Date(date) : undefined}
                    onSelect={(d) => setDate(d ? format(d, "yyyy-MM-dd") : "")}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <Clock className="h-4 w-4" /> Jam Mulai
              </label>
              <Input
                type="time"
                value={startTime}
                onChange={e => setStartTime(e.target.value)}
                className="mt-1"
                required
              />
            </div>
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <Clock className="h-4 w-4" /> Jam Selesai
              </label>
              <Input
                type="time"
                value={endTime}
                onChange={e => setEndTime(e.target.value)}
                className="mt-1"
                required
              />
            </div>
          </div>

          <div>
            <div className="flex justify-between items-center mb-2">
              <label className="font-semibold text-gray-700 flex items-center gap-1">
                <UserIcon className="h-4 w-4" />
                Pilih Karyawan
              </label>
            </div>
            <div className="relative mb-2">
              <Input
                placeholder="Ketik nama atau NIK karyawan..."
                value={searchEmp}
                onChange={e => setSearchEmp(e.target.value)}
              />
              {isSearching && (
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
                </div>
              )}
              {searchEmp && searchResults.length > 0 && (
                <div className="absolute z-10 bg-white border rounded-lg mt-1 left-0 right-0 max-h-52 overflow-auto shadow-lg">
                  {searchResults.map(emp => (
                    <div
                      key={emp.id}
                      className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm flex items-center"
                      onClick={() => addEmployee(emp)}
                    >
                      <span className="font-medium">{emp.name}</span>
                      <span className="ml-2 text-xs text-gray-500">({emp.nik})</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              {pickedEmployees.map(emp => (
                <div key={emp.id} className="bg-blue-50 text-blue-800 rounded-full px-3 py-1.5 flex items-center gap-2 text-sm font-medium shadow-sm border border-blue-100">
                  <div className="flex flex-col leading-tight">
                    <span>{emp.name}</span>
                    <span className="text-[10px] text-blue-500 font-normal uppercase tracking-wider">{(emp as any).position || "-"} • {emp.nik}</span>
                  </div>
                  <button aria-label="Delete" className="ml-1 focus:outline-none text-blue-400 hover:text-red-500 transition-colors" onClick={() => removeEmployee(emp.id)}>
                    &times;
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div>
            <label className="text-gray-700 font-semibold mb-1 block">Alasan Lembur</label>
            <Textarea
              placeholder="Jelaskan urgensi dan tugas yang akan dikerjakan selama jam lembur…"
              value={reason}
              onChange={e => setReason(e.target.value)}
              rows={4}
              required
            />
          </div>

          <div className="flex items-center justify-end gap-3 mt-8">
            <Button variant="outline" onClick={() => handleUpdate("draft")} disabled={isSubmitting}>
              {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Simpan Perubahan (Draft)
            </Button>
            <Button className="bg-blue-600 hover:bg-blue-700 text-white shadow" onClick={() => handleUpdate("submitted")} disabled={isSubmitting}>
              {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Submit ke HR
            </Button>
          </div>
        </div>

        <div className="w-full md:w-1/3 max-w-xs md:max-w-sm bg-white rounded-xl shadow-sm border p-6 h-fit self-start min-w-[260px]">
          <div className="font-semibold text-blue-800 flex items-center gap-2 text-lg mb-4">
            <CalendarDays className="h-5 w-5" /> Ringkasan Pengajuan
          </div>
          <dl className="space-y-2">
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Tanggal Lembur</dt>
              <dd className="font-medium text-gray-900">{formatDate(date)}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Estimasi Jam</dt>
              <dd className="font-medium text-gray-900">{estHours} Jam</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Total Karyawan</dt>
              <dd className="font-medium text-gray-900">
                {pickedEmployees.length} Orang
              </dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Status Saat Ini</dt>
              <dd>
                <span className="bg-zinc-100 text-zinc-800 rounded-full px-3 py-0.5 text-xs font-medium uppercase">{status}</span>
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}
