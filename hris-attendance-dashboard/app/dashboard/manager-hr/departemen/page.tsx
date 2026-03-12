"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Search, Plus, Edit2, Trash2, Users } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { departmentApi } from "@/lib/api/department";
import { employeeService } from "@/lib/api/employee";
import toast from "react-hot-toast";

interface Department {
  id: string;
  name: string;
  icon: string;
  managerName: string;
  managerTitle: string;
  totalStaff: number;
  status: "Aktif" | "Nonaktif";
}

export default function DepartmentsPage() {
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState("");
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    loadDepartments();
  }, []);
  
  const loadDepartments = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await departmentApi.getAll();

      let staffCountByDepartmentId: Record<string, number> = {};
      try {
        const employees = await employeeService.getAllEmployees();
        staffCountByDepartmentId = employees.reduce<Record<string, number>>(
          (acc, emp) => {
            const deptId = emp.department_id;
            if (!deptId) return acc;
            acc[deptId] = (acc[deptId] || 0) + 1;
            return acc;
          },
          {}
        );
      } catch (err) {
        console.error("Failed to fetch employees for staff count:", err);
      }

      const mapped: Department[] = data.map((d) => ({
        id: d.id,
        name: d.name,
        icon: d.icon || "🏢",
        managerName: d.managerName || "-",
        managerTitle: "Kepala",
        totalStaff: staffCountByDepartmentId[d.id] ?? d.totalEmployees ?? 0,
        status: d.isActive ? "Aktif" : "Nonaktif",
      }));
      setDepartments(mapped);
    } catch (err) {
      setError("Gagal memuat departemen");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };


  const filteredDepartments = departments.filter((dept) =>
    dept.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleAddDepartment = () => {
    // Navigate to add page instead of modal
    router.push("/dashboard/manager-hr/departemen/tambah-departemen");
  };

  const handleEdit = (id: string) => {
    router.push(`/dashboard/manager-hr/departemen/tambah-departemen?edit=${id}`);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Apakah Anda yakin ingin menghapus departemen ini?")) return;
    try {
      await departmentApi.delete(id);
      toast.success("Departemen berhasil dihapus");
      await loadDepartments();
    } catch (err) {
      console.error(err);
      toast.error("Gagal menghapus departemen");
    }
  };

  return (
    <div className="p-6">
      <Card>
        <CardContent className="p-6">
          {loading && (
            <div className="mb-4 text-sm text-gray-600">Memuat departemen...</div>
          )}
          {error && (
            <div className="mb-4 text-sm text-red-600">{error}</div>
          )}
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">
              Manajemen Departemen
            </h2>
            <Button
              onClick={handleAddDepartment}
              className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
            >
              <Plus className="h-4 w-4" />
              Tambah Departemen
            </Button>
          </div>

          {/* Search Bar */}
          <div className="mb-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Cari departemen..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full rounded-lg border border-gray-300 bg-white py-2 pl-10 pr-4 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Table */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Icon
                  </th>
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Nama Departemen
                  </th>
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Kepala Departemen
                  </th>
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Jumlah Staf
                  </th>
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Status
                  </th>
                  <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                    Aksi
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredDepartments.map((dept) => (
                  <tr key={dept.id} className="hover:bg-gray-50 transition-colors">
                    <td className="py-4">
                      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100">
                        <span className="text-xl">{dept.icon}</span>
                      </div>
                    </td>
                    <td className="py-4">
                      <span className="font-medium text-gray-900">
                        {dept.name}
                      </span>
                    </td>
                    <td className="py-4">
                      <div>
                        <span className="text-sm text-gray-900">
                          {dept.managerName}
                        </span>
                        <p className="text-xs text-gray-500">
                          {dept.managerTitle}
                        </p>
                      </div>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center gap-1 text-sm text-gray-900">
                        <Users className="h-4 w-4 text-gray-400" />
                        <span>{dept.totalStaff} Staf</span>
                      </div>
                    </td>
                    <td className="py-4">
                      <Badge
                        variant={dept.status === "Aktif" ? "success" : "secondary"}
                      >
                        {dept.status}
                      </Badge>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => handleEdit(dept.id)}
                          className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                          title="Edit"
                        >
                          <Edit2 className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(dept.id)}
                          className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                          title="Hapus"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Empty State */}
            {filteredDepartments.length === 0 && (
              <div className="text-center py-12">
                <div className="inline-flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 mb-4">
                  <Search className="h-8 w-8 text-gray-400" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">
                  Tidak ada departemen ditemukan
                </h3>
                <p className="text-sm text-gray-500 mb-4">
                  Coba ubah kata kunci pencarian Anda
                </p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
