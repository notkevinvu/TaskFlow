'use client';

import { useState, useMemo, useCallback } from 'react';
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
  getFilters: () => TaskFilterState; // Function to get fresh dates
}

// Parse date string as local date (avoids UTC timezone shift)
function parseLocalDate(dateStr: string): Date {
  const [year, month, day] = dateStr.split('-').map(Number);
  return new Date(year, month - 1, day);
}

// Safe date formatting that handles invalid dates
function safeFormatDate(dateStr: string | undefined, formatStr: string, fallback: string = '...'): string {
  if (!dateStr) return fallback;
  try {
    const date = parseLocalDate(dateStr);
    if (isNaN(date.getTime())) return fallback;
    return format(date, formatStr);
  } catch {
    return fallback;
  }
}

// Generate presets with fresh dates (called at render time)
function getFilterPresets(): FilterPreset[] {
  return [
    {
      id: 'high-priority',
      label: 'High Priority',
      getFilters: () => ({ minPriority: 75, maxPriority: 100 }),
    },
    {
      id: 'critical',
      label: 'Critical Only',
      getFilters: () => ({ minPriority: 90, maxPriority: 100 }),
    },
    {
      id: 'due-this-week',
      label: 'Due This Week',
      getFilters: () => {
        const today = new Date();
        const endOfWeek = new Date();
        endOfWeek.setDate(endOfWeek.getDate() + (7 - endOfWeek.getDay()));
        return {
          dueDateStart: format(today, 'yyyy-MM-dd'),
          dueDateEnd: format(endOfWeek, 'yyyy-MM-dd'),
        };
      },
    },
    {
      id: 'overdue',
      label: 'Overdue',
      getFilters: () => {
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        return {
          dueDateStart: '2000-01-01',
          dueDateEnd: format(yesterday, 'yyyy-MM-dd'),
        };
      },
    },
    {
      id: 'in-progress',
      label: 'In Progress',
      getFilters: () => ({ status: 'in_progress' }),
    },
  ];
}

// Check if two filter states are equal
function filtersEqual(a: TaskFilterState, b: TaskFilterState): boolean {
  return (
    a.status === b.status &&
    a.category === b.category &&
    a.minPriority === b.minPriority &&
    a.maxPriority === b.maxPriority &&
    a.dueDateStart === b.dueDateStart &&
    a.dueDateEnd === b.dueDateEnd
  );
}

interface TaskFiltersProps {
  filters: TaskFilterState;
  onChange: (filters: TaskFilterState) => void;
  onClear: () => void;
  availableCategories: string[]; // Categories passed from parent to avoid duplicate fetch
}

