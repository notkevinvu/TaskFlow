'use client';

import { useState } from 'react';
import { useUpdateTask } from '@/hooks/useTasks';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Task } from '@/lib/api';
import { CategorySelect } from '@/components/CategorySelect';

interface EditTaskDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task: Task;
}

// Note: Parent should pass key={task.id} to ensure form resets when task changes
export function EditTaskDialog({ open, onOpenChange, task }: EditTaskDialogProps) {
  const updateTask = useUpdateTask();

  // Track dialog state for keyboard shortcuts
  useDialogKeyboardShortcuts(open);

  // Initialize form data from task prop
  // Parent uses key={task.id} to remount this component when task changes
  const [formData, setFormData] = useState({
    title: task.title,
    description: task.description || '',
    category: task.category || '',
    estimated_effort: (task.estimated_effort || 'medium') as 'small' | 'medium' | 'large' | 'xlarge',
    user_priority: task.user_priority,
    due_date: task.due_date ? new Date(task.due_date).toISOString().split('T')[0] : '',
    context: task.context || '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Convert date string to RFC3339 format for backend
      // Use noon local time to avoid timezone day-shift issues
      const dueDate = formData.due_date
        ? new Date(`${formData.due_date}T12:00:00`).toISOString()
        : undefined;

      await updateTask.mutateAsync({
        id: task.id,
        data: {
          title: formData.title,
          description: formData.description || undefined,
          category: formData.category || undefined,
          estimated_effort: formData.estimated_effort,
          user_priority: formData.user_priority,
          due_date: dueDate,
          context: formData.context || undefined,
        },
      });

      onOpenChange(false);
    } catch (error) {
      console.error('Failed to update task:', error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Task</DialogTitle>
            <DialogDescription>
              Update task details. Priority will be recalculated automatically.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Title */}
            <div className="grid gap-2">
              <Label htmlFor="title">Title *</Label>
              <Input
                id="title"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="What needs to be done?"
                required
              />
            </div>

            {/* Description */}
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Add more details..."
                rows={3}
              />
            </div>

            {/* Category */}
            <div className="grid gap-2">
              <Label htmlFor="category">Category</Label>
              <CategorySelect
                id="category"
                value={formData.category}
                onChange={(value) => setFormData({ ...formData, category: value })}
              />
            </div>

            {/* Estimated Effort */}
            <div className="grid gap-2">
              <Label htmlFor="effort">Estimated Effort</Label>
              <Select
                value={formData.estimated_effort}
                onValueChange={(value: 'small' | 'medium' | 'large' | 'xlarge') => setFormData({ ...formData, estimated_effort: value })}
              >
                <SelectTrigger id="effort">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="small">Small (&lt; 1 hour)</SelectItem>
                  <SelectItem value="medium">Medium (1-4 hours)</SelectItem>
                  <SelectItem value="large">Large (4-8 hours)</SelectItem>
                  <SelectItem value="xlarge">X-Large (&gt; 8 hours)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* User Priority */}
            <div className="grid gap-2">
              <Label htmlFor="priority">Your Priority (1-10)</Label>
              <Select
                value={formData.user_priority.toString()}
                onValueChange={(value) => setFormData({ ...formData, user_priority: parseInt(value) })}
              >
                <SelectTrigger id="priority">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">1 - Lowest</SelectItem>
                  <SelectItem value="2">2</SelectItem>
                  <SelectItem value="3">3</SelectItem>
                  <SelectItem value="4">4</SelectItem>
                  <SelectItem value="5">5 - Medium</SelectItem>
                  <SelectItem value="6">6</SelectItem>
                  <SelectItem value="7">7</SelectItem>
                  <SelectItem value="8">8</SelectItem>
                  <SelectItem value="9">9</SelectItem>
                  <SelectItem value="10">10 - Highest</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Due Date */}
            <div className="grid gap-2">
              <Label htmlFor="due_date">Due Date</Label>
              <Input
                id="due_date"
                type="date"
                value={formData.due_date}
                onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
              />
            </div>

            {/* Context */}
            <div className="grid gap-2">
              <Label htmlFor="context">Context</Label>
              <Input
                id="context"
                value={formData.context}
                onChange={(e) => setFormData({ ...formData, context: e.target.value })}
                placeholder="e.g., From Alice - needs review"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={updateTask.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={updateTask.isPending} disabled={!formData.title}>
              {updateTask.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
