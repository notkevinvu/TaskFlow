import { Skeleton } from "@/components/ui/skeleton";

export default function AnalyticsLoading() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold">Analytics</h2>
          <p className="text-muted-foreground">
            Insights into your task patterns and productivity
          </p>
        </div>
        <Skeleton className="w-48 h-10" />
      </div>

      {/* Summary Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>

      {/* Charts Grid */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Completion Trends (full width) */}
        <div className="lg:col-span-2">
          <Skeleton className="h-80" />
        </div>

        {/* Priority Distribution */}
        <Skeleton className="h-80" />

        {/* Category Breakdown */}
        <Skeleton className="h-80" />

        {/* Bump Analysis (full width) */}
        <div className="lg:col-span-2">
          <Skeleton className="h-96" />
        </div>
      </div>
    </div>
  );
}
