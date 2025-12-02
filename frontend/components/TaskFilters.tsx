'use client';

import { useState } from 'react';
import { format } from 'date-fns';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Calendar } from '@/components/ui/calendar';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Filter, X, ChevronDown, ChevronUp, CalendarIcon, Zap } from 'lucide-react';
import type { DateRange } from 'react-day-picker';

export interface TaskFilterState {
  status?: string;
  category?: string;
  minPriority?: number;
  maxPriority?: number;
  dueDateStart?: string; // YYYY-MM-DD
  dueDateEnd?: string;   // YYYY-MM-DD
}

// Filter presets for quick access
export interface FilterPreset {
  id: string;
  label: string;
  filters: TaskFilterState;
}

export const FILTER_PRESETS: FilterPreset[] = [
  {
    id: 'high-priority',
    label: 'High Priority',
    filters: { minPriority: 75, maxPriority: 100 },
  },
  {
    id: 'critical',
    label: 'Critical Only',
    filters: { minPriority: 90, maxPriority: 100 },
  },
  {
    id: 'due-this-week',
    label: 'Due This Week',
    filters: {
      dueDateStart: (() => {
        const today = new Date();
        return format(today, 'yyyy-MM-dd');
      })(),
      dueDateEnd: (() => {
        const endOfWeek = new Date();
        endOfWeek.setDate(endOfWeek.getDate() + (7 - endOfWeek.getDay()));
        return format(endOfWeek, 'yyyy-MM-dd');
      })(),
    },
  },
  {
    id: 'overdue',
    label: 'Overdue',
    filters: {
      dueDateStart: '2000-01-01',
      dueDateEnd: (() => {
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        return format(yesterday, 'yyyy-MM-dd');
      })(),
    },
  },
  {
    id: 'in-progress',
    label: 'In Progress',
    filters: { status: 'in_progress' },
  },
];

interface TaskFiltersProps {
  filters: TaskFilterState;
  onChange: (filters: TaskFilterState) => void;
  onClear: () => void;
  availableCategories: string[]; // Categories passed from parent to avoid duplicate fetch
}

export function TaskFilters({ filters, onChange, onClear, availableCategories }: TaskFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const updateFilter = (key: keyof TaskFilterState, value: string | number) => {
    onChange({ ...filters, [key]: value });
  };

  const removeFilter = (key: keyof TaskFilterState) => {
    const newFilters = { ...filters };
    delete newFilters[key];
    onChange(newFilters);
  };

  const applyPreset = (preset: FilterPreset) => {
    onChange({ ...filters, ...preset.filters });
  };

  // Convert filter date strings to Date objects for the calendar
  const dateRange: DateRange | undefined =
    filters.dueDateStart || filters.dueDateEnd
      ? {
          from: filters.dueDateStart ? new Date(filters.dueDateStart) : undefined,
          to: filters.dueDateEnd ? new Date(filters.dueDateEnd) : undefined,
        }
      : undefined;

  const handleDateRangeChange = (range: DateRange | undefined) => {
    const newFilters = { ...filters };
    if (range?.from) {
      newFilters.dueDateStart = format(range.from, 'yyyy-MM-dd');
    } else {
      delete newFilters.dueDateStart;
    }
    if (range?.to) {
      newFilters.dueDateEnd = format(range.to, 'yyyy-MM-dd');
    } else {
      delete newFilters.dueDateEnd;
    }
    onChange(newFilters);
  };

  const clearDateRange = () => {
    const newFilters = { ...filters };
    delete newFilters.dueDateStart;
    delete newFilters.dueDateEnd;
    onChange(newFilters);
  };

  // Count active filters (date range counts as 1)
  const activeFilterCount = Object.keys(filters).filter(
    (key) => key !== 'dueDateEnd' || !filters.dueDateStart
  ).length - (filters.dueDateStart && filters.dueDateEnd ? 1 : 0);

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
          {(filters.dueDateStart || filters.dueDateEnd) && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Due: {filters.dueDateStart ? format(new Date(filters.dueDateStart), 'MMM d') : '...'}
              {' - '}
              {filters.dueDateEnd ? format(new Date(filters.dueDateEnd), 'MMM d') : '...'}
              <button
                onClick={clearDateRange}
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
        <div className="space-y-4 p-4 border rounded-lg bg-card">
          {/* Quick Presets */}
          <div className="space-y-2">
            <label className="text-sm font-medium flex items-center gap-2">
              <Zap className="h-4 w-4" />
              Quick Filters
            </label>
            <div className="flex flex-wrap gap-2">
              {FILTER_PRESETS.map((preset) => (
                <Button
                  key={preset.id}
                  variant="outline"
                  size="sm"
                  onClick={() => applyPreset(preset)}
                  className="h-7 text-xs"
                >
                  {preset.label}
                </Button>
              ))}
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {/* Status Filter */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <Select
                value={filters.status || ''}
                onValueChange={(value) =>
                  value && value !== '__all__' ? updateFilter('status', value) : removeFilter('status')
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
                  value && value !== '__all__' ? updateFilter('category', value) : removeFilter('category')
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All categories" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">All categories</SelectItem>
                  {availableCategories.length === 0 ? (
                    <div className="px-2 py-6 text-center text-sm text-muted-foreground">
                      No categories yet
                    </div>
                  ) : (
                    availableCategories.map((category) => (
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

            {/* Date Range Filter */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Due Date Range</label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className="w-full justify-start text-left font-normal"
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dateRange?.from ? (
                      dateRange.to ? (
                        <>
                          {format(dateRange.from, 'MMM d')} - {format(dateRange.to, 'MMM d')}
                        </>
                      ) : (
                        format(dateRange.from, 'MMM d, yyyy')
                      )
                    ) : (
                      <span className="text-muted-foreground">Pick a date range</span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="range"
                    selected={dateRange}
                    onSelect={handleDateRangeChange}
                    numberOfMonths={2}
                  />
                  {dateRange && (
                    <div className="p-3 border-t">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={clearDateRange}
                        className="w-full"
                      >
                        Clear dates
                      </Button>
                    </div>
                  )}
                </PopoverContent>
              </Popover>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
