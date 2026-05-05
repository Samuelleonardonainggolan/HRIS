"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { CalendarDays, Clock, User } from "lucide-react";

type Employee = {
  id: string;
  name: string;
  nik: string;
};

const MOCK_EMPLOYEES: Employee[] = [
  { id: "1", name: "Aditya Pratama", nik: "IT-001" },
  { id: "2", name: "Siska Amelia", nik: "IT-045" },
  { id: "3", name: "Wawan Permadi", nik: "IT-118" },
  // ... tambahkan lagi karyawan lain
];

export default function BuatPengajuanLemburBaru() {
  // Form state
  const [date, setDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [searchEmp, setSearchEmp] = useState("");
  const [pickedEmployees, setPickedEmployees] = useState<Employee[]>([MOCK_EMPLOYEES[0], MOCK_EMPLOYEES[1]]);
  const [reason, setReason] = useState("");

  // Estimasi jam lembur
  const estHours = (() => {
    if (!startTime || !endTime) return 0;
    const [sH, sM] = startTime.split(":").map(Number);
    const [eH, eM] = endTime.split(":").map(Number);
    let jam = (eH + eM / 60) - (sH + sM / 60);
    if (jam < 0) jam += 24; // handle lewat tengah malam
    return Math.max(jam, 0);
  })();

  // Multi-select karyawan: basic search, tidak autocomplete Highlighter
  const filteredEmp = !searchEmp
    ? MOCK_EMPLOYEES
    : MOCK_EMPLOYEES.filter(
        (e) =>
          e.name.toLowerCase().includes(searchEmp.toLowerCase()) ||
          e.nik.toLowerCase().includes(searchEmp.toLowerCase())
      );
  const isPicked = (id: string) => pickedEmployees.some((e) => e.id === id);
  const addEmployee = (emp: Employee) => {
    if (!isPicked(emp.id)) setPickedEmployees((prev) => [...prev, emp]);
    setSearchEmp("");
  };
  const removeEmployee = (id: string) => setPickedEmployees((prev) => prev.filter((e) => e.id !== id));

  // Submit / Simpan
  const handleSimpanDraft = () => {
    // TODO: Call API as draft
    alert("Saved as draft (dummy).");
  };
  const handleSubmitHR = () => {
    // TODO: Call API for submit
    alert("Submit ke HR (dummy).");
  };

  return (
    <div className="px-8 py-8 max-w-[1300px] mx-auto">
      {/* Breadcrumb */}
      <div className="mb-4 text-sm text-gray-500 flex items-center gap-2">
        <span className="hover:underline cursor-pointer">Dashboard</span>
        <span>/</span>
        <span className="hover:underline cursor-pointer">Pengajuan Lembur</span>
        <span>/</span>
        <span className="text-blue-700 font-bold">Baru</span>
      </div>

      <h2 className="text-2xl font-bold text-gray-900 mb-2">Buat Pengajuan Lembur Baru</h2>
      <p className="mb-7 text-gray-500">
        Lengkapi formulir di bawah ini untuk mengajukan instruksi lembur karyawan.
      </p>

      {/* Flex row, 2/3 form, 1/3 ringkasan, responsif */}
      <div className="flex flex-col md:flex-row md:items-start gap-8">
        {/* ====== Form Kiri ====== */}
        <div className="w-full md:w-2/3 bg-white rounded-xl shadow-sm border p-10 space-y-6 min-w-[350px]">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            <div>
              <label className="text-gray-700 font-semibold mb-1 block flex items-center gap-1">
                <CalendarDays className="h-4 w-4" /> Tanggal Lembur
              </label>
              <Input
                type="date"
                value={date}
                onChange={e => setDate(e.target.value)}
                className="mt-1"
                required
              />
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
                <User className="h-4 w-4" />
                Pilih Karyawan
              </label>
              <a href="#all-karyawan" className="text-blue-600 text-sm hover:underline font-medium">
                Lihat Semua Karyawan
              </a>
            </div>
            <div className="relative mb-2">
              <Input
                placeholder="Ketik nama atau NIK karyawan..."
                value={searchEmp}
                onChange={e => setSearchEmp(e.target.value)}
              />
              {searchEmp && filteredEmp.length > 0 && (
                <div className="absolute z-10 bg-white border rounded-lg mt-1 left-0 right-0 max-h-52 overflow-auto shadow-lg">
                  {filteredEmp.map(emp => (
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
                <div key={emp.id} className="bg-blue-50 text-blue-800 rounded-full px-3 py-1 flex items-center gap-1 text-sm font-medium">
                  {emp.name} <span className="text-xs text-blue-500 ml-2">({emp.nik})</span>
                  <button aria-label="Delete" className="ml-1 focus:outline-none text-blue-600 hover:text-red-500" onClick={() => removeEmployee(emp.id)}>
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
            <Button variant="outline" onClick={handleSimpanDraft}>
              Simpan Draft
            </Button>
            <Button className="bg-blue-600 hover:bg-blue-700 text-white shadow" onClick={handleSubmitHR}>
              Submit ke HR
            </Button>
          </div>
        </div>

        {/* ====== Ringkasan (Kanan, full tinggi form, jaga lebar min) ====== */}
        <div className="w-full md:w-1/3 max-w-xs md:max-w-sm bg-white rounded-xl shadow-sm border p-6 h-fit self-start min-w-[260px]">
          <div className="font-semibold text-blue-800 flex items-center gap-2 text-lg mb-4">
            <CalendarDays className="h-5 w-5" /> Ringkasan Pengajuan
          </div>
          <dl className="space-y-2">
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
              <dt className="text-gray-500 text-sm">Status Pengajuan</dt>
              <dd>
                <span className="bg-zinc-100 text-zinc-800 rounded-full px-3 py-0.5 text-xs font-medium">DRAFT</span>
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}