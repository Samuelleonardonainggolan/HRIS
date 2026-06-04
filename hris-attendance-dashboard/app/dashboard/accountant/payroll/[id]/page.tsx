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
import jsPDF from "jspdf";
import autoTable from "jspdf-autotable";

function formatIDR(n: any) {
  const num = Number(n);
  if (isNaN(num)) return "Rp 0";
  return `Rp ${Math.round(num).toLocaleString("id-ID")}`;
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
        payrollNumber: user.payroll_number || "-",
        dept: user.department_name,
        rekening: "-",
        avatarUrl: user.avatar_url || user.avatar,
      },
      attendance: {
        workDays: parseInt(payroll.total_days_present) || 0,
        lateMinutes: payroll.late_minutes_total || 0,
        absentDays: payroll.absent_days || 0,
      },
      earnings: [
        { label: "Gaji Pokok", desc: `Basis: ${formatIDR((Number(payroll.basic_salary_value) || 0) / (Number(payroll.workdays_divisor) || 24))} / hari`, amount: Number(payroll.basic_salary_value) || 0 },
        { label: "Upah Lembur", desc: `Total ${payroll.overtime_hours_paid || 0} Jam`, amount: Number(payroll.overtime_pay_value) || 0 },
      ],
      deductions: [
        { label: "Potongan Keterlambatan", desc: `Total ${payroll.late_minutes_total || 0} Menit`, amount: Number(payroll.late_deduction_value) || 0 },
        { label: "Potongan Mangkir", desc: `Total ${payroll.absent_days || 0} Hari`, amount: Number(payroll.absent_deduction_value) || 0 },
      ],
      netSalary: Number(payroll.net_salary_value) || 0,
      status: payroll.status,
      updatedAt: payroll.updated_at
    };
  }, [data]);

  const attendanceDetails = useMemo(() => {
    if (!data || !data.payroll) return [];
    const { payroll, attendances, jam_kerja, leaves } = data;
    const year = payroll.year;
    const month = payroll.month;

    // Get number of days in the month
    const daysInMonth = new Date(year, month, 0).getDate();
    const result = [];

    // Current date for comparison to determine if a scheduled day has passed
    const now = new Date();

    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month - 1, day);
      const dayNameStr = format(date, "EEEE", { locale: localeID }); // Senin, Selasa...
      
      const isScheduled = jam_kerja?.day_of_week?.includes(dayNameStr) ?? (dayNameStr !== "Minggu");
      
      const attendance = attendances?.find((a: any) => {
        const d = new Date(a.date);
        return d.getDate() === day && (d.getMonth() + 1) === month && d.getFullYear() === year;
      });

      // Cek apakah hari ini merupakan hari cuti/izin
      const leave = leaves?.find((l: any) => {
        const start = new Date(l.start_date);
        const end = new Date(l.end_date);
        // Normalize time to 00:00:00 for accurate comparison
        start.setHours(0, 0, 0, 0);
        end.setHours(23, 59, 59, 999);
        const current = new Date(date);
        return current >= start && current <= end;
      });

      let status = "Libur";
      if (leave) {
        status = "Cuti/Izin";
      } else if (isScheduled) {
        if (attendance && attendance.clock_in_time) {
          // Mangkir hanya jika tidak ada clock_in sama sekali
          status = attendance.status === "Late" ? "Terlambat" : "Hadir";
        } else if (attendance && !attendance.clock_in_time) {
          // Ada record tapi tidak ada clock_in — anggap hadir (anomali data)
          status = "Hadir";
        } else {
          // Tidak ada record attendance sama sekali
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

  const handleDownloadPDF = () => {
    if (!detail) return;

    const doc = new jsPDF({ orientation: "portrait", unit: "mm", format: "a4" });

    // Logo (sama seperti SPKL)
    try {
      doc.addImage("/logo.jpg", "JPG", 20, 10, 25, 25);
    } catch {
      // skip jika logo tidak ditemukan
    }

    // Header: PT. Labersa Hutahaean (warna gold, sama seperti SPKL)
    doc.setFont("helvetica", "bold");
    doc.setFontSize(18);
    doc.setTextColor(152, 131, 0); // gold
    doc.text("PT. Labersa Hutahaean", 105, 18, { align: "center" });

    // Sub-header (warna hijau, sama seperti SPKL)
    doc.setFontSize(14);
    doc.setTextColor(0, 100, 0); // hijau
    doc.text("HEAD OFFICE - WILAYAH TOBA", 105, 26, { align: "center" });

    // Garis Pemisah
    doc.setDrawColor(0, 0, 0);
    doc.setLineWidth(0.5);
    doc.line(20, 38, 190, 38);

    // Judul Slip Gaji
    doc.setFont("helvetica", "bold");
    doc.setFontSize(14);
    doc.setTextColor(0, 0, 0);
    doc.text(`SLIP GAJI KARYAWAN - ${detail.periodLabel.toUpperCase()}`, 105, 50, { align: "center" });

    // Info Karyawan
    doc.setFontSize(10);
    doc.setFont("helvetica", "normal");
    const startY = 62;
    const lh = 6;
    doc.text(`Nama Karyawan  : ${detail.employee.name}`, 20, startY);
    doc.text(`No. Payroll           : ${detail.employee.payrollNumber}`, 20, startY + lh);
    doc.text(`Jabatan                : ${detail.employee.title || "-"}`, 20, startY + lh * 2);
    doc.text(`Departemen        : ${detail.employee.dept}`, 20, startY + lh * 3);

    doc.text(`Hari Kerja : ${detail.attendance.workDays} Hari`, 130, startY);
    doc.text(`Mangkir    : ${detail.attendance.absentDays} Hari`, 130, startY + lh);
    doc.text(`Telat          : ${detail.attendance.lateMinutes} Menit`, 130, startY + lh * 2);

    // Tabel Penghasilan
    autoTable(doc, {
      startY: startY + lh * 4 + 4,
      margin: { left: 20, right: 20 },
      head: [["Penghasilan", "Keterangan", "Jumlah"]],
      body: detail.earnings.map(e => [e.label, e.desc, formatIDR(e.amount)]),
      theme: "grid",
      headStyles: {
        fillColor: [255, 255, 255],
        textColor: [0, 0, 0],
        fontStyle: "bold",
        halign: "center",
        lineWidth: 0.2,
        lineColor: [0, 0, 0],
      },
      styles: { fontSize: 10, textColor: [0, 0, 0], lineWidth: 0.2, lineColor: [0, 0, 0] },
      foot: [["Total Penghasilan", "", formatIDR(totalEarnings)]],
      footStyles: { fontStyle: "bold", fillColor: [240, 255, 240] },
    });

    // Tabel Potongan
    autoTable(doc, {
      startY: (doc as any).lastAutoTable.finalY + 6,
      margin: { left: 20, right: 20 },
      head: [["Potongan", "Keterangan", "Jumlah"]],
      body: detail.deductions.map(d => [d.label, d.desc, `-${formatIDR(d.amount)}`]),
      theme: "grid",
      headStyles: {
        fillColor: [255, 255, 255],
        textColor: [0, 0, 0],
        fontStyle: "bold",
        halign: "center",
        lineWidth: 0.2,
        lineColor: [0, 0, 0],
      },
      styles: { fontSize: 10, textColor: [0, 0, 0], lineWidth: 0.2, lineColor: [0, 0, 0] },
      foot: [["Total Potongan", "", `-${formatIDR(totalDeductions)}`]],
      footStyles: { fontStyle: "bold", fillColor: [255, 240, 240] },
    });

    // Total Take Home Pay
    const finalY = (doc as any).lastAutoTable.finalY + 12;
    doc.setFontSize(11);
    doc.setFont("helvetica", "bold");
    doc.text("Total Bersih Diterima (Take Home Pay):", 20, finalY);
    doc.text(formatIDR(detail.netSalary), 190, finalY, { align: "right" });

    // Tanda Tangan (sama format SPKL: 3 kolom)
    const col1X = 48;
    const col2X = 105;
    const col3X = 162;
    const today = format(new Date(), "dd MMMM yyyy", { locale: localeID });
    doc.setFontSize(10);
    doc.setFont("helvetica", "normal");
    doc.text(`Balige, ${today}`, col2X, finalY + 18, { align: "center" });

    doc.text("Departement Head,", col1X, finalY + 25, { align: "center" });
    doc.text("Karyawan,", col2X, finalY + 25, { align: "center" });
    doc.text("Disetujui Oleh,", col3X, finalY + 25, { align: "center" });

    const signLineY = finalY + 50;
    doc.text("( ____________________ )", col1X, signLineY, { align: "center" });
    doc.text("( ____________________ )", col2X, signLineY, { align: "center" });
    doc.text("( ____________________ )", col3X, signLineY, { align: "center" });

    doc.setFontSize(9);
    doc.setFont("helvetica", "bold");
    doc.text("Departement Head", col1X, signLineY + 5, { align: "center" });
    doc.text("Karyawan", col2X, signLineY + 5, { align: "center" });
    doc.text("Office Manager / HRM /", col3X, signLineY + 5, { align: "center" });
    doc.text("General Manager", col3X, signLineY + 10, { align: "center" });

    doc.save(`Slip_Gaji_${detail.employee.name.replace(/\s+/g, '_')}_${detail.periodLabel.replace(/\s+/g, '_')}.pdf`);
  };

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

        <Button className="rounded-xl gap-2" variant="outline" onClick={handleDownloadPDF}>
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
                          {day.status === "Cuti/Izin" && (
                            <Badge variant="outline" className="text-blue-700 bg-blue-50 border-blue-100 gap-1">
                              <CheckCircle2 className="h-3 w-3" /> Cuti/Izin
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
                  <span>No. Payroll</span>
                  <span className="text-gray-900 font-medium">{detail.employee.payrollNumber}</span>
                </div>
                <div className="flex items-center justify-between text-gray-600">
                  <span>Jabatan</span>
                  <span className="text-gray-900 font-medium text-right max-w-[55%]">{detail.employee.title || "-"}</span>
                </div>
                <div className="flex items-center justify-between text-gray-600">
                  <span>Dept.</span>
                  <span className="text-gray-900 font-medium">{detail.employee.dept}</span>
                </div>
                <div className="flex items-center justify-between text-gray-600">
                  <span>Status</span>
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