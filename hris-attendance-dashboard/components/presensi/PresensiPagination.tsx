import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

type Props = {
  page: number;
  totalPages: number;
  onChange: (nextPage: number) => void;
};

function clamp(n: number, min: number, max: number) {
  return Math.max(min, Math.min(max, n));
}

export function PresensiPagination({ page, totalPages, onChange }: Props) {
  const safeTotal = Math.max(1, totalPages);
  const safePage = clamp(page, 1, safeTotal);
  const pages = Array.from({ length: Math.min(3, safeTotal) }, (_, i) => i + 1);

  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        className={cn(
          "h-8 w-8 rounded-full border border-gray-200 bg-white text-gray-600 transition-colors",
          safePage === 1 ? "opacity-50" : "hover:bg-gray-50"
        )}
        onClick={() => onChange(Math.max(1, safePage - 1))}
        disabled={safePage === 1}
        aria-label="Halaman sebelumnya"
      >
        <ChevronLeft className="mx-auto h-4 w-4" />
      </button>

      {pages.map((p) => (
        <button
          key={p}
          type="button"
          className={cn(
            "h-8 w-8 rounded-full border border-gray-200 text-sm font-medium transition-colors",
            p === safePage
              ? "border-blue-600 bg-blue-600 text-white"
              : "bg-white text-gray-700 hover:bg-gray-50"
          )}
          onClick={() => onChange(p)}
          aria-current={p === safePage ? "page" : undefined}
        >
          {p}
        </button>
      ))}

      <button
        type="button"
        className={cn(
          "h-8 w-8 rounded-full border border-gray-200 bg-white text-gray-600 transition-colors",
          safePage === safeTotal ? "opacity-50" : "hover:bg-gray-50"
        )}
        onClick={() => onChange(Math.min(safeTotal, safePage + 1))}
        disabled={safePage === safeTotal}
        aria-label="Halaman berikutnya"
      >
        <ChevronRight className="mx-auto h-4 w-4" />
      </button>
    </div>
  );
}

