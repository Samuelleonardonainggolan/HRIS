import { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  iconColor?: string;
  iconBgColor?: string;
  loading?: boolean;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  badge?: {
    text: string;
    variant: "success" | "warning";
  };
  link?: {
    text: string;
    href: string;
  };
}

export function StatsCard({
  title,
  value,
  icon: Icon,
  iconColor = "text-gray-600",
  iconBgColor = "bg-gray-100",
  loading = false,
  trend,
  badge,
  link,
}: StatsCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            {/* Icon dengan background bulat */}
            <div className={cn("flex h-10 w-10 items-center justify-center rounded-full mb-3", iconBgColor)}>
              <Icon className={cn("h-5 w-5", iconColor)} />
            </div>

            {/* Title */}
            <div className="text-sm text-gray-600 mb-2">
              {title}
            </div>

            {/* Value */}
            {loading ? (
              <div className="h-9 w-24 bg-gray-200 animate-pulse rounded mb-2" />
            ) : (
              <div className="text-3xl font-bold text-gray-900 mb-2">
                {value.toLocaleString("id-ID")}
              </div>
            )}

            {/* Trend / Badge / Link */}
            {loading ? (
              <div className="h-4 w-16 bg-gray-100 animate-pulse rounded" />
            ) : (
              <>
                {trend && (
                  <div className="flex items-center gap-1 text-sm">
                    <span
                      className={cn(
                        "font-medium",
                        trend.isPositive ? "text-green-600" : "text-red-600"
                      )}
                    >
                      {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
                    </span>
                  </div>
                )}

                {badge && (
                  <div
                    className={cn(
                      "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold",
                      badge.variant === "success"
                        ? "bg-green-100 text-green-800"
                        : "bg-orange-100 text-orange-800"
                    )}
                  >
                    {badge.text}
                  </div>
                )}

                {link && (
                  <a
                    href={link.href}
                    className="inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-700"
                  >
                    {link.text} →
                  </a>
                )}
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