export function TaskFilters({ filters, onChange, onClear, availableCategories }: TaskFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  // Pending filters - local state that only applies on "Apply" click
  const [pendingFilters, setPendingFilters] = useState<TaskFilterState>(filters);

  // Date picker popover state
  const [datePickerOpen, setDatePickerOpen] = useState(false);
  const [pendingDateRange, setPendingDateRange] = useState<DateRange | undefined>(undefined);

  // Memoize preset definitions - dates are computed fresh when getFilters() is called
  const presets = useMemo(() => getFilterPresets(), []);

  // Handle panel expand/collapse - sync pending filters when opening
  const handleExpandToggle = useCallback(() => {
    setIsExpanded(prev => {
      if (!prev) {
        // Opening the panel - sync pending filters from applied filters
        setPendingFilters(filters);
      }
      return !prev;
    });
  }, [filters]);

  // Handle date picker open/close - initialize pending date range when opening
  const handleDatePickerOpenChange = useCallback((open: boolean) => {
    if (open) {
      // Initialize pending date range from current pending filters
      // Note: We read pendingFilters here, which may be stale in the callback,
      // but this is acceptable since we're syncing from the current pending state
      const range: DateRange | undefined =
        pendingFilters.dueDateStart || pendingFilters.dueDateEnd
          ? {
              from: pendingFilters.dueDateStart ? parseLocalDate(pendingFilters.dueDateStart) : undefined,
              to: pendingFilters.dueDateEnd ? parseLocalDate(pendingFilters.dueDateEnd) : undefined,
            }
          : undefined;
      setPendingDateRange(range);
    }
    setDatePickerOpen(open);
  }, [pendingFilters.dueDateStart, pendingFilters.dueDateEnd]);

  const updatePendingFilter = useCallback((key: keyof TaskFilterState, value: string | number | undefined) => {
    setPendingFilters(prev => {
      if (value === undefined) {
        const newFilters = { ...prev };
        delete newFilters[key];
        return newFilters;
      }
      return { ...prev, [key]: value };
    });
  }, []);

  const removePendingFilter = useCallback((key: keyof TaskFilterState) => {
    setPendingFilters(prev => {
      const newFilters = { ...prev };
      delete newFilters[key];
      return newFilters;
    });
  }, []);

  const applyPreset = useCallback((preset: FilterPreset) => {
    // Replace all pending filters with preset
    setPendingFilters(preset.getFilters());
  }, []);

  // Handle date selection in calendar - supports deselection
  const handleDateSelect = useCallback((range: DateRange | undefined) => {
    if (!range) {
      setPendingDateRange(undefined);
      return;
    }

    setPendingDateRange(prev => {
      // Handle deselection: if clicking on already selected start or end date
      if (prev) {
        const clickedDate = range.from;
        if (clickedDate) {
          const clickedTime = clickedDate.getTime();

          // If clicking on the start date, clear it
          if (prev.from && clickedTime === prev.from.getTime() && !range.to) {
            return { from: undefined, to: prev.to };
          }

          // If clicking on the end date, clear it
          if (prev.to && clickedTime === prev.to.getTime()) {
            return { from: prev.from, to: undefined };
          }
        }
      }

      return range;
    });
  }, []);

  // Confirm date range selection
  const confirmDateRange = useCallback(() => {
    setPendingFilters(prev => {
      const newFilters = { ...prev };
      if (pendingDateRange?.from) {
        newFilters.dueDateStart = format(pendingDateRange.from, 'yyyy-MM-dd');
      } else {
        delete newFilters.dueDateStart;
      }
      if (pendingDateRange?.to) {
        newFilters.dueDateEnd = format(pendingDateRange.to, 'yyyy-MM-dd');
      } else {
        delete newFilters.dueDateEnd;
      }
      return newFilters;
    });
    setDatePickerOpen(false);
  }, [pendingDateRange]);

  // Reset date range in picker
  const resetDateRange = useCallback(() => {
    setPendingDateRange(undefined);
  }, []);

  // Clear date range from pending filters
  const clearPendingDateRange = useCallback(() => {
    setPendingFilters(prev => {
      const newFilters = { ...prev };
      delete newFilters.dueDateStart;
      delete newFilters.dueDateEnd;
      return newFilters;
    });
  }, []);

  // Apply all pending filters
  const applyFilters = useCallback(() => {
    onChange(pendingFilters);
  }, [onChange, pendingFilters]);

  // Reset pending filters to empty
  const resetFilters = useCallback(() => {
    setPendingFilters({});
  }, []);

  // Helper to remove applied filters (used by filter chips)
  const removeAppliedFilter = useCallback((...keys: (keyof TaskFilterState)[]) => {
    const newFilters = { ...filters };
    keys.forEach(key => delete newFilters[key]);
    onChange(newFilters);
  }, [filters, onChange]);

  // Check if there are pending changes
  const hasPendingChanges = !filtersEqual(pendingFilters, filters);

  // Count of currently applied filters (shown in badge)
  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (filters.status) count++;
    if (filters.category) count++;
    if (filters.minPriority !== undefined && filters.maxPriority !== undefined) count++;
    if (filters.dueDateStart || filters.dueDateEnd) count++;
    return count;
  }, [filters]);

  // Count of pending filter selections
  const pendingFilterCount = useMemo(() => {
    let count = 0;
    if (pendingFilters.status) count++;
    if (pendingFilters.category) count++;
    if (pendingFilters.minPriority !== undefined && pendingFilters.maxPriority !== undefined) count++;
    if (pendingFilters.dueDateStart || pendingFilters.dueDateEnd) count++;
    return count;
  }, [pendingFilters]);

  const priorityRanges = [
    { label: 'Critical (90-100)', min: 90, max: 100 },
    { label: 'High (75-89)', min: 75, max: 89 },
    { label: 'Medium (50-74)', min: 50, max: 74 },
    { label: 'Low (0-49)', min: 0, max: 49 },
  ];

  // Check if date range is complete (both dates selected)
  const isDateRangeComplete = Boolean(pendingDateRange?.from && pendingDateRange?.to);

  return (
    <div className="space-y-3">
      {/* Filter Toggle Button */}
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          onClick={handleExpandToggle}
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

      {/* Active Filter Chips (showing applied filters) */}
      {activeFilterCount > 0 && (
        <div className="flex flex-wrap gap-2">
          {filters.status && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Status: {filters.status}
              <button
                onClick={() => removeAppliedFilter('status')}
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
                onClick={() => removeAppliedFilter('category')}
                className="ml-1 hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {filters.minPriority !== undefined && filters.maxPriority !== undefined && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Priority: {filters.minPriority}-{filters.maxPriority}
              <button
                onClick={() => removeAppliedFilter('minPriority', 'maxPriority')}
                className="ml-1 hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {(filters.dueDateStart || filters.dueDateEnd) && (
            <Badge variant="secondary" className="flex items-center gap-1">
              Due: {safeFormatDate(filters.dueDateStart, 'MMM d')}
              {' - '}
              {safeFormatDate(filters.dueDateEnd, 'MMM d')}
              <button
                onClick={() => removeAppliedFilter('dueDateStart', 'dueDateEnd')}
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
              {presets.map((preset) => (
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
                value={pendingFilters.status || ''}
                onValueChange={(value) =>
                  value && value !== '__all__'
                    ? updatePendingFilter('status', value)
                    : removePendingFilter('status')
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
                value={pendingFilters.category || ''}
                onValueChange={(value) =>
                  value && value !== '__all__'
                    ? updatePendingFilter('category', value)
                    : removePendingFilter('category')
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
                  pendingFilters.minPriority !== undefined && pendingFilters.maxPriority !== undefined
                    ? `${pendingFilters.minPriority}-${pendingFilters.maxPriority}`
                    : ''
                }
                onValueChange={(value) => {
                  if (value === '__all__') {
                    const newFilters = { ...pendingFilters };
                    delete newFilters.minPriority;
                    delete newFilters.maxPriority;
                    setPendingFilters(newFilters);
                  } else {
                    const [min, max] = value.split('-').map(Number);
                    setPendingFilters({ ...pendingFilters, minPriority: min, maxPriority: max });
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
              <Popover open={datePickerOpen} onOpenChange={handleDatePickerOpenChange}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className="w-full justify-start text-left font-normal"
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {pendingFilters.dueDateStart || pendingFilters.dueDateEnd ? (
                      <>
                        {safeFormatDate(pendingFilters.dueDateStart, 'MMM d')}
                        {' - '}
                        {safeFormatDate(pendingFilters.dueDateEnd, 'MMM d')}
                      </>
                    ) : (
                      <span className="text-muted-foreground">Pick a date range</span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="range"
                    selected={pendingDateRange}
                    onSelect={handleDateSelect}
                    numberOfMonths={2}
                  />
                  <div className="p-3 border-t space-y-2">
                    {/* Selection status */}
                    <p className="text-xs text-muted-foreground text-center">
                      {!pendingDateRange?.from && !pendingDateRange?.to && 'Select start and end dates'}
                      {pendingDateRange?.from && !pendingDateRange?.to && 'Now select an end date'}
                      {pendingDateRange?.from && pendingDateRange?.to && 'Date range selected'}
                    </p>
                    {/* Action buttons */}
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={resetDateRange}
                        className="flex-1"
                        disabled={!pendingDateRange?.from && !pendingDateRange?.to}
                      >
                        Reset
                      </Button>
                      <Button
                        size="sm"
                        onClick={confirmDateRange}
                        className="flex-1"
                        disabled={!isDateRangeComplete}
                      >
                        Confirm
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
              {/* Clear date range button (when dates are set) */}
              {(pendingFilters.dueDateStart || pendingFilters.dueDateEnd) && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearPendingDateRange}
                  className="w-full h-7 text-xs"
                >
                  Clear dates
                </Button>
              )}
            </div>
          </div>

          {/* Apply/Reset Buttons */}
          <div className="flex justify-end gap-2 pt-2 border-t">
            <Button
              variant="outline"
              size="sm"
              onClick={resetFilters}
              disabled={pendingFilterCount === 0}
            >
              Reset
            </Button>
            <Button
              size="sm"
              onClick={applyFilters}
              disabled={!hasPendingChanges}
            >
              Apply{pendingFilterCount > 0 && ` (${pendingFilterCount})`}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
