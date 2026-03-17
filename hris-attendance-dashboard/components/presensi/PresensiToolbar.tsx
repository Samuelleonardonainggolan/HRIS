"use client";

import { Calendar, ChevronDown, Download, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type Props = {
  monthRangeLabel: string;
  selectedDepartment: string;
  departments: string[];
  onChangeMonthRangeLabel: (value: string) => void;
  onChangeDepartment: (value: string) => void;
  onApply: () => void;
  onExportPdf: () => void;
  onExportExcel: () => void;
};

export function PresensiToolbar({
  monthRangeLabel,
  selectedDepartment,
  departments,
  onChangeMonthRangeLabel,
  onChangeDepartment,
  onApply,
  onExportPdf,
  onExportExcel,
}: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[1fr_auto]">
      <div className="rounded-2xl border border-gray-200 bg-white p-4">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="grid flex-1 grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-[1fr_1fr_auto]">
            <div>
              <div className="text-[11px] font-semibold tracking-wide text-gray-500">
                FILTER TANGGAL
              </div>
              <Select value={monthRangeLabel} onValueChange={onChangeMonthRangeLabel}>
                <SelectTrigger className="mt-2 h-10 rounded-full">
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-gray-500" />
                    <SelectValue />
                  </div>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={monthRangeLabel}>{monthRangeLabel}</SelectItem>
                  <SelectItem value="Apr 2024 (01/04 - 30/04)">
                    Apr 2024 (01/04 - 30/04)
                  </SelectItem>
                  <SelectItem value="Jun 2024 (01/06 - 30/06)">
                    Jun 2024 (01/06 - 30/06)
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <div className="text-[11px] font-semibold tracking-wide text-gray-500">
                DEPARTEMEN
              </div>
              <Select value={selectedDepartment} onValueChange={onChangeDepartment}>
                <SelectTrigger className="mt-2 h-10 rounded-full">
                  <div className="flex items-center gap-2 text-sm">
                    <SelectValue />
                    <ChevronDown className="h-4 w-4 text-gray-400" />
                  </div>
                </SelectTrigger>
                <SelectContent>
                  {departments.map((dept) => (
                    <SelectItem key={dept} value={dept}>
                      {dept}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-end">
              <Button
                variant="primary"
                size="sm"
                className="h-10 w-full rounded-full px-4"
                onClick={onApply}
              >
                <Filter className="mr-2 h-4 w-4" />
                Terapkan
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-gray-200 bg-white p-4">
        <div className="text-[11px] font-semibold tracking-wide text-gray-500">
          EXPORT LAPORAN
        </div>
        <div className="mt-2 flex flex-wrap gap-3">
          <Button
            variant="outline"
            size="sm"
            className="h-10 rounded-full px-4"
            onClick={onExportPdf}
          >
            <Download className="mr-2 h-4 w-4 text-red-500" />
            Unduh PDF
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-10 rounded-full px-4"
            onClick={onExportExcel}
          >
            <Download className="mr-2 h-4 w-4 text-green-600" />
            Export Excel
          </Button>
        </div>
      </div>
    </div>
  );
}

