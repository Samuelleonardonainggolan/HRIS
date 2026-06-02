"use client";

import { useEffect, useMemo, useState } from "react";
import { generateLastMonths, Period } from "@/lib/utils/date-periods";
import { useRouter } from "next/navigation";
import {
  Eye,
  Download,
  Search,
  ChevronDown,
  CalendarDays,
  Loader2,
  RefreshCw,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { payrollApi, PayrollRecord, PayrollStatus } from "@/lib/api/payroll";
import { toast } from "react-hot-toast";

type Period = {
  label: string;
  month: number;
  year: number;
};

function formatIDR(n: number) {
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function StatusBadge({ status }: { status: string }) {
  const s = status.toUpperCase();
  switch (s) {
    case "DRAFT":
      return (
        <Badge className="rounded-full bg-gray-100 text-gray-800 border border-gray-200">
          DRAFT
        </Badge>
      );
    case "PENDING":
      return (
        <Badge className="rounded-full bg-yellow-50 text-yellow-800 border border-yellow-200">
          PENDING
        </Badge>
      );
    case "APPROVED":
      return (
        <Badge className="rounded-full bg-blue-50 text-blue-700 border border-blue-200">
          APPROVED
        </Badge>
      );
    case "PAID":
      return (
        <Badge className="rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
          PAID
        </Badge>
      );
    case "NOT_GENERATED":
      return (
        <Badge className="rounded-full bg-rose-50 text-rose-700 border border-rose-200">
          BELUM DIGENERATE
        </Badge>
      );
    default:
      return <Badge className="rounded-full">{s}</Badge>;
  }
}

export default function PayrollPage() {
  const router = useRouter();

  const [q, setQ] = useState("");
  const [dept, setDept] = useState<string>("Semua Departemen");
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [rows, setRows] = useState<PayrollRecord[]>([]);

  const periods = useMemo(() => generateLastMonths(12, true), []);
  const [period, setPeriod] = useState<Period>(periods[1]);

  const departments = [
    "Semua Departemen",
    "Front Office",
    "Housekeeping",
    "Food & Beverage",
    "IT",
    "HR",
  ];

  const fetchPayrolls = async () => {
    setLoading(true);
    try {
      const data = await payrollApi.getPayrolls({
        month: period.month,
        year: period.year,
      });
      setRows(data);
    } catch (error) {
      console.error(error);
      toast.error("Gagal mengambil data payroll");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPayrolls();
  }, [period]);

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      await payrollApi.generatePayrolls(period.month, period.year);
      toast.success("Payroll berhasil digenerate");
      fetchPayrolls();
    } catch (error) {
      console.error(error);
      toast.error("Gagal generate payroll");
    } finally {
      setGenerating(false);
    }
  };

  const filtered = useMemo(() => {
    const qq = q.trim().toLowerCase();

    if (!rows) return [];

    return rows.filter((r) => {
      const matchQ =
        !qq ||
        r.name.toLowerCase().includes(qq) ||
        r.position.toLowerCase().includes(qq) ||
        r.department.toLowerCase().includes(qq);

      const matchDept = dept === "Semua Departemen" || r.department === dept;

      return matchQ && matchDept;
    });
  }, [rows, q, dept]);

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Payroll</h1>
        <p className="text-gray-600">Daftar payroll karyawan per periode.</p>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-5 space-y-4">
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="relative w-full md:max-w-lg">
              <Search className="h-4 w-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari nama karyawan..."
                className="pl-9 rounded-xl"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="rounded-xl gap-2">
                    <CalendarDays className="h-4 w-4 text-gray-600" />
                    {period.label}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {periods.map((p) => (
                    <DropdownMenuItem
                      key={`${p.year}-${p.month}`}
                      onClick={() => setPeriod(p)}
                    >
                      {p.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="rounded-xl gap-2">
                    {dept}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {departments.map((d) => (
                    <DropdownMenuItem key={d} onClick={() => setDept(d)}>
                      {d}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              <Button 
                className="rounded-xl gap-2" 
                variant="outline"
                onClick={handleGenerate}
                disabled={generating || period.value === "all"}
              >
                {generating ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
                Generate Payroll
              </Button>

              <Button className="rounded-xl gap-2" variant="outline">
                <Download className="h-4 w-4" />
                Export
              </Button>
            </div>
          </div>

          <div className="overflow-hidden rounded-xl border border-gray-100">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-100">
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Nama Karyawan
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Gaji Pokok
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Lembur
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Potongan
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Total Bersih
                  </th>
                  <th className="px-6 py-3" />
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-100 bg-white">
                {filtered.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                          {r.initials}
                        </div>
                        <div>
                          <div className="text-sm font-semibold text-gray-900">
                            {r.name}
                          </div>
                          <div className="text-xs text-gray-500">{r.position}</div>
                        </div>
                      </div>
                    </td>

                    <td className="px-6 py-4 text-sm text-gray-700">
                      {formatIDR(r.basicSalary)}
                    </td>

                    <td className="px-6 py-4 text-sm">
                      {r.overtime > 0 ? (
                        <span className="font-semibold text-emerald-600">
                          +{formatIDR(r.overtime)}
                        </span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>

                    <td className="px-6 py-4 text-sm">
                      {r.deduction > 0 ? (
                        <span className="font-semibold text-rose-600">
                          -{formatIDR(r.deduction)}
                        </span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>

                    <td className="px-6 py-4">
                      <StatusBadge status={r.status} />
                    </td>

                    <td className="px-6 py-4 text-sm font-semibold text-gray-900">
                      {formatIDR(r.netTotal)}
                    </td>

                    <td className="px-6 py-4 text-right">
                      {r.id && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="rounded-xl"
                          onClick={() => router.push(`/dashboard/accountant/payroll/${r.id}`)}
                          aria-label="Lihat detail gaji"
                        >
                          <Eye className="h-4 w-4 text-blue-600" />
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}

                {loading ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-10 text-center text-sm text-gray-500">
                      <div className="flex items-center justify-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Memuat data...
                      </div>
                    </td>
                  </tr>
                ) : filtered.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-10 text-center text-sm text-gray-500">
                      Tidak ada data payroll untuk periode <b>{period.label}</b>.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>

          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between text-sm text-gray-500">
            <div>
              Menampilkan 1-{Math.min(filtered.length, 10)} dari {filtered.length} data
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" className="rounded-xl" size="sm">
                &lt;
              </Button>
              <Button variant="secondary" className="rounded-xl" size="sm">
                1
              </Button>
              <Button variant="ghost" className="rounded-xl" size="sm">
                2
              </Button>
              <Button variant="ghost" className="rounded-xl" size="sm">
                3
              </Button>
              <Button variant="outline" className="rounded-xl" size="sm">
                &gt;
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}