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
import { tokens, ACCENT_COLORS } from '@/lib/tokens';

interface CategoryTrendsChartProps {
  data: CategoryTrendsResponse;
}

// Color palette for categories using design tokens
const CATEGORY_COLORS = ACCENT_COLORS.map(
  (color) => tokens.accent[color].default
);

export function CategoryTrendsChart({ data }: CategoryTrendsChartProps) {
  const { trends } = data;

  // Guard against missing or empty data
  if (!trends?.weeks || trends.weeks.length === 0 || !trends?.categories) {
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

    // Add each category's count (with null guard for week.categories)
    trends.categories.forEach(category => {
      point[category] = week.categories?.[category] ?? 0;
    });

    return point;
  });

  // Calculate total tasks per category for legend
  const categoryTotals: Record<string, number> = {};
  trends.categories.forEach(category => {
    categoryTotals[category] = trends.weeks.reduce((sum, week) => {
      return sum + (week.categories?.[category] ?? 0);
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
                color: 'hsl(var(--card-foreground))',
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
