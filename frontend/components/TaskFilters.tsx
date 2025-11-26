'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Filter, X, ChevronDown, ChevronUp } from 'lucide-react';
import { useTasks } from '@/hooks/useTasks';

export interface TaskFilterState {
  status?: string;
  category?: string;
  minPriority?: number;
  maxPriority?: number;
}

interface TaskFiltersProps {
  filters: TaskFilterState;
  onChange: (filters: TaskFilterState) => void;
  onClear: () => void;
}

export function TaskFilters({ filters, onChange, onClear }: TaskFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const { data: tasksData } = useTasks();

  // Extract unique categories from tasks
  const categories = Array.from(
    new Set(
      tasksData?.tasks
        ?.map((task) => task.category)
        .filter((cat): cat is string => !!cat) || []
    )
  ).sort();

  const updateFilter = (key: keyof TaskFilterState, value: any) => {
    onChange({ ...filters, [key]: value });
  };

  const removeFilter = (key: keyof TaskFilterState) => {
    const newFilters = { ...filters };
    delete newFilters[key];
    onChange(newFilters);
  };

  const activeFilterCount = Object.keys(filters).length;

  const priorityRanges = [
    { label: 'Critical (90-100)', min: 90, max: 100 },
    { label: 'High (75-89)', min: 75, max: 89 },
    { label: 'Medium (50-74)', min: 50, max: 74 },
    { label: 'Low (0-49)', min: 0, max: 49 },
  ];

  return (
    <div className="space-y-3">
      {/* Filter Toggle Button */}
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex items-center gap-2"
        >
          <Filter className="h-4 w-4" />
          Filters
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-1">
              {activeFilterCount}
            </Badge>
          )}
          {isExpanded ? (
            <ChevronUp className="h-4 w-4 ml-1" />
          ) : (
            <ChevronDown className="h-4 w-4 ml-1" />
          )}
        </Button>

        {activeFilterCount > 0 && (
          <Button variant="ghost" size="sm" onClick={onClear}>
            Clear all
          </Button>
        )}
      </div>

      {/* Active Filter Chips */}
      {activeFilterCount > 0 && (
        <div className="flex flex-wrap gap-2">
          {filters.status && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Status: {filters.status}
              <button
                onClick={() => removeFilter('status')}
                className="ml-1 hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {filters.category && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Category: {filters.category}
              <button
                onClick={() => removeFilter('category')}
                className="ml-1 hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {(filters.minPriority !== undefined || filters.maxPriority !== undefined) && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Priority: {filters.minPriority ?? 0}-{filters.maxPriority ?? 100}
              <button
                onClick={() => {
                  removeFilter('minPriority');
                  removeFilter('maxPriority');
                }}
                className="ml-1 hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
        </div>
      )}

      {/* Filter Panel */}
      {isExpanded && (
        <div className="grid gap-4 p-4 border rounded-lg bg-card md:grid-cols-3">
          {/* Status Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Status</label>
            <Select
              value={filters.status || ''}
              onValueChange={(value) =>
                value ? updateFilter('status', value) : removeFilter('status')
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All statuses</SelectItem>
                <SelectItem value="todo">To Do</SelectItem>
                <SelectItem value="in_progress">In Progress</SelectItem>
                <SelectItem value="done">Done</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Category Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Category</label>
            <Select
              value={filters.category || ''}
              onValueChange={(value) =>
                value ? updateFilter('category', value) : removeFilter('category')
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="All categories" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All categories</SelectItem>
                {categories.length === 0 ? (
                  <div className="px-2 py-6 text-center text-sm text-muted-foreground">
                    No categories yet
                  </div>
                ) : (
                  categories.map((category) => (
                    <SelectItem key={category} value={category}>
                      {category}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Priority Range Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Priority Range</label>
            <Select
              value={
                filters.minPriority !== undefined
                  ? `${filters.minPriority}-${filters.maxPriority}`
                  : ''
              }
              onValueChange={(value) => {
                if (value === '__all__') {
                  removeFilter('minPriority');
                  removeFilter('maxPriority');
                } else {
                  const [min, max] = value.split('-').map(Number);
                  updateFilter('minPriority', min);
                  updateFilter('maxPriority', max);
                }
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="All priorities" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All priorities</SelectItem>
                {priorityRanges.map((range) => (
                  <SelectItem key={range.label} value={`${range.min}-${range.max}`}>
                    {range.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      )}
    </div>
  );
}
