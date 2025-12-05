'use client';

import { useState, useMemo } from 'react';
import { useTasks } from '@/hooks/useTasks';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Pencil, Trash2, X, Check } from 'lucide-react';
import { toast } from 'sonner';
import { categoryAPI, getApiErrorMessage } from '@/lib/api';
import { useQueryClient } from '@tanstack/react-query';

interface ManageCategoriesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ManageCategoriesDialog({ open, onOpenChange }: ManageCategoriesDialogProps) {
  const { data: tasksData } = useTasks();
  const queryClient = useQueryClient();
  const [editingCategory, setEditingCategory] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [deleteCategory, setDeleteCategory] = useState<string | null>(null);
  const [isRenaming, setIsRenaming] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  // Track dialog state for keyboard shortcuts
  useDialogKeyboardShortcuts(open);

  // Extract categories with task counts
  const categories = useMemo(() => {
    if (!tasksData?.tasks) return [];

    const categoryMap = new Map<string, number>();
    tasksData.tasks.forEach(task => {
      if (task.category && task.category.trim() !== '') {
        const count = categoryMap.get(task.category) || 0;
        categoryMap.set(task.category, count + 1);
      }
    });

    return Array.from(categoryMap.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [tasksData]);

  const handleStartEdit = (category: string) => {
    setEditingCategory(category);
    setEditValue(category);
  };

  const handleCancelEdit = () => {
    setEditingCategory(null);
    setEditValue('');
  };

  const handleSaveEdit = async (oldName: string) => {
    if (!editValue.trim()) {
      toast.error('Category name cannot be empty');
      return;
    }

    if (editValue === oldName) {
      handleCancelEdit();
      return;
    }

    // Check if new name already exists
    const exists = categories.some(c => c.name.toLowerCase() === editValue.trim().toLowerCase() && c.name !== oldName);
    if (exists) {
      toast.error('A category with this name already exists');
      return;
    }

    setIsRenaming(true);
    try {
      await categoryAPI.rename(oldName, editValue.trim());
      await queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success(`Renamed "${oldName}" to "${editValue.trim()}"`);
      handleCancelEdit();
    } catch (err: unknown) {
      toast.error(getApiErrorMessage(err, 'Failed to rename category', 'Category Rename'));
    } finally {
      setIsRenaming(false);
    }
  };

  const handleDelete = async (categoryName: string) => {
    setIsDeleting(true);
    try {
      await categoryAPI.delete(categoryName);
      await queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success(`Deleted category "${categoryName}"`);
      setDeleteCategory(null);
    } catch (err: unknown) {
      toast.error(getApiErrorMessage(err, 'Failed to delete category', 'Category Delete'));
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Manage Categories</DialogTitle>
            <DialogDescription>
              Rename or delete categories. Changes will affect all tasks using these categories.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2 max-h-[400px] overflow-y-auto py-2">
            {categories.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No categories yet</p>
                <p className="text-sm mt-1">Categories will appear here once you add them to tasks</p>
              </div>
            ) : (
              categories.map((category) => (
                <div
                  key={category.name}
                  className="flex items-center gap-2 p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                >
                  {editingCategory === category.name ? (
                    <>
                      <Input
                        value={editValue}
                        onChange={(e) => setEditValue(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleSaveEdit(category.name);
                          if (e.key === 'Escape') handleCancelEdit();
                        }}
                        className="flex-1"
                        maxLength={50}
                        autoFocus
                        disabled={isRenaming}
                      />
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleSaveEdit(category.name)}
                        disabled={isRenaming}
                      >
                        <Check className="h-4 w-4 text-green-600" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={handleCancelEdit}
                        disabled={isRenaming}
                      >
                        <X className="h-4 w-4 text-red-600" />
                      </Button>
                    </>
                  ) : (
                    <>
                      <div className="flex-1 flex items-center gap-2">
                        <span className="font-medium">{category.name}</span>
                        <Badge variant="secondary" className="text-xs">
                          {category.count} {category.count === 1 ? 'task' : 'tasks'}
                        </Badge>
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleStartEdit(category.name)}
                        className="transition-all hover:scale-105 cursor-pointer"
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setDeleteCategory(category.name)}
                        className="transition-all hover:scale-105 cursor-pointer text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                </div>
              ))
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={!!deleteCategory} onOpenChange={(open) => !open && setDeleteCategory(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Category</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the category &quot;{deleteCategory}&quot;?
              <br />
              <br />
              This will <strong>remove the category from {categories.find(c => c.name === deleteCategory)?.count || 0} task(s)</strong>, but the tasks themselves will not be deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteCategory && handleDelete(deleteCategory)}
              disabled={isDeleting}
              className="bg-red-600 hover:bg-red-700"
            >
              {isDeleting ? 'Deleting...' : 'Delete Category'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
