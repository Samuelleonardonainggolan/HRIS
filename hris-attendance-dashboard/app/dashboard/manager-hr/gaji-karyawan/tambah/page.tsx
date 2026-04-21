"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Save } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type EmployeeOption = { id: string; name: string; nik: string };

export default function TambahMasterGajiPage() {
  const router = useRouter();

  const employees: EmployeeOption[] = useMemo(
    () => [
      { id: "u1", name: "Arya Wijaya", nik: "EMP-2023-041" },
      { id: "u2", name: "Budi Nugroho", nik: "EMP-2023-089" },
      { id: "u3", name: "Citra Dewi", nik: "EMP-2022-112" },
    ],
    []
  );

  const [employee, setEmployee] = useState<EmployeeOption | null>(null);
  const [basicSalary, setBasicSalary] = useState<string>("");
  const [effectiveFrom, setEffectiveFrom] = useState<string>(() => {
    const d = new Date();
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    const dd = String(d.getDate()).padStart(2, "0");
    return `${yyyy}-${mm}-${dd}`;
  });
  const [isActive, setIsActive] = useState(true);

  function onSubmit() {
    // TODO: POST /employee-basic-salaries
    console.log({
      user_id: employee?.id,
      basic_salary: Number(basicSalary),
      effective_from: effectiveFrom,
      is_active: isActive,
    });

    router.push("/dashboard/manager-hr/master-gaji");
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="outline" className="rounded-xl" onClick={() => router.back()}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Kembali
          </Button>

          <div>
            <h1 className="text-2xl font-bold text-gray-900">Tambah Gaji Pokok</h1>
            <p className="text-gray-600">Tambahkan basic salary untuk karyawan.</p>
          </div>
        </div>

        <Button className="rounded-xl gap-2 bg-blue-600 hover:bg-blue-700 text-white" onClick={onSubmit}>
          <Save className="h-4 w-4" />
          Simpan
        </Button>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-6 space-y-6">
          <div className="space-y-2">
            <Label>Karyawan</Label>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="w-full justify-between rounded-xl">
                  {employee ? `${employee.name} (${employee.nik})` : "Pilih karyawan"}
                  <span className="text-gray-400">▼</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width]">
                {employees.map((e) => (
                  <DropdownMenuItem key={e.id} onClick={() => setEmployee(e)}>
                    {e.name} ({e.nik})
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="space-y-2">
            <Label>Gaji Pokok (Rp)</Label>
            <Input
              value={basicSalary}
              onChange={(e) => setBasicSalary(e.target.value)}
              placeholder="contoh: 7500000"
              className="rounded-xl"
              inputMode="numeric"
            />
            <p className="text-xs text-gray-500">
              Simpan sebagai angka tanpa titik/koma. (Mis. 7500000)
            </p>
          </div>

          <div className="space-y-2">
            <Label>Effective From</Label>
            <Input
              type="date"
              value={effectiveFrom}
              onChange={(e) => setEffectiveFrom(e.target.value)}
              className="rounded-xl"
            />
          </div>

          <div className="space-y-2">
            <Label>Status Aktif</Label>
            <div className="flex items-center gap-3 rounded-xl border border-gray-200 p-3">
              <Switch checked={isActive} onCheckedChange={setIsActive} />
              <div className="text-sm text-gray-700">{isActive ? "Active" : "Inactive"}</div>
            </div>
            <p className="text-xs text-gray-500">
              Catatan: 1 karyawan hanya boleh punya 1 basic salary yang aktif.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}