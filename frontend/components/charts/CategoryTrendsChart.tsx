'use client';

import { CategoryTrendsResponse } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';

interface CategoryTrendsChartProps {
  data: CategoryTrendsResponse;
}

// Color palette for categories
const CATEGORY_COLORS = [
  'hsl(var(--primary))',
  'hsl(210, 80%, 55%)',
  'hsl(150, 60%, 45%)',
  'hsl(45, 90%, 55%)',
  'hsl(350, 70%, 55%)',
  'hsl(280, 60%, 55%)',
  'hsl(30, 80%, 55%)',
  'hsl(180, 50%, 45%)',
];

export function CategoryTrendsChart({ data }: CategoryTrendsChartProps) {
  const { trends } = data;

  if (!trends.weeks || trends.weeks.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Category Trends</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            No category data available for this period
          </div>
        </CardContent>
      </Card>
    );
  }

  // Transform data for recharts
  const chartData = trends.weeks.map(week => {
    const point: Record<string, string | number> = {
      week: new Date(week.week_start).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    };

    // Add each category's count
    trends.categories.forEach(category => {
      point[category] = week.categories[category] || 0;
    });

    return point;
  });

  // Calculate total tasks per category for legend
  const categoryTotals: Record<string, number> = {};
  trends.categories.forEach(category => {
    categoryTotals[category] = trends.weeks.reduce((sum, week) => {
      return sum + (week.categories[category] || 0);
    }, 0);
  });

  // Sort categories by total (descending) and take top 8
  const sortedCategories = [...trends.categories]
    .sort((a, b) => categoryTotals[b] - categoryTotals[a])
    .slice(0, 8);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Category Trends</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={350}>
          <AreaChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="week"
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
            />
            <Legend />
            {sortedCategories.map((category, index) => (
              <Area
                key={category}
                type="monotone"
                dataKey={category}
                stackId="1"
                stroke={CATEGORY_COLORS[index % CATEGORY_COLORS.length]}
                fill={CATEGORY_COLORS[index % CATEGORY_COLORS.length]}
                fillOpacity={0.6}
              />
            ))}
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
