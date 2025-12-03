'use client';

import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Search, X } from "lucide-react";

export interface ArchiveFiltersState {
  search: string;
  category: string;
  dateRange: 'all' | '7d' | '30d' | '90d' | 'year';
}

interface ArchiveFiltersProps {
  filters: ArchiveFiltersState;
  onChange: (filters: ArchiveFiltersState) => void;
  onClear: () => void;
  categories: string[];
}

export function ArchiveFilters({ filters, onChange, onClear, categories }: ArchiveFiltersProps) {
  const hasActiveFilters = filters.search || filters.category || filters.dateRange !== 'all';

  return (
    <div className="flex flex-col sm:flex-row gap-3">
      {/* Search Input */}
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search completed tasks..."
          value={filters.search}
          onChange={(e) => onChange({ ...filters, search: e.target.value })}
          className="pl-9"
        />
      </div>

      {/* Category Filter */}
      <Select
        value={filters.category || 'all'}
        onValueChange={(value) => onChange({ ...filters, category: value === 'all' ? '' : value })}
      >
        <SelectTrigger className="w-[180px]">
          <SelectValue placeholder="All Categories" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Categories</SelectItem>
          {categories.map((cat) => (
            <SelectItem key={cat} value={cat}>
              {cat}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Date Range Filter */}
      <Select
        value={filters.dateRange}
        onValueChange={(value) => onChange({ ...filters, dateRange: value as ArchiveFiltersState['dateRange'] })}
      >
        <SelectTrigger className="w-[150px]">
          <SelectValue placeholder="Date Range" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Time</SelectItem>
          <SelectItem value="7d">Last 7 days</SelectItem>
          <SelectItem value="30d">Last 30 days</SelectItem>
          <SelectItem value="90d">Last 90 days</SelectItem>
          <SelectItem value="year">Last year</SelectItem>
        </SelectContent>
      </Select>

      {/* Clear Filters Button */}
      {hasActiveFilters && (
        <Button variant="ghost" size="icon" onClick={onClear} title="Clear filters">
          <X className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
