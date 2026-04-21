"use client";

import { useMemo } from "react";
import { useParams, useRouter } from "next/navigation";
import { Download, ChevronRight, ArrowLeft, CalendarDays } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

type PayrollStatus = "DRAFT" | "APPROVED" | "PENDING" | "PAID";

function formatIDR(n: number) {
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function StatusBadge({ status }: { status: PayrollStatus }) {
  switch (status) {
    case "PAID":
      return (
        <Badge className="rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
          PAID
        </Badge>
      );
    case "APPROVED":
      return (
        <Badge className="rounded-full bg-blue-50 text-blue-700 border border-blue-200">
          APPROVED
        </Badge>
      );
    case "PENDING":
      return (
        <Badge className="rounded-full bg-yellow-50 text-yellow-800 border border-yellow-200">
          PENDING
        </Badge>
      );
    default:
      return (
        <Badge className="rounded-full bg-gray-100 text-gray-800 border border-gray-200">
          DRAFT
        </Badge>
      );
  }
}

export default function PayrollDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();

  // mock detail by id (nanti ganti fetch API by params.id)
  const detail = useMemo(() => {
    // contoh data mirip gambar
    return {
      id: params.id,
      periodLabel: "01 Oktober 2023 - 31 Oktober 2023",

      employee: {
        name: "Budi Santoso",
        title: "Senior Manager",
        nik: "SL-2021-0045",
        dept: "Sales & Marketing",
        rekening: "BCA 8901234567",
        bankName: "BCA",
        avatarUrl: "", // optional
      },

      attendance: {
        workDays: 21,
        lateMinutes: 45,
      },

      earnings: [
        { label: "Gaji Pokok", desc: "Base Salary - 160 Jam Kerja", amount: 12_000_000 },
        { label: "Bonus 10% Revenue", desc: "Performance Q3 Achievement", amount: 3_500_000 },
        { label: "Upah Lembur", desc: "Overtime: 12 Jam x Rp 150.000", amount: 1_800_000 },
        { label: "Tunjangan Transportasi", desc: "Fixed Monthly Allowance", amount: 1_000_000 },
      ],

      deductions: [
        { label: "Pajak Penghasilan (PPh 21)", desc: "Sesuai tarif progresif DJP", amount: 1_250_000 },
        { label: "BPJS Ketenagakerjaan", desc: "JHT, JP (Potongan Karyawan 3%)", amount: 360_000 },
        { label: "BPJS Kesehatan", desc: "Potongan Karyawan 1%", amount: 120_000 },
        { label: "Potongan Keterlambatan", desc: "Total 45 Menit (3 kejadian)", amount: 150_000 },
      ],

      netSalary: 16_420_000,
      transferDate: "28 Okt 2023",
      status: "PAID" as PayrollStatus,
    };
  }, [params.id]);

  const totalEarnings = detail.earnings.reduce((sum, x) => sum + x.amount, 0);
  const totalDeductions = detail.deductions.reduce((sum, x) => sum + x.amount, 0);

  return (
    <div className="p-6 space-y-6">
      {/* Breadcrumb + back */}
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <button
              onClick={() => router.push("/dashboard/accountant")}
              className="hover:text-gray-700"
            >
              Dashboard
            </button>
            <ChevronRight className="h-4 w-4" />
            <button
              onClick={() => router.push("/dashboard/accountant/payroll")}
              className="hover:text-gray-700"
            >
              Payroll
            </button>
            <ChevronRight className="h-4 w-4" />
            <span className="text-gray-700">Detail Gaji</span>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              className="rounded-xl"
              onClick={() => router.back()}
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Kembali
            </Button>

            <div>
              <h1 className="text-2xl font-bold text-gray-900">Detail Daftar Gaji</h1>
              <div className="flex items-center gap-2 text-sm text-gray-600 mt-1">
                <CalendarDays className="h-4 w-4" />
                <span>Periode: {detail.periodLabel}</span>
              </div>
            </div>
          </div>
        </div>

        <Button className="rounded-xl gap-2" variant="outline">
          <Download className="h-4 w-4" />
          Download PDF
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* LEFT: Earnings + Deductions */}
        <div className="lg:col-span-2 space-y-4">
          {/* Earnings */}
          <Card className="rounded-2xl">
            <CardContent className="p-6 space-y-4">
              <div className="flex items-center gap-2 font-semibold text-gray-900">
                <span className="h-7 w-7 rounded-full bg-emerald-50 text-emerald-700 flex items-center justify-center text-sm">
                  +
                </span>
                Penghasilan (Earnings)
              </div>

              <div className="divide-y divide-gray-100">
                {detail.earnings.map((e, idx) => (
                  <div key={idx} className="py-4 flex items-start justify-between gap-4">
                    <div>
                      <div className="text-sm font-semibold text-gray-900">{e.label}</div>
                      <div className="text-xs text-gray-500 mt-1">{e.desc}</div>
                    </div>
                    <div className="text-sm font-semibold text-gray-900 whitespace-nowrap">
                      {formatIDR(e.amount)}
                    </div>
                  </div>
                ))}
              </div>

              <div className="pt-2 flex items-center justify-between border-t border-gray-100">
                <div className="text-sm font-semibold text-gray-900">Total Penghasilan Kotor</div>
                <div className="text-lg font-bold text-gray-900">{formatIDR(totalEarnings)}</div>
              </div>
            </CardContent>
          </Card>

          {/* Deductions */}
          <Card className="rounded-2xl">
            <CardContent className="p-6 space-y-4">
              <div className="flex items-center gap-2 font-semibold text-gray-900">
                <span className="h-7 w-7 rounded-full bg-rose-50 text-rose-700 flex items-center justify-center text-sm">
                  -
                </span>
                Potongan (Deductions)
              </div>

              <div className="divide-y divide-gray-100">
                {detail.deductions.map((d, idx) => (
                  <div key={idx} className="py-4 flex items-start justify-between gap-4">
                    <div>
                      <div className="text-sm font-semibold text-gray-900">{d.label}</div>
                      <div className="text-xs text-gray-500 mt-1">{d.desc}</div>
                    </div>
                    <div className="text-sm font-semibold text-rose-600 whitespace-nowrap">
                      -{formatIDR(d.amount)}
                    </div>
                  </div>
                ))}
              </div>

              <div className="pt-2 flex items-center justify-between border-t border-gray-100">
                <div className="text-sm font-semibold text-gray-900">Total Potongan</div>
                <div className="text-lg font-bold text-rose-600">-{formatIDR(totalDeductions)}</div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* RIGHT: Profile + attendance + net */}
        <div className="space-y-4">
          {/* Profile card */}
          <Card className="rounded-2xl">
            <CardContent className="p-6">
              <div className="flex items-center gap-3">
                <div className="h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center text-sm font-semibold text-gray-700">
                  {detail.employee.name
                    .split(" ")
                    .slice(0, 2)
                    .map((x) => x[0])
                    .join("")}
                </div>

                <div className="min-w-0">
                  <div className="text-sm font-semibold text-gray-900">{detail.employee.name}</div>
                  <div className="text-xs text-blue-600">{detail.employee.title}</div>
                </div>
              </div>

              <div className="mt-4 space-y-2 text-sm">
                <div className="flex items-center justify-between text-gray-600">
                  <span>NIK</span>
                  <span className="text-gray-900 font-medium">{detail.employee.nik}</span>
                </div>
                <div className="flex items-center justify-between text-gray-600">
                  <span>DEPT.</span>
                  <span className="text-gray-900 font-medium">{detail.employee.dept}</span>
                </div>
                <div className="flex items-center justify-between text-gray-600">
                  <span>REKENING</span>
                  <span className="text-gray-900 font-medium">{detail.employee.rekening}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Attendance summary */}
          <Card className="rounded-2xl">
            <CardContent className="p-6 space-y-3">
              <div className="font-semibold text-gray-900">Ringkasan Kehadiran</div>

              <div className="grid grid-cols-2 gap-3">
                <div className="rounded-xl border border-gray-100 p-4">
                  <div className="text-xs text-gray-500">Hari Kerja</div>
                  <div className="text-lg font-bold text-gray-900 mt-1">
                    {detail.attendance.workDays} <span className="text-sm font-semibold">Hari</span>
                  </div>
                </div>

                <div className="rounded-xl border border-rose-100 bg-rose-50 p-4">
                  <div className="text-xs text-rose-700">Terlambat</div>
                  <div className="text-lg font-bold text-rose-700 mt-1">
                    {detail.attendance.lateMinutes} <span className="text-sm font-semibold">Menit</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Net salary */}
          <Card className="rounded-2xl bg-blue-600 text-white border-0">
            <CardContent className="p-6 space-y-2">
              <div className="text-sm font-semibold opacity-95">
                Total Bersih Diterima (Net Salary)
              </div>
              <div className="text-3xl font-bold">{formatIDR(detail.netSalary)}</div>

              <div className="flex items-center justify-between pt-2 text-xs opacity-95">
                <span>Ditransfer pada: {detail.transferDate}</span>
                <StatusBadge status={detail.status} />
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}