import { apiClient } from "./client";

export type PayrollStatus = "draft" | "pending" | "approved" | "paid";

export type PayrollRecord = {
  id: string;
  user_id: string;
  name: string;
  initials: string;
  position: string;
  department: string;
  basicSalary: number;
  bonus10: number;
  overtime: number;
  deduction: number;
  netTotal: number;
  status: PayrollStatus;
  month: number;
  year: number;
};

export type PayrollSummary = {
  totalPayrollNet: number;
  totalLateDeduction: number;
  bonusPool: number;
  employees: number;
  pipeline: {
    draft: number;
    approved: number;
    paid: number;
  };
};

export const payrollApi = {
  getPayrolls: async (params: { month?: number; year?: number; q?: string; department_id?: string }) => {
    const queryParams: any = { ...params };
    if (queryParams.month === 0) delete queryParams.month;
    if (queryParams.year === 0) delete queryParams.year;
    
    const res = await apiClient.get<PayrollRecord[]>("/payrolls", { params: queryParams });
    return res.data;
  },

  getPayrollDetail: async (id: string) => {
    const res = await apiClient.get<{ payroll: any; user: any }>(`/payrolls/${id}`);
    return res.data;
  },

  generatePayrolls: async (month: number, year: number) => {
    const res = await apiClient.post("/payrolls/generate", { month, year });
    return res.data;
  },

  // Mock for dashboard summary since we haven't implemented a specific summary endpoint
  // We can calculate this from getPayrolls or add a new endpoint later
  getSummary: async (month: number, year: number) => {
    const payrolls = await payrollApi.getPayrolls({ month, year });
    
    const summary: PayrollSummary = {
      totalPayrollNet: 0,
      totalLateDeduction: 0,
      bonusPool: 0,
      employees: 0,
      pipeline: { draft: 0, approved: 0, paid: 0 }
    };

    const uniqueUsers = new Set();
    payrolls.forEach(p => {
      summary.totalPayrollNet += p.netTotal;
      summary.totalLateDeduction += p.deduction;
      summary.bonusPool += p.bonus10;
      summary.pipeline[p.status]++;
      uniqueUsers.add(p.user_id);
    });
    summary.employees = uniqueUsers.size;

    return summary;
  }
};
