'use client';

import { useState, memo } from 'react';
import { Task } from '@/lib/api';
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import { Button } from "@/components/ui/button";

type SortField = 'title' | 'completed_at' | 'category' | 'bump_count';
type SortDirection = 'asc' | 'desc';

interface SortIconProps {
  field: SortField;
  currentSortField: SortField;
  sortDirection: SortDirection;
}

function SortIcon({ field, currentSortField, sortDirection }: SortIconProps) {
  if (currentSortField !== field) {
    return <ArrowUpDown className="h-4 w-4 ml-1" />;
  }
  return sortDirection === 'asc' ? (
    <ArrowUp className="h-4 w-4 ml-1" />
  ) : (
    <ArrowDown className="h-4 w-4 ml-1" />
  );
}

interface ArchiveTableProps {
  tasks: Task[];
  selectedIds: Set<string>;
  onSelectionChange: (selectedIds: Set<string>) => void;
  onTaskClick?: (task: Task) => void;
}

export const ArchiveTable = memo(function ArchiveTable({ tasks, selectedIds, onSelectionChange, onTaskClick }: ArchiveTableProps) {
  const [sortField, setSortField] = useState<SortField>('completed_at');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedTasks = [...tasks].sort((a, b) => {
    let comparison = 0;
    switch (sortField) {
      case 'title':
        comparison = a.title.localeCompare(b.title);
        break;
      case 'completed_at':
        const dateA = new Date(a.updated_at).getTime();
        const dateB = new Date(b.updated_at).getTime();
        comparison = dateA - dateB;
        break;
      case 'category':
        comparison = (a.category || '').localeCompare(b.category || '');
        break;
      case 'bump_count':
        comparison = a.bump_count - b.bump_count;
        break;
    }
    return sortDirection === 'asc' ? comparison : -comparison;
  });

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      onSelectionChange(new Set(tasks.map(t => t.id)));
    } else {
      onSelectionChange(new Set());
    }
  };

  const handleSelectOne = (taskId: string, checked: boolean) => {
    const newSelection = new Set(selectedIds);
    if (checked) {
      newSelection.add(taskId);
    } else {
      newSelection.delete(taskId);
    }
    onSelectionChange(newSelection);
  };

  const allSelected = tasks.length > 0 && selectedIds.size === tasks.length;
  const someSelected = selectedIds.size > 0 && selectedIds.size < tasks.length;

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[50px]">
              <Checkbox
                checked={allSelected}
                onCheckedChange={handleSelectAll}
                aria-label="Select all"
                className={someSelected ? "data-[state=checked]:bg-primary/50" : ""}
              />
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                className="p-0 h-auto font-medium hover:bg-transparent"
                onClick={() => handleSort('title')}
              >
                Title
                <SortIcon field="title" currentSortField={sortField} sortDirection={sortDirection} />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                className="p-0 h-auto font-medium hover:bg-transparent"
                onClick={() => handleSort('category')}
              >
                Category
                <SortIcon field="category" currentSortField={sortField} sortDirection={sortDirection} />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                className="p-0 h-auto font-medium hover:bg-transparent"
                onClick={() => handleSort('completed_at')}
              >
                Completed
                <SortIcon field="completed_at" currentSortField={sortField} sortDirection={sortDirection} />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                className="p-0 h-auto font-medium hover:bg-transparent"
                onClick={() => handleSort('bump_count')}
              >
                Bumps
                <SortIcon field="bump_count" currentSortField={sortField} sortDirection={sortDirection} />
              </Button>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedTasks.map((task) => (
            <TableRow
              key={task.id}
              className="cursor-pointer hover:bg-muted/50"
              onClick={() => onTaskClick?.(task)}
            >
              <TableCell onClick={(e) => e.stopPropagation()}>
                <Checkbox
                  checked={selectedIds.has(task.id)}
                  onCheckedChange={(checked) => handleSelectOne(task.id, checked as boolean)}
                  aria-label={`Select ${task.title}`}
                />
              </TableCell>
              <TableCell className="font-medium">{task.title}</TableCell>
              <TableCell>
                {task.category ? (
                  <Badge variant="outline" className="bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300">
                    {task.category}
                  </Badge>
                ) : (
                  <span className="text-muted-foreground">—</span>
                )}
              </TableCell>
              <TableCell>
                {new Date(task.updated_at).toLocaleDateString()}
              </TableCell>
              <TableCell>
                {task.bump_count > 0 ? (
                  <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                    {task.bump_count}x
                  </Badge>
                ) : (
                  <span className="text-muted-foreground">0</span>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
});
