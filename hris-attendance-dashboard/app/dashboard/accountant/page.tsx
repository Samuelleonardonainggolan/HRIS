"use client";

import { useEffect, useMemo, useState } from "react";
import { generateLastMonths, Period } from "@/lib/utils/date-periods";
import {
  CalendarDays,
  ChevronDown,
  CircleAlert,
  CircleCheck,
  CircleDollarSign,
  FileText,
  ShieldAlert,
  Users,
  Loader2,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { payrollApi, PayrollSummary } from "@/lib/api/payroll";
import { toast } from "react-hot-toast";

type Period = { monthLabel: string; month: number; year: number; value: string }; // YYYY-MM

type TopLate = {
  initials: string;
  name: string;
  department: string;
  lateMinutes: number;
  deductionAmount: number;
};

function formatIDRShort(n: number) {
  if (n >= 1_000_000_000) return `Rp ${(n / 1_000_000_000).toFixed(1)}B`;
  if (n >= 1_000_000) return `Rp ${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `Rp ${(n / 1_000).toFixed(1)}K`;
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function formatIDRFull(n: number) {
  return `Rp ${n.toLocaleString("id-ID")}`;
}



export default function AccountantDashboardPage() {
  const periods = useMemo(() => generateLastMonths(12, true), []);

  const [period, setPeriod] = useState<Period>(periods[1]); // Default to current month (not All)
  const [loading, setLoading] = useState(true);
  const [summary, setSummary] = useState<PayrollSummary | null>(null);
  const [topLate, setTopLate] = useState<TopLate[]>([]);

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      const data = await payrollApi.getSummary(period.month, period.year);
      setSummary(data);

      // Get top 10 late from all payrolls for that month
      const allPayrolls = (await payrollApi.getPayrolls({ month: period.month, year: period.year })) || [];
      const sortedByLate = allPayrolls
        .filter(p => p.deduction > 0)
        .sort((a, b) => b.deduction - a.deduction)
        .slice(0, 10)
        .map(p => ({
          initials: p.initials,
          name: p.name,
          department: p.department,
          lateMinutes: 0, // We'd need late minutes in the API response to show this accurately
          deductionAmount: p.deduction
        }));
      setTopLate(sortedByLate);

    } catch (error) {
      console.error(error);
      toast.error("Gagal mengambil data dashboard");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, [period]);

  const kpis = useMemo(() => {
    if (!summary) return null;
    return {
      totalPayrollNet: summary.totalPayrollNet,
      totalLateDeduction: summary.totalLateDeduction,
      bonusPool: summary.bonusPool,
      employees: summary.employees,
      payrollNetTrend: 0, // Requires historical data
      lateDeductionTrend: 0,
      bonusTag: "Target Met",
      employeeTag: "Total Active",
    };
  }, [summary]);

  const pipeline = summary?.pipeline || { draft: 0, approved: 0, paid: 0 };

  // ✅ PENTING:
  // Jangan bungkus page dengan <Sidebar/> atau <div className="flex ..."> lagi,
  // karena itu sudah dilakukan oleh app/dashboard/layout.tsx
  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">
            Accounting Dashboard
          </h1>
          <p className="text-sm text-gray-600">
            Overview of HRIS financial metrics and statuses.
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Period dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="rounded-xl gap-2">
                <CalendarDays className="h-4 w-4" />
                {period.monthLabel}
                <ChevronDown className="h-4 w-4 text-gray-500" />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              {periods.map((p) => (
                <DropdownMenuItem key={p.value} onClick={() => setPeriod(p)}>
                  {p.monthLabel}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl">
            Buka Periode
          </Button>
        </div>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="rounded-2xl h-32 animate-pulse bg-gray-50" />
          ))
        ) : kpis ? (
          <>
            <KpiCard
              icon={<CircleDollarSign className="h-5 w-5 text-blue-600" />}
              title="Total Payroll (Net)"
              value={formatIDRFull(kpis.totalPayrollNet)}
              trendLabel={`${kpis.payrollNetTrend >= 0 ? "↗" : "↘"} ${Math.abs(
                kpis.payrollNetTrend
              ).toFixed(1)}%`}
              trendTone={kpis.payrollNetTrend >= 0 ? "success" : "danger"}
            />

            <KpiCard
              icon={<ShieldAlert className="h-5 w-5 text-rose-600" />}
              title="Total Deduction (Late)"
              value={formatIDRShort(kpis.totalLateDeduction)}
              trendLabel={`${kpis.lateDeductionTrend >= 0 ? "↗" : "↘"} ${Math.abs(
                kpis.lateDeductionTrend
              ).toFixed(1)}%`}
              trendTone={kpis.lateDeductionTrend <= 0 ? "success" : "danger"}
            />

            <KpiCard
              icon={<FileText className="h-5 w-5 text-orange-600" />}
              title="Bonus Pool (Overtime)"
              value={formatIDRShort(kpis.bonusPool)}
              rightBadge={
                <Badge className="rounded-full bg-gray-900 text-white hover:bg-gray-900">
                  {kpis.bonusTag}
                </Badge>
              }
            />

            <KpiCard
              icon={<Users className="h-5 w-5 text-gray-900" />}
              title="Number of Employees"
              value={`${kpis.employees}`}
              rightBadge={
                <Badge className="rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                  {kpis.employeeTag}
                </Badge>
              }
            />
          </>
        ) : null}
      </div>

      {/* Mid section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Payroll pipeline */}
        <Card className="rounded-2xl lg:col-span-2">
          <CardContent className="p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="font-semibold text-gray-900">
                Payroll Pipeline Progress
              </div>
              <Button variant="link" className="text-blue-600 px-0">
                View Details
              </Button>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <PipelineStat
                icon={<FileText className="h-4 w-4 text-gray-600" />}
                value={pipeline.draft}
                label="DRAFT"
              />
              <PipelineStat
                icon={<CircleCheck className="h-4 w-4 text-blue-600" />}
                value={pipeline.approved}
                label="APPROVED"
                active
              />
              <PipelineStat
                icon={<CircleCheck className="h-4 w-4 text-emerald-600" />}
                value={pipeline.paid}
                label="PAID"
              />
            </div>

            <div className="rounded-xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-900 flex items-start gap-2">
              <CircleAlert className="h-4 w-4 mt-0.5 text-blue-700" />
              <div>
                Currently processing <b>{pipeline.approved}</b> approved records.{" "}
                <b>{pipeline.paid}</b> payments have successfully cleared the
                banking gateway for this period.
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Alerts */}
        <Card className="rounded-2xl">
          <CardContent className="p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="font-semibold text-gray-900">System Alerts</div>
              <Badge className="rounded-full bg-rose-50 text-rose-700 border border-rose-200">
                2 Action Req.
              </Badge>
            </div>

            <div className="space-y-3">
              {alerts.map((a, idx) => (
                <div
                  key={idx}
                  className="flex items-start gap-3 rounded-xl border border-gray-100 p-3"
                >
                  <div
                    className={[
                      "h-9 w-9 rounded-full flex items-center justify-center",
                      a.tone === "warning"
                        ? "bg-orange-50 text-orange-700"
                        : "bg-rose-50 text-rose-700",
                    ].join(" ")}
                  >
                    <CircleAlert className="h-4 w-4" />
                  </div>

                  <div className="flex-1">
                    <div className="text-sm font-semibold text-gray-900">
                      {a.title}
                    </div>
                    <div className="text-xs text-gray-600">{a.desc}</div>
                  </div>
                </div>
              ))}
            </div>

            <Button variant="link" className="text-blue-600 px-0">
              View All Notifications
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Top Late table */}
      <Card className="rounded-2xl">
        <CardContent className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div className="font-semibold text-gray-900">
              Top 10 Late Deductions
            </div>
            <Button variant="ghost" className="rounded-xl">
              ...
            </Button>
          </div>

          <div className="overflow-hidden rounded-xl border border-gray-100">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-100">
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Employee Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Department
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Late Minutes (Total)
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Deduction Amount
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-100 bg-white">
                {loading ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-10 text-center text-sm text-gray-500">
                      <div className="flex items-center justify-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Memuat data...
                      </div>
                    </td>
                  </tr>
                ) : topLate.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-10 text-center text-sm text-gray-500">
                      Tidak ada data pemotongan untuk periode ini.
                    </td>
                  </tr>
                ) : (
                  topLate.map((r, idx) => (
                    <tr key={idx} className="hover:bg-gray-50 transition-colors">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                            {r.initials}
                          </div>
                          <div className="text-sm font-semibold text-gray-900">
                            {r.name}
                          </div>
                        </div>
                      </td>

                      <td className="px-6 py-4 text-sm text-gray-700">
                        {r.department}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-700">
                        {r.lateMinutes > 0 ? `${r.lateMinutes} mins` : "-"}
                      </td>

                      <td className="px-6 py-4 text-right text-sm font-semibold text-rose-600">
                        {`Rp ${r.deductionAmount.toLocaleString("id-ID")}`}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function KpiCard(props: {
  icon: React.ReactNode;
  title: string;
  value: string;
  trendLabel?: string;
  trendTone?: "success" | "danger";
  rightBadge?: React.ReactNode;
}) {
  return (
    <Card className="rounded-2xl">
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div className="h-10 w-10 rounded-xl bg-gray-50 border border-gray-100 flex items-center justify-center">
            {props.icon}
          </div>

          {props.trendLabel ? (
            <Badge
              className={[
                "rounded-full",
                props.trendTone === "success"
                  ? "bg-emerald-50 text-emerald-700 border border-emerald-200"
                  : "bg-rose-50 text-rose-700 border border-rose-200",
              ].join(" ")}
            >
              {props.trendLabel}
            </Badge>
          ) : (
            props.rightBadge ?? null
          )}
        </div>

        <div className="mt-4 text-xs font-semibold uppercase tracking-wider text-gray-500">
          {props.title}
        </div>
        <div className="mt-1 text-lg font-semibold text-gray-900">
          {props.value}
        </div>
      </CardContent>
    </Card>
  );
}

function PipelineStat(props: {
  icon: React.ReactNode;
  value: number;
  label: string;
  active?: boolean;
}) {
  return (
    <div className="rounded-xl border border-gray-100 p-4 flex items-center gap-3">
      <div
        className={[
          "h-9 w-9 rounded-full flex items-center justify-center",
          props.active ? "bg-blue-50 text-blue-700" : "bg-gray-50 text-gray-700",
        ].join(" ")}
      >
        {props.icon}
      </div>

      <div>
        <div className="text-sm font-semibold text-gray-900">{props.value}</div>
        <div
          className={[
            "text-[11px] font-semibold",
            props.active ? "text-blue-600" : "text-gray-500",
          ].join(" ")}
        >
          {props.label}
        </div>
      </div>
    </div>
  );
}