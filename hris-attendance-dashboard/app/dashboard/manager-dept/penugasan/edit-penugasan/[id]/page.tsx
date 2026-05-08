"use client";

import { useMemo, useState, useEffect, use } from "react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { useRouter, useParams } from "next/navigation";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { CalendarDays, Clock, User, Loader2, Save, Send, ArrowLeft } from "lucide-react";
import { assignmentsApi } from "@/lib/api/assignments";
import { employeeService } from "@/lib/api/employee";
import { useAuth } from "@/contexts/AuthContext";
import { toast } from "react-hot-toast";

type Employee = {
  id: string;
  name: string;
  nik: string;
  originalShift: {
    type: "shift" | "off";
    time?: string;
  };
};

function calcHours(start: string, end: string) {
  if (!start || !end) return 0;
  const [sH, sM] = start.split(":").map(Number);
  const [eH, eM] = end.split(":").map(Number);
  let jam = (eH + eM / 60) - (sH + sM / 60);
  if (jam < 0) jam += 24;
  return Math.max(0, Number(jam.toFixed(1)));
}

export default function EditPenugasanKadep() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;
  const { user } = useAuth();

  const [date, setDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [reason, setReason] = useState("");
  const [notes, setNotes] = useState("");
  const [status, setStatus] = useState("");

  const [searchEmp, setSearchEmp] = useState("");
  const [availableEmployees, setAvailableEmployees] = useState<any[]>([]);
  const [picked, setPicked] = useState<Employee[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isFetchingData, setIsFetchingData] = useState(true);

  // Fetch initial data
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsFetchingData(true);
        const [assignment, employees] = await Promise.all([
          assignmentsApi.getById(id),
          employeeService.getEmployeesMyDepartment()
        ]);

        setDate(assignment.date.substring(0, 10));
        setStartTime(assignment.start_time);
        setEndTime(assignment.end_time);
        setReason(assignment.reason);
        setNotes(assignment.notes || "");
        setStatus(assignment.status);
        setAvailableEmployees(employees || []);

        const mappedPicked: Employee[] = assignment.employees.map((emp: any) => ({
          id: emp.user_id,
          name: emp.full_name,
          nik: emp.payroll_number,
          originalShift: {
            type: emp.original_shift_type === "off" ? "off" : "shift",
            time: emp.original_shift_type === "shift" ? `${emp.original_start_time} - ${emp.original_end_time}` : undefined
          }
        }));
        setPicked(mappedPicked);
      } catch (err) {
        console.error("Failed to fetch assignment data", err);
        toast.error("Gagal memuat data penugasan");
      } finally {
        setIsFetchingData(false);
      }
    };
    if (id) fetchData();
  }, [id]);

  const filteredEmp = useMemo(() => {
    if (!searchEmp) return [];
    const q = searchEmp.toLowerCase();
    return availableEmployees.filter((e) => 
      e.full_name.toLowerCase().includes(q) || e.payroll_number.toLowerCase().includes(q)
    );
  }, [searchEmp, availableEmployees]);

  const isPicked = (id: string) => picked.some((e) => e.id === id);
  
  const addEmployee = async (empData: any) => {
    if (isPicked(empData.id)) return;
    
    try {
      const preview = await assignmentsApi.previewSchedule(empData.id, date);
      const newEmp: Employee = {
        id: empData.id,
        name: empData.full_name,
        nik: empData.payroll_number,
        originalShift: {
          type: preview.type === "off" ? "off" : "shift",
          time: preview.type === "shift" ? `${preview.start_time} - ${preview.end_time}` : undefined
        }
      };
      setPicked((prev) => [...prev, newEmp]);
      setSearchEmp("");
    } catch (err) {
      console.error("Failed to preview schedule", err);
      toast.error("Gagal mengambil jadwal asli karyawan");
    }
  };

  const removeEmployee = (id: string) => setPicked((prev) => prev.filter((e) => e.id !== id));
  const estHours = calcHours(startTime, endTime);

  const handleUpdate = async (newStatus?: string) => {
    if (!date || !startTime || !endTime || !reason || picked.length === 0) {
      toast.error("Harap lengkapi semua data penugasan");
      return;
    }

    try {
      setIsLoading(true);
      const payload = {
        date,
        reason,
        status: newStatus || status,
        notes,
        start_time: startTime,
        end_time: endTime,
        employees: picked.map(p => ({ user_id: p.id }))
      };

      await assignmentsApi.update(id, payload);
      toast.success("Penugasan berhasil diperbarui");
      router.push("/dashboard/manager-dept/penugasan");
    } catch (err: any) {
      console.error("Failed to update assignment", err);
      toast.error(err.response?.data?.message || "Gagal memperbarui penugasan");
    } finally {
      setIsLoading(false);
    }
  };

  if (isFetchingData) {
    return (
      <div className="flex h-[400px] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="px-8 py-8 max-w-[1300px] mx-auto">
      <div className="mb-4 text-sm text-gray-500 flex items-center gap-2">
        <Link href="/dashboard" className="hover:underline">Dashboard</Link>
        <span>/</span>
        <Link href="/dashboard/manager-dept/penugasan" className="hover:underline">Penugasan</Link>
        <span>/</span>
        <span className="text-blue-700 font-bold">Edit</span>
      </div>

      <div className="flex items-center gap-4 mb-6">
        <Button variant="outline" size="icon" onClick={() => router.back()} className="rounded-full h-10 w-10">
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Edit Penugasan</h2>
          <p className="text-gray-500 text-sm">Sesuaikan detail penugasan yang telah dibuat.</p>
        </div>
      </div>

      <div className="flex flex-col md:flex-row md:items-start gap-8">
        <div className="w-full md:w-2/3 bg-white rounded-xl shadow-sm border p-10 space-y-6 min-w-0">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <CalendarDays className="h-4 w-4" /> Tanggal Penugasan
              </label>
              <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} className="mt-1" />
            </div>
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <Clock className="h-4 w-4" /> Jam Mulai
              </label>
              <Input type="time" value={startTime} onChange={(e) => setStartTime(e.target.value)} className="mt-1" />
            </div>
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <Clock className="h-4 w-4" /> Jam Selesai
              </label>
              <Input type="time" value={endTime} onChange={(e) => setEndTime(e.target.value)} className="mt-1" />
            </div>
          </div>

          <div>
            <label className="font-semibold text-gray-700 flex items-center gap-1 mb-2">
              <User className="h-4 w-4" /> Pilih Karyawan
            </label>
            <div className="relative mb-2">
              <Input
                placeholder="Ketik nama atau NIK karyawan..."
                value={searchEmp}
                onChange={(e) => setSearchEmp(e.target.value)}
              />
              {searchEmp && filteredEmp.length > 0 && (
                <div className="absolute z-10 bg-white border rounded-lg mt-1 left-0 right-0 max-h-52 overflow-auto shadow-lg">
                  {filteredEmp.map((emp) => (
                    <div
                      key={emp.id}
                      className="px-4 py-2 hover:bg-gray-100 cursor-pointer text-sm flex items-center justify-between"
                      onClick={() => addEmployee(emp)}
                    >
                      <div className="flex items-center">
                        <span className="font-medium text-gray-900">{emp.full_name}</span>
                        <span className="ml-2 text-xs text-gray-500">({emp.payroll_number})</span>
                      </div>
                      {isPicked(emp.id) && (
                        <span className="text-[10px] bg-blue-50 text-blue-600 px-2 py-0.5 rounded-full font-bold">Terpilih</span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-2">
              {picked.map((emp) => (
                <div key={emp.id} className="flex items-center justify-between bg-white border border-blue-100 rounded-xl p-3 shadow-sm">
                  <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-sm">
                      {emp.name.split(" ").map(n => n[0]).join("")}
                    </div>
                    <div>
                      <div className="font-bold text-gray-900">{emp.name} <span className="text-xs font-normal text-gray-500">({emp.nik})</span></div>
                      <div className="text-xs flex items-center gap-1 mt-0.5">
                        <span className="text-gray-400">Jadwal Asli:</span>
                        {emp.originalShift.type === "off" ? (
                          <span className="text-red-500 font-semibold uppercase">OFF / Libur</span>
                        ) : (
                          <span className="text-zinc-600 font-medium">{emp.originalShift.time}</span>
                        )}
                        <span className="text-gray-300">→</span>
                        <span className="text-blue-600 font-bold">{startTime || "--:--"} - {endTime || "--:--"}</span>
                      </div>
                    </div>
                  </div>
                  <Button variant="ghost" size="sm" className="text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg" onClick={() => removeEmployee(emp.id)}>
                    Hapus
                  </Button>
                </div>
              ))}
            </div>
          </div>

          <div>
            <label className="text-gray-700 font-semibold mb-1 block">Alasan Penugasan</label>
            <Textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={4} />
          </div>

          <div className="flex items-center justify-end gap-3 mt-8">
            <Button variant="outline" onClick={() => router.back()}>Batal</Button>
            {status === "draft" && (
              <Button className="bg-blue-600 hover:bg-blue-700 text-white flex items-center gap-2" onClick={() => handleUpdate("submitted")} disabled={isLoading}>
                {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                Submit Penugasan
              </Button>
            )}
            <Button className="bg-emerald-600 hover:bg-emerald-700 text-white flex items-center gap-2" onClick={() => handleUpdate()} disabled={isLoading}>
              {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              Simpan Perubahan
            </Button>
          </div>
        </div>

        <div className="w-full md:w-1/3 max-w-xs md:max-w-sm bg-white rounded-xl shadow-sm border p-6 h-fit self-start min-w-[260px]">
          <div className="font-semibold text-blue-800 flex items-center gap-2 text-lg mb-4">
            <CalendarDays className="h-5 w-5" /> Ringkasan Penugasan
          </div>
          <dl className="space-y-2">
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Estimasi Jam</dt>
              <dd className="font-medium text-gray-900">{estHours} Jam</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Total Karyawan</dt>
              <dd className="font-medium text-gray-900">{picked.length} Orang</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Reward (Off)</dt>
              <dd className="font-medium text-green-600">{picked.filter(e => e.originalShift.type === "off").length} Orang</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 text-sm">Status</dt>
              <dd><span className="bg-zinc-100 text-zinc-800 rounded-full px-3 py-0.5 text-xs font-medium uppercase">{status}</span></dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}
