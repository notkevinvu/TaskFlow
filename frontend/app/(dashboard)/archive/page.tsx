'use client';

import { useState, useMemo, useCallback } from 'react';
import { useCompletedTasks, useBulkDelete, useBulkRestore } from '@/hooks/useTasks';
import { ArchiveFilters, type ArchiveFiltersState } from '@/components/archive/ArchiveFilters';
import { ArchiveTable } from '@/components/archive/ArchiveTable';
import { BulkActionsBar } from '@/components/archive/BulkActionsBar';
import { TaskDetailsSidebar } from '@/components/TaskDetailsSidebar';
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Archive, ChevronLeft, ChevronRight } from "lucide-react";
import { Task } from '@/lib/api';

const ITEMS_PER_PAGE = 20;

export default function ArchivePage() {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [filters, setFilters] = useState<ArchiveFiltersState>({
    search: '',
    category: '',
    dateRange: 'all',
  });

  const { data: completedTasksData, isLoading, isFetching } = useCompletedTasks();
  const bulkDelete = useBulkDelete();
  const bulkRestore = useBulkRestore();

  // Extract unique categories from completed tasks
  const categories = useMemo(() => {
    if (!completedTasksData?.tasks) return [];
    return Array.from(
      new Set(
        completedTasksData.tasks
          .map((t) => t.category)
          .filter((c): c is string => !!c)
      )
    ).sort();
  }, [completedTasksData?.tasks]);

  // Apply filters to tasks
  const filteredTasks = useMemo(() => {
    if (!completedTasksData?.tasks) return [];

    let tasks = completedTasksData.tasks;

    // Search filter
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      tasks = tasks.filter(
        (t) =>
          t.title.toLowerCase().includes(searchLower) ||
          (t.description && t.description.toLowerCase().includes(searchLower))
      );
    }

    // Category filter
    if (filters.category) {
      tasks = tasks.filter((t) => t.category === filters.category);
    }

    // Date range filter
    if (filters.dateRange !== 'all') {
      const now = new Date();
      let cutoffDate: Date;
      switch (filters.dateRange) {
        case '7d':
          cutoffDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          break;
        case '30d':
          cutoffDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
          break;
        case '90d':
          cutoffDate = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000);
          break;
        case 'year':
          cutoffDate = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
          break;
      }
      tasks = tasks.filter((t) => new Date(t.updated_at) >= cutoffDate);
    }

    return tasks;
  }, [completedTasksData?.tasks, filters]);

  // Pagination
  const totalPages = Math.ceil(filteredTasks.length / ITEMS_PER_PAGE);
  const paginatedTasks = useMemo(() => {
    const start = (currentPage - 1) * ITEMS_PER_PAGE;
    return filteredTasks.slice(start, start + ITEMS_PER_PAGE);
  }, [filteredTasks, currentPage]);

  // Reset to page 1 when filters change
  const handleFiltersChange = useCallback((newFilters: ArchiveFiltersState) => {
    setFilters(newFilters);
    setCurrentPage(1);
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters({ search: '', category: '', dateRange: 'all' });
    setCurrentPage(1);
  }, []);

  const handleSelectionChange = useCallback((newSelection: Set<string>) => {
    setSelectedIds(newSelection);
  }, []);

  const handleTaskClick = useCallback((task: Task) => {
    setSelectedTaskId(task.id);
  }, []);

  const handleBulkDelete = () => {
    setShowDeleteConfirm(true);
  };

  const confirmBulkDelete = async () => {
    const taskIds = Array.from(selectedIds);
    await bulkDelete.mutateAsync(taskIds);
    setSelectedIds(new Set());
    setShowDeleteConfirm(false);
  };

  const handleBulkRestore = async () => {
    const taskIds = Array.from(selectedIds);
    await bulkRestore.mutateAsync(taskIds);
    setSelectedIds(new Set());
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <Archive className="h-8 w-8" />
            Archive
          </h2>
          <p className="text-muted-foreground">
            Loading completed tasks...
          </p>
        </div>
        <div className="space-y-4">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-[400px] w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className={`transition-all duration-[180ms] ${selectedTaskId ? 'lg:pr-96' : ''}`}>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <Archive className="h-8 w-8" />
            Archive
          </h2>
          <p className="text-muted-foreground">
            {filteredTasks.length} completed task{filteredTasks.length !== 1 ? 's' : ''}
            {isFetching && ' • Updating...'}
          </p>
        </div>

        {/* Filters */}
        <ArchiveFilters
          filters={filters}
          onChange={handleFiltersChange}
          onClear={handleClearFilters}
          categories={categories}
        />

        {/* Bulk Actions Bar */}
        <BulkActionsBar
          selectedCount={selectedIds.size}
          onDelete={handleBulkDelete}
          onRestore={handleBulkRestore}
          onClearSelection={() => setSelectedIds(new Set())}
          isDeleting={bulkDelete.isPending}
          isRestoring={bulkRestore.isPending}
        />

        {/* Table or Empty State */}
        {filteredTasks.length === 0 ? (
          <Card>
            <CardContent className="pt-6 text-center py-12">
              {completedTasksData?.tasks.length === 0 ? (
                <div className="space-y-2">
                  <Archive className="h-12 w-12 mx-auto text-muted-foreground/50" />
                  <p className="text-muted-foreground">
                    No completed tasks yet.
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Complete a task to see it here!
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  <p className="text-muted-foreground">
                    No tasks match your current filters.
                  </p>
                  <Button variant="outline" onClick={handleClearFilters}>
                    Clear filters
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        ) : (
          <>
            <ArchiveTable
              tasks={paginatedTasks}
              selectedIds={selectedIds}
              onSelectionChange={handleSelectionChange}
              onTaskClick={handleTaskClick}
            />

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Showing {(currentPage - 1) * ITEMS_PER_PAGE + 1} to{' '}
                  {Math.min(currentPage * ITEMS_PER_PAGE, filteredTasks.length)} of{' '}
                  {filteredTasks.length} tasks
                </p>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                  >
                    <ChevronLeft className="h-4 w-4 mr-1" />
                    Previous
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {currentPage} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages}
                  >
                    Next
                    <ChevronRight className="h-4 w-4 ml-1" />
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Task Details Sidebar */}
      {selectedTaskId && (
        <TaskDetailsSidebar
          taskId={selectedTaskId}
          onClose={() => setSelectedTaskId(null)}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {selectedIds.size} task{selectedIds.size !== 1 ? 's' : ''}?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. The selected tasks will be permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmBulkDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {bulkDelete.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
