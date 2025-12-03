'use client';

import { useState, useMemo, useCallback, useEffect } from 'react';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import { useTasks, useBumpTask, useCompleteTask, useDeleteTask, useAtRiskTasks, useCompletedTasks, type TaskFilters as TaskFiltersType } from '@/hooks/useTasks';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { TaskDetailsSidebar } from "@/components/TaskDetailsSidebar";
import { CreateTaskDialog } from "@/components/CreateTaskDialog";
import { EditTaskDialog } from "@/components/EditTaskDialog";
import { ManageCategoriesDialog } from "@/components/ManageCategoriesDialog";
import { TaskSearch } from "@/components/TaskSearch";
import { TaskFilters, type TaskFilterState } from "@/components/TaskFilters";
import { Plus, Trash2, Pencil, FolderKanban, Loader2 } from "lucide-react";
import { Task } from "@/lib/api";

// Valid status values for validation
const VALID_STATUSES = ['todo', 'in_progress', 'done'];

// Validate date string format (YYYY-MM-DD) and that it parses to a valid date
function isValidDateString(dateStr: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) return false;
  const [year, month, day] = dateStr.split('-').map(Number);
  const date = new Date(year, month - 1, day);
  return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day;
}

// Helper to parse filters from URL search params with validation
function parseFiltersFromURL(searchParams: URLSearchParams): TaskFilterState {
  const filters: TaskFilterState = {};
  const status = searchParams.get('status');
  const category = searchParams.get('category');
  const minPriority = searchParams.get('minPriority');
  const maxPriority = searchParams.get('maxPriority');
  const dueDateStart = searchParams.get('dueDateStart');
  const dueDateEnd = searchParams.get('dueDateEnd');

  // Validate status against known values
  if (status && VALID_STATUSES.includes(status)) {
    filters.status = status;
  }

  if (category) filters.category = category;

  // Validate priority values are numbers in valid range (0-100)
  if (minPriority) {
    const parsed = parseInt(minPriority, 10);
    if (!isNaN(parsed) && parsed >= 0 && parsed <= 100) {
      filters.minPriority = parsed;
    }
  }
  if (maxPriority) {
    const parsed = parseInt(maxPriority, 10);
    if (!isNaN(parsed) && parsed >= 0 && parsed <= 100) {
      filters.maxPriority = parsed;
    }
  }

  // Validate date strings
  if (dueDateStart && isValidDateString(dueDateStart)) {
    filters.dueDateStart = dueDateStart;
  }
  if (dueDateEnd && isValidDateString(dueDateEnd)) {
    filters.dueDateEnd = dueDateEnd;
  }

  return filters;
}

// Helper to serialize filters to URL search params
function serializeFiltersToURL(
  filters: TaskFilterState,
  search: string,
  taskId: string | null
): string {
  const params = new URLSearchParams();

  if (search) params.set('search', search);
  if (taskId) params.set('taskId', taskId);
  if (filters.status) params.set('status', filters.status);
  if (filters.category) params.set('category', filters.category);
  if (filters.minPriority !== undefined) params.set('minPriority', filters.minPriority.toString());
  if (filters.maxPriority !== undefined) params.set('maxPriority', filters.maxPriority.toString());
  if (filters.dueDateStart) params.set('dueDateStart', filters.dueDateStart);
  if (filters.dueDateEnd) params.set('dueDateEnd', filters.dueDateEnd);

  const queryString = params.toString();
  return queryString ? `?${queryString}` : '';
}

