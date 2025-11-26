'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
import { useTasks, useBumpTask, useCompleteTask, useDeleteTask, useAtRiskTasks, type TaskFilters as TaskFiltersType } from '@/hooks/useTasks';
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

export default function DashboardPage() {
  const searchParams = useSearchParams();
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [manageCategoriesOpen, setManageCategoriesOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<TaskFilterState>({});

  // Build filter params for API
  const filterParams: TaskFiltersType = {
    search: searchQuery || undefined,
    status: filters.status,
    category: filters.category,
    min_priority: filters.minPriority,
    max_priority: filters.maxPriority,
  };

  const { data: tasksData, isLoading, isFetching } = useTasks(filterParams);
  const { data: atRiskData } = useAtRiskTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();
  const deleteTask = useDeleteTask();

  // Read taskId from URL query params
  useEffect(() => {
    const taskId = searchParams.get('taskId');
    if (taskId) {
      setSelectedTaskId(taskId);
    }
  }, [searchParams]);

  const tasks = tasksData?.tasks || [];
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
          <h2 className="text-3xl font-bold">Today's Priorities</h2>
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
    <div className={`transition-all duration-[180ms] ${selectedTaskId ? 'lg:pr-96' : ''}`}>
      {/* Main Content */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold">Today's Priorities</h2>
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

      {/* Task List */}
      <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Your Tasks</h3>
            <div className="flex items-center gap-2">
              {isFetching && (
                <div className="flex items-center gap-1 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Updating...</span>
                </div>
              )}
              {tasks.length > 0 && (
                <p className="text-sm text-muted-foreground">
                  Showing {tasks.length} task{tasks.length !== 1 ? 's' : ''}
                </p>
              )}
            </div>
          </div>
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
                  <p className="text-muted-foreground">
                    No tasks yet. Create your first task to get started!
                  </p>
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
                        "{task.context}"
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
                          if (selectedTaskId === task.id) {
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
        </div>
    </div>

    {/* Task Details Sidebar */}
    {selectedTaskId && (
      <TaskDetailsSidebar
        taskId={selectedTaskId}
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
