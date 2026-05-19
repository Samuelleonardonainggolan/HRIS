export type Period = {
  monthLabel: string;
  month: number;
  year: number;
  value: string; // YYYY-MM
  label: string; // Alternative label for some components
};

export function generateLastMonths(count: number, includeAll: boolean = false): Period[] {
  const periods: Period[] = [];
  
  if (includeAll) {
    periods.push({
      monthLabel: "Semua Periode",
      month: 0,
      year: 0,
      value: "all",
      label: "Semua Periode",
    });
  }

  const now = new Date();

  for (let i = 0; i < count; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const monthLabel = d.toLocaleString("en-US", { month: "long", year: "numeric" });
    const month = d.getMonth() + 1;
    const year = d.getFullYear();
    const value = `${year}-${month.toString().padStart(2, "0")}`;

    periods.push({
      monthLabel,
      month,
      year,
      value,
      label: monthLabel,
    });
  }

  return periods;
}
