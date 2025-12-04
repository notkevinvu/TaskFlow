'use client';

import { useState, useRef } from 'react';
import { useCreateTask } from '@/hooks/useTasks';
import { RecurrenceRule } from '@/lib/api';
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
import { CategorySelect } from '@/components/CategorySelect';
import { RecurrenceSelector } from '@/components/RecurrenceSelector';

interface CreateTaskDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialDueDate?: string; // YYYY-MM-DD format
}

export function CreateTaskDialog({ open, onOpenChange, initialDueDate }: CreateTaskDialogProps) {
  const createTask = useCreateTask();
  const prevOpenRef = useRef(false);

  const getEmptyFormData = () => ({
    title: '',
    description: '',
    category: '',
    estimated_effort: 'medium' as 'small' | 'medium' | 'large' | 'xlarge',
    user_priority: 5,
    due_date: '',
    context: '',
    recurrence: null as RecurrenceRule | null,
  });

  const [formData, setFormData] = useState(getEmptyFormData);

  // Handle dialog open/close transitions
  const handleOpenChange = (newOpen: boolean) => {
    if (newOpen && !prevOpenRef.current) {
      // Dialog is opening - reset form and apply initialDueDate if provided
      setFormData({
        ...getEmptyFormData(),
        due_date: initialDueDate || '',
      });
    }
    prevOpenRef.current = newOpen;
    onOpenChange(newOpen);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Convert date string to RFC3339 format for backend
      const dueDate = formData.due_date
        ? new Date(formData.due_date).toISOString()
        : undefined;

      // Convert recurrence end date if present
      const recurrence = formData.recurrence
        ? {
            ...formData.recurrence,
            end_date: formData.recurrence.end_date
              ? new Date(formData.recurrence.end_date).toISOString()
              : undefined,
          }
        : undefined;

      await createTask.mutateAsync({
        title: formData.title,
        description: formData.description || undefined,
        category: formData.category || undefined,
        estimated_effort: formData.estimated_effort,
        user_priority: formData.user_priority,
        due_date: dueDate,
        context: formData.context || undefined,
        recurrence,
      });

      // Reset form and close dialog
      setFormData(getEmptyFormData());
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to create task:', error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Task</DialogTitle>
            <DialogDescription>
              Add a new task to your list. The priority will be calculated automatically.
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

            {/* Recurrence */}
            <RecurrenceSelector
              value={formData.recurrence}
              onChange={(recurrence) => setFormData({ ...formData, recurrence })}
              showEndDate
            />

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
              disabled={createTask.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createTask.isPending || !formData.title}>
              {createTask.isPending ? 'Creating...' : 'Create Task'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