export default function DashboardPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  // Initialize state from URL on first render
  const initialTaskId = searchParams.get('taskId');
  const initialSearch = searchParams.get('search') || '';
  const initialFilters = parseFiltersFromURL(searchParams);

  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(initialTaskId);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [manageCategoriesOpen, setManageCategoriesOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [searchQuery, setSearchQuery] = useState(initialSearch);
  const [filters, setFilters] = useState<TaskFilterState>(initialFilters);
  const [activeTab, setActiveTab] = useState<'active' | 'completed'>('active');

  // Update URL when filters change (debounced via useEffect)
  useEffect(() => {
    const newUrl = pathname + serializeFiltersToURL(filters, searchQuery, selectedTaskId);
    const currentUrl = pathname + (searchParams.toString() ? `?${searchParams.toString()}` : '');

    // Only update if URL actually changed
    if (newUrl !== currentUrl) {
      router.replace(newUrl, { scroll: false });
    }
  }, [filters, searchQuery, selectedTaskId, pathname, router, searchParams]);

  // Build filter params for API
  const filterParams: TaskFiltersType = {
    search: searchQuery || undefined,
    status: filters.status,
    category: filters.category,
    min_priority: filters.minPriority,
    max_priority: filters.maxPriority,
    due_date_start: filters.dueDateStart,
    due_date_end: filters.dueDateEnd,
  };

  const { data: tasksData, isLoading, isFetching } = useTasks(filterParams);
  const { data: completedTasksData, isLoading: isLoadingCompleted, isFetching: isFetchingCompleted } = useCompletedTasks();
  const { data: atRiskData } = useAtRiskTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();
  const deleteTask = useDeleteTask();

  // Sync selectedTaskId when URL changes externally (e.g., browser back/forward)
  // Using useMemo to derive the effective taskId from URL without setState in useEffect
  const urlTaskId = searchParams.get('taskId');
  const effectiveTaskId = urlTaskId !== null ? urlTaskId : selectedTaskId;

  // Memoize tasks to prevent useMemo dependency issues
  const allTasks = useMemo(() => tasksData?.tasks || [], [tasksData?.tasks]);

  // Filter to show only active tasks (not completed) by default
  // Only show completed tasks if user explicitly filters by status=done
  const tasks = useMemo(() => {
    // If user explicitly selected a status filter, respect it
    if (filters.status) {
      return allTasks;
    }
    // Otherwise, filter out completed tasks
    return allTasks.filter((t) => t.status !== 'done');
  }, [allTasks, filters.status]);

  // Completed tasks for the Completed tab
  const completedTasks = useMemo(() => completedTasksData?.tasks || [], [completedTasksData?.tasks]);

  const atRiskCount = atRiskData?.count || 0;
  const quickWins = tasks.filter(
    (t) => t.estimated_effort === 'small' && t.bump_count === 0
  ).length;
  const totalTasks = tasks.length;

  // Extract unique categories from tasks (memoized to avoid recalculation)
  const availableCategories = useMemo(() => {
    return Array.from(
      new Set(
        tasks
          .map((task) => task.category)
          .filter((cat): cat is string => !!cat)
      )
    ).sort();
  }, [tasks]);

  // Memoize handlers to prevent unnecessary re-renders
  const handleSearchChange = useCallback((value: string) => {
    setSearchQuery(value);
  }, []);

  const handleFiltersChange = useCallback((newFilters: TaskFilterState) => {
    setFilters(newFilters);
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters({});
    setSearchQuery('');
  }, []);

  // Show full skeleton only on initial load (no data yet)
  if (isLoading && !tasksData) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold">Today&apos;s Priorities</h2>
          <p className="text-muted-foreground">
            Loading your intelligently prioritized tasks...
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={`transition-all duration-[180ms] ${effectiveTaskId ? 'lg:pr-96' : ''}`}>
      {/* Main Content */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold">Today&apos;s Priorities</h2>
            <p className="text-muted-foreground">
              Tasks sorted by intelligent priority algorithm
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => setManageCategoriesOpen(true)}
              className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
            >
              <FolderKanban className="mr-2 h-4 w-4" />
              Manage Categories
            </Button>
            <Button
              onClick={() => setCreateDialogOpen(true)}
              className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
            >
              <Plus className="mr-2 h-4 w-4" />
              Quick Add
            </Button>
          </div>
        </div>

        {/* Search and Filters */}
        <div className="space-y-4">
          <TaskSearch
            value={searchQuery}
            onChange={handleSearchChange}
          />
          <TaskFilters
            filters={filters}
            onChange={handleFiltersChange}
            onClear={handleClearFilters}
            availableCategories={availableCategories}
          />
        </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Tasks
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalTasks}</div>
            <p className="text-xs text-muted-foreground">
              Active tasks to complete
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              At Risk
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{atRiskCount}</div>
            <p className="text-xs text-muted-foreground">
              Tasks bumped 3+ times
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Quick Wins
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{quickWins}</div>
            <p className="text-xs text-muted-foreground">
              Small tasks ready to complete
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Task List with Tabs */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'active' | 'completed')} className="space-y-4">
        <div className="flex items-center justify-between">
          <TabsList>
            <TabsTrigger value="active" className="gap-2">
              Active
              {tasks.length > 0 && (
                <Badge variant="secondary" className="ml-1">{tasks.length}</Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="completed" className="gap-2">
              Completed
              {completedTasks.length > 0 && (
                <Badge variant="secondary" className="ml-1">{completedTasks.length}</Badge>
              )}
            </TabsTrigger>
          </TabsList>
          <div className="flex items-center gap-2">
            {(activeTab === 'active' ? isFetching : isFetchingCompleted) && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Updating...</span>
              </div>
            )}
          </div>
        </div>

        {/* Active Tasks Tab */}
        <TabsContent value="active" className="space-y-4">
          <div className={`space-y-4 transition-opacity duration-200 ${isFetching ? 'opacity-50' : 'opacity-100'}`}>
          {tasks.length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                {searchQuery || Object.keys(filters).length > 0 ? (
                  <div className="space-y-2">
                    <p className="text-muted-foreground">
                      No tasks match your search or filters.
                    </p>
                    <Button
                      variant="outline"
                      onClick={handleClearFilters}
                    >
                      Clear filters
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <p className="text-muted-foreground">
                      {allTasks.length > 0 && allTasks.every(t => t.status === 'done')
                        ? "All tasks completed! You're all caught up. 🎉"
                        : "No active tasks yet. Create your first task to get started!"}
                    </p>
                    {completedTasks.length > 0 && (
                      <Button
                        variant="link"
                        onClick={() => setActiveTab('completed')}
                        className="text-primary"
                      >
                        View {completedTasks.length} completed task{completedTasks.length !== 1 ? 's' : ''}
                      </Button>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ) : (
            tasks.map((task) => (
            <Card
              key={task.id}
              className="hover:shadow-md transition-shadow cursor-pointer py-0"
              onClick={() => setSelectedTaskId(task.id)}
            >
              <CardContent className="pt-4 px-6 pb-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h4 className="font-semibold">{task.title}</h4>
                      <Badge
                        variant={
                          task.priority_score >= 90
                            ? "destructive"
                            : task.priority_score >= 75
                            ? "default"
                            : "secondary"
                        }
                      >
                        {Math.round(task.priority_score)}
                      </Badge>
                      {task.category && (
                        <Badge variant="outline" className="bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800">
                          {task.category}
                        </Badge>
                      )}
                      {task.bump_count > 0 && (
                        <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                          ⚠️ Bumped {task.bump_count}x
                        </Badge>
                      )}
                      {task.estimated_effort && (
                        <Badge variant="outline" className="capitalize">
                          {task.estimated_effort}
                        </Badge>
                      )}
                    </div>
                    {task.description && (
                      <p className="text-sm text-muted-foreground mb-2">
                        {task.description}
                      </p>
                    )}
                    {task.context && (
                      <p className="text-sm italic text-muted-foreground mb-2">
                        &quot;{task.context}&quot;
                      </p>
                    )}
                    <div className="flex gap-4 text-sm text-muted-foreground">
                      {task.due_date && (
                        <span>📅 Due: {new Date(task.due_date).toLocaleDateString()}</span>
                      )}
                      {task.related_people && task.related_people.length > 0 && (
                        <span>👥 {task.related_people.join(', ')}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2 ml-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditingTask(task);
                      }}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        bumpTask.mutate({ id: task.id });
                      }}
                      disabled={bumpTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      Bump
                    </Button>
                    <Button
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        completeTask.mutate(task.id);
                      }}
                      disabled={completeTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      Complete
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (window.confirm('Are you sure you want to delete this task?')) {
                          deleteTask.mutate(task.id);
                          if (effectiveTaskId === task.id) {
                            setSelectedTaskId(null);
                          }
                        }
                      }}
                      disabled={deleteTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-lg cursor-pointer"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
          )}
          </div>
        </TabsContent>

        {/* Completed Tasks Tab */}
        <TabsContent value="completed" className="space-y-4">
          <div className={`space-y-4 transition-opacity duration-200 ${isFetchingCompleted ? 'opacity-50' : 'opacity-100'}`}>
          {isLoadingCompleted ? (
            <div className="space-y-4">
              {[...Array(3)].map((_, i) => (
                <Skeleton key={i} className="h-32" />
              ))}
            </div>
          ) : completedTasks.length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                <p className="text-muted-foreground">
                  No completed tasks yet. Complete a task to see it here!
                </p>
              </CardContent>
            </Card>
          ) : (
            completedTasks.map((task) => (
              <Card
                key={task.id}
                className="hover:shadow-md transition-shadow cursor-pointer py-0 opacity-80"
                onClick={() => setSelectedTaskId(task.id)}
              >
                <CardContent className="pt-4 px-6 pb-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h4 className="font-semibold line-through text-muted-foreground">{task.title}</h4>
                        <Badge variant="secondary" className="bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300">
                          ✓ Completed
                        </Badge>
                        {task.category && (
                          <Badge variant="outline" className="bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800">
                            {task.category}
                          </Badge>
                        )}
                      </div>
                      {task.description && (
                        <p className="text-sm text-muted-foreground mb-2">
                          {task.description}
                        </p>
                      )}
                      <div className="flex gap-4 text-sm text-muted-foreground">
                        <span>Completed: {new Date(task.updated_at).toLocaleDateString()}</span>
                        {task.bump_count > 0 && (
                          <span>Bumped {task.bump_count}x before completion</span>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2 ml-4">
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (window.confirm('Are you sure you want to delete this completed task?')) {
                            deleteTask.mutate(task.id);
                            if (effectiveTaskId === task.id) {
                              setSelectedTaskId(null);
                            }
                          }
                        }}
                        disabled={deleteTask.isPending}
                        className="transition-all hover:scale-105 hover:shadow-lg cursor-pointer"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
          </div>
        </TabsContent>
      </Tabs>
    </div>

    {/* Task Details Sidebar */}
    {effectiveTaskId && (
      <TaskDetailsSidebar
        taskId={effectiveTaskId}
        onClose={() => setSelectedTaskId(null)}
      />
    )}

    {/* Create Task Dialog */}
    <CreateTaskDialog
      open={createDialogOpen}
      onOpenChange={setCreateDialogOpen}
    />

    {/* Edit Task Dialog */}
    {editingTask && (
      <EditTaskDialog
        key={editingTask.id}
        open={!!editingTask}
        onOpenChange={(open) => !open && setEditingTask(null)}
        task={editingTask}
      />
    )}

    {/* Manage Categories Dialog */}
    <ManageCategoriesDialog
      open={manageCategoriesOpen}
      onOpenChange={setManageCategoriesOpen}
    />
  </div>
);
}
