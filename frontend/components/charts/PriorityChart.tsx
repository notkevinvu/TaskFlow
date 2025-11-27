'use client';

import { PriorityDistribution } from '@/lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend, Cell } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface PriorityChartProps {
  data: PriorityDistribution[];
}

// Colors matching priority levels
const PRIORITY_COLORS: Record<string, string> = {
  'Critical (90-100)': 'hsl(var(--destructive))',
  'High (75-89)': 'hsl(var(--chart-5))',
  'Medium (50-74)': 'hsl(var(--primary))',
  'Low (0-49)': 'hsl(var(--muted))',
};

export function PriorityChart({ data }: PriorityChartProps) {
  // Format data for chart
  const chartData = data.map(d => ({
    range: d.priority_range.replace(/\s*\(\d+-\d+\)/, ''), // Shorten labels
    fullRange: d.priority_range,
    count: d.task_count,
    fill: PRIORITY_COLORS[d.priority_range] || 'hsl(var(--primary))',
  }));

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Priority Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            No priority data available
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Priority Distribution</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="range"
              className="text-xs"
              tick={{ fontSize: 12 }}
            />
            <YAxis
              className="text-xs"
              tick={{ fontSize: 12 }}
              allowDecimals={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
              }}
              formatter={(value: number, name: string, props: any) => [
                `${value} tasks`,
                props.payload.fullRange,
              ]}
            />
            <Legend />
            <Bar dataKey="count" name="Task Count" radius={[8, 8, 0, 0]}>
              {chartData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.fill} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
