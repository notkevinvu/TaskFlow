'use client';

import { HeatmapResponse } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

interface ProductivityHeatmapProps {
  data: HeatmapResponse;
}

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

// Get display hour label (show every 3 hours)
function getHourLabel(hour: number): string {
  if (hour % 3 !== 0) return '';
  if (hour === 0) return '12am';
  if (hour === 12) return '12pm';
  if (hour < 12) return `${hour}am`;
  return `${hour - 12}pm`;
}

// Get color intensity based on count
function getColorIntensity(count: number, maxCount: number): string {
  if (count === 0 || maxCount === 0) return 'bg-muted/30';

  const intensity = count / maxCount;
  if (intensity < 0.2) return 'bg-primary/20';
  if (intensity < 0.4) return 'bg-primary/40';
  if (intensity < 0.6) return 'bg-primary/60';
  if (intensity < 0.8) return 'bg-primary/80';
  return 'bg-primary';
}

export function ProductivityHeatmap({ data }: ProductivityHeatmapProps) {
  const { heatmap } = data;

  // Guard against missing heatmap data
  if (!heatmap?.cells) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Productivity Heatmap</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-48 flex items-center justify-center text-muted-foreground">
            No heatmap data available
          </div>
        </CardContent>
      </Card>
    );
  }

  // Create a lookup map for quick access
  const cellMap = new Map<string, number>();
  heatmap.cells.forEach(cell => {
    cellMap.set(`${cell.day_of_week}-${cell.hour}`, cell.count);
  });

  // Calculate total completions
  const totalCompletions = heatmap.cells.reduce((sum, cell) => sum + cell.count, 0);

  if (totalCompletions === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Productivity Heatmap</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-48 flex items-center justify-center text-muted-foreground">
            No completion data available for this period
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Productivity Heatmap</span>
          <span className="text-sm font-normal text-muted-foreground">
            {totalCompletions} tasks completed
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <TooltipProvider>
          <div className="overflow-x-auto">
            <div className="min-w-[600px]">
              {/* Hour labels */}
              <div className="flex mb-1">
                <div className="w-10" /> {/* Spacer for day labels */}
                {HOURS.map(hour => (
                  <div key={hour} className="flex-1 text-xs text-muted-foreground text-center">
                    {getHourLabel(hour)}
                  </div>
                ))}
              </div>

              {/* Grid */}
              {DAYS.map((day, dayIndex) => (
                <div key={day} className="flex items-center gap-1 mb-1">
                  <div className="w-10 text-xs text-muted-foreground">{day}</div>
                  {HOURS.map(hour => {
                    const count = cellMap.get(`${dayIndex}-${hour}`) || 0;
                    return (
                      <Tooltip key={hour} delayDuration={100}>
                        <TooltipTrigger asChild>
                          <div
                            className={`flex-1 h-5 rounded-sm ${getColorIntensity(count, heatmap.max_count)} transition-colors hover:ring-1 hover:ring-primary cursor-default`}
                          />
                        </TooltipTrigger>
                        <TooltipContent>
                          <p className="font-medium">{day} {hour}:00</p>
                          <p className="text-muted-foreground">
                            {count} task{count !== 1 ? 's' : ''} completed
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    );
                  })}
                </div>
              ))}

              {/* Legend */}
              <div className="flex items-center justify-end gap-2 mt-4 text-xs text-muted-foreground">
                <span>Less</span>
                <div className="flex gap-1">
                  <div className="w-4 h-4 rounded-sm bg-muted/30" />
                  <div className="w-4 h-4 rounded-sm bg-primary/20" />
                  <div className="w-4 h-4 rounded-sm bg-primary/40" />
                  <div className="w-4 h-4 rounded-sm bg-primary/60" />
                  <div className="w-4 h-4 rounded-sm bg-primary/80" />
                  <div className="w-4 h-4 rounded-sm bg-primary" />
                </div>
                <span>More</span>
              </div>
            </div>
          </div>
        </TooltipProvider>
      </CardContent>
    </Card>
  );
}
