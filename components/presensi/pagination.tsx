import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type Props = {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

function clampPage(page: number, totalPages: number) {
  return Math.min(Math.max(page, 1), Math.max(totalPages, 1));
}

function getVisiblePages(page: number, totalPages: number) {
  const maxPages = Math.min(totalPages, 3);
  if (maxPages <= 0) return [];
  if (totalPages <= 3) return Array.from({ length: totalPages }, (_, i) => i + 1);

  const start = clampPage(page - 1, totalPages);
  const pages = [start, start + 1, start + 2].filter((p) => p >= 1 && p <= totalPages);
  if (pages.length === 3) return pages;
  if (pages[0] === 1) return [1, 2, 3];
  return [totalPages - 2, totalPages - 1, totalPages];
}

export function Pagination({ page, totalPages, onPageChange }: Props) {
  const safeTotal = Math.max(totalPages, 1);
  const safePage = clampPage(page, safeTotal);
  const visiblePages = getVisiblePages(safePage, safeTotal);

  return (
    <div className="flex items-center gap-2">
      <Button
        type="button"
        variant="outline"
        size="icon"
        className="h-8 w-8 rounded-full"
        disabled={safePage <= 1}
        onClick={() => onPageChange(safePage - 1)}
        aria-label="Halaman sebelumnya"
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>
      {visiblePages.map((p) => {
        const active = p === safePage;
        return (
          <button
            key={p}
            type="button"
            onClick={() => onPageChange(p)}
            className={cn(
              "h-8 w-8 rounded-full text-sm font-medium transition-colors",
              active
                ? "bg-blue-600 text-white"
                : "border border-gray-200 bg-white text-gray-700 hover:bg-gray-50"
            )}
            aria-current={active ? "page" : undefined}
          >
            {p}
          </button>
        );
      })}
      <Button
        type="button"
        variant="outline"
        size="icon"
        className="h-8 w-8 rounded-full"
        disabled={safePage >= safeTotal}
        onClick={() => onPageChange(safePage + 1)}
        aria-label="Halaman berikutnya"
      >
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  );
}

