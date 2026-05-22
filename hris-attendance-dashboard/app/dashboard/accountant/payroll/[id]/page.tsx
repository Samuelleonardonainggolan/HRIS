"use client";

import { useMemo, useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Download, ChevronRight, ArrowLeft, CalendarDays, CheckCircle2, XCircle, Clock } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { payrollApi, PayrollStatus } from "@/lib/api/payroll";
import { format } from "date-fns";
import { id as localeID } from "date-fns/locale";

function formatIDR(n: number) {
  return `Rp ${Math.round(n).toLocaleString("id-ID")}`;
}

function StatusBadge({ status }: { status: string }) {
  const s = status?.toUpperCase();
  switch (s) {
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
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (params.id) {
      payrollApi.getPayrollDetail(params.id)
        .then(res => setData(res))
        .finally(() => setLoading(false));
    }
  }, [params.id]);

  const detail = useMemo(() => {
    if (!data) return null;
    const { payroll, user } = data;

    const monthNames = [
      "Januari", "Februari", "Maret", "April", "Mei", "Juni",
      "Juli", "Agustus", "September", "Oktober", "November", "Desember"
    ];

    return {
      id: payroll.id,
      periodLabel: `${monthNames[payroll.month - 1]} ${payroll.year}`,
      employee: {
        name: user.full_name,
        title: user.position_name,
        nik: user.nik || "-",
        dept: user.department_name,
        rekening: "-", // Mock for now
        avatarUrl: user.avatar_url,
      },
      attendance: {
        workDays: parseInt(payroll.total_days_present) || 0,
        lateMinutes: payroll.late_minutes_total || 0,
        absentDays: payroll.absent_days || 0,
      },
      earnings: [
        { label: "Gaji Pokok", desc: `Basis: ${formatIDR(payroll.basic_salary_value / (payroll.workdays_divisor || 24))} / hari`, amount: payroll.basic_salary_value },
        { label: "Bonus 10% Revenue", desc: "Performance Incentive", amount: parseInt(payroll.other_earnings) || 0 },
        { label: "Upah Lembur", desc: `Total ${payroll.overtime_hours_paid} Jam`, amount: payroll.overtime_pay_value },
      ],
      deductions: [
        { label: "Potongan Keterlambatan", desc: `Total ${payroll.late_minutes_total} Menit`, amount: payroll.late_deduction_value },
        { label: "Potongan Mangkir", desc: `Total ${payroll.absent_days} Hari`, amount: payroll.absent_deduction_value },
      ],
      netSalary: payroll.net_salary_value,
      status: payroll.status,
      updatedAt: payroll.updated_at
    };
  }, [data]);

  const attendanceDetails = useMemo(() => {
    if (!data || !data.payroll) return [];
    const { payroll, attendances, jam_kerja } = data;
    const year = payroll.year;
    const month = payroll.month;

    // Get number of days in the month
    const daysInMonth = new Date(year, month, 0).getDate();
    const result = [];

    // Current date for comparison (2026-05-22 according to context)
    const now = new Date(2026, 4, 22); // Month is 0-indexed in JS Date

    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month - 1, day);
      const dayNameStr = format(date, "EEEE", { locale: localeID }); // Senin, Selasa...
      
      const isScheduled = jam_kerja?.day_of_week?.includes(dayNameStr);
      const attendance = attendances?.find((a: any) => {
        const d = new Date(a.date);
        return d.getDate() === day && (d.getMonth() + 1) === month && d.getFullYear() === year;
      });

      let status = "Libur";
      if (isScheduled) {
        if (attendance) {
          status = attendance.status === "late" ? "Terlambat" : "Hadir";
        } else {
          // Jika sudah lewat dan tidak ada attendance
          if (date <= now) {
            status = "Mangkir";
          } else {
            status = "Scheduled";
          }
        }
      }

      result.push({
        date: format(date, "dd MMM yyyy", { locale: localeID }),
        dayName: dayNameStr,
        status,
        clockIn: attendance?.clock_in_time ? format(new Date(attendance.clock_in_time), "HH:mm") : "-",
        clockOut: attendance?.clock_out_time ? format(new Date(attendance.clock_out_time), "HH:mm") : "-",
      });
    }

    return result;
  }, [data]);

  if (loading) return <div className="p-10 text-center">Loading Payroll Detail...</div>;
  if (!detail) return <div className="p-10 text-center">Payroll not found</div>;

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
        {/* LEFT: Earnings + Deductions + Attendance Table */}
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

          {/* Attendance Breakdown Table */}
          <Card className="rounded-2xl">
            <CardContent className="p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 font-semibold text-gray-900">
                  <CalendarDays className="h-5 w-5 text-blue-600" />
                  Rincian Kehadiran Harian
                </div>
                <div className="flex items-center gap-4 text-xs font-medium text-gray-500">
                  <div className="flex items-center gap-1">
                    <CheckCircle2 className="h-3 w-3 text-emerald-500" /> Hadir
                  </div>
                  <div className="flex items-center gap-1">
                    <Clock className="h-3 w-3 text-yellow-500" /> Telat
                  </div>
                  <div className="flex items-center gap-1">
                    <XCircle className="h-3 w-3 text-rose-500" /> Mangkir
                  </div>
                </div>
              </div>

              <div className="relative overflow-x-auto rounded-xl border border-gray-100">
                <table className="w-full text-left text-sm">
                  <thead className="bg-gray-50 text-gray-600 font-medium">
                    <tr>
                      <th className="px-4 py-3">Tanggal</th>
                      <th className="px-4 py-3">Hari</th>
                      <th className="px-4 py-3">Status</th>
                      <th className="px-4 py-3 text-center">In</th>
                      <th className="px-4 py-3 text-center">Out</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {attendanceDetails.map((day, idx) => (
                      <tr key={idx} className={day.status === "Libur" ? "bg-gray-50/50" : ""}>
                        <td className="px-4 py-3 text-gray-900">{day.date}</td>
                        <td className="px-4 py-3 text-gray-500">{day.dayName}</td>
                        <td className="px-4 py-3">
                          {day.status === "Hadir" && (
                            <Badge variant="outline" className="text-emerald-700 bg-emerald-50 border-emerald-100 gap-1">
                              <CheckCircle2 className="h-3 w-3" /> Hadir
                            </Badge>
                          )}
                          {day.status === "Terlambat" && (
                            <Badge variant="outline" className="text-yellow-700 bg-yellow-50 border-yellow-100 gap-1">
                              <Clock className="h-3 w-3" /> Telat
                            </Badge>
                          )}
                          {day.status === "Mangkir" && (
                            <Badge variant="outline" className="text-rose-700 bg-rose-50 border-rose-100 gap-1">
                              <XCircle className="h-3 w-3" /> Mangkir
                            </Badge>
                          )}
                          {day.status === "Libur" && (
                            <span className="text-xs text-gray-400">Off</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-center text-gray-600 font-mono text-xs">{day.clockIn}</td>
                        <td className="px-4 py-3 text-center text-gray-600 font-mono text-xs">{day.clockOut}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
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
                <div className="h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center text-sm font-semibold text-gray-700 overflow-hidden">
                  {detail.employee.avatarUrl ? (
                    <img src={detail.employee.avatarUrl} alt="" className="h-full w-full object-cover" />
                  ) : (
                    detail.employee.name.split(" ").slice(0, 2).map((x: string) => x[0]).join("")
                  )}
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
                  <span>STATUS</span>
                  <StatusBadge status={detail.status} />
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
                  <div className="text-xs text-gray-500">Hadir</div>
                  <div className="text-lg font-bold text-gray-900 mt-1">
                    {detail.attendance.workDays} <span className="text-sm font-semibold">Hari</span>
                  </div>
                </div>

                <div className="rounded-xl border border-rose-100 bg-rose-50 p-4">
                  <div className="text-xs text-rose-700">Mangkir</div>
                  <div className="text-lg font-bold text-rose-700 mt-1">
                    {detail.attendance.absentDays} <span className="text-sm font-semibold">Hari</span>
                  </div>
                </div>
              </div>
              
              <div className="p-3 rounded-xl bg-gray-50 border border-gray-100 flex items-center justify-between">
                <div className="text-xs text-gray-600">Total Terlambat</div>
                <div className="text-sm font-bold text-gray-900">{detail.attendance.lateMinutes} Menit</div>
              </div>
            </CardContent>
          </Card>

          {/* Net salary */}
          <Card className="rounded-2xl bg-blue-600 text-white border-0 shadow-lg shadow-blue-200">
            <CardContent className="p-6 space-y-2">
              <div className="text-sm font-semibold opacity-95">
                Total Bersih Diterima (Net Salary)
              </div>
              <div className="text-3xl font-bold">{formatIDR(detail.netSalary)}</div>

              <div className="pt-2 text-[10px] opacity-70 italic border-t border-white/20">
                Terakhir diperbarui: {detail.updatedAt ? new Date(detail.updatedAt).toLocaleString("id-ID") : "-"}
              </div>
            </CardContent>
          </Card>

          <Button className="w-full rounded-xl py-6 font-semibold" variant="outline">
            Kirim Slip Gaji via WhatsApp
          </Button>
        </div>
      </div>
    </div>
  );
}