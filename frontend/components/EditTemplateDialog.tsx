'use client';

import { useState, useRef } from 'react';
import { useUpdateTemplate } from '@/hooks/useTemplates';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import { TaskTemplate, UpdateTaskTemplateDTO } from '@/lib/api';
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
import { Pencil, Info } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface EditTemplateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  template: TaskTemplate | null;
}

interface FormData {
  name: string;
  title: string;
  description: string;
  category: string;
  estimated_effort: 'small' | 'medium' | 'large' | 'xlarge';
  user_priority: number;
  due_date_offset: string;
  context: string;
}

export function EditTemplateDialog({
  open,
  onOpenChange,
  template,
}: EditTemplateDialogProps) {
  const updateTemplate = useUpdateTemplate();
  const prevOpenRef = useRef(false);

  useDialogKeyboardShortcuts(open);

  const [formData, setFormData] = useState<FormData>({
    name: '',
    title: '',
    description: '',
    category: '',
    estimated_effort: 'medium',
    user_priority: 5,
    due_date_offset: '',
    context: '',
  });

  // Handle dialog open/close transitions and load template data
  const handleOpenChange = (newOpen: boolean) => {
    if (newOpen && !prevOpenRef.current && template) {
      // Dialog is opening - load template data
      setFormData({
        name: template.name || '',
        title: template.title || '',
        description: template.description || '',
        category: template.category || '',
        estimated_effort: template.estimated_effort || 'medium',
        user_priority: template.user_priority || 5,
        due_date_offset:
          template.due_date_offset !== undefined && template.due_date_offset !== null
            ? template.due_date_offset.toString()
            : '',
        context: template.context || '',
      });
    }
    prevOpenRef.current = newOpen;
    onOpenChange(newOpen);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!template) return;

    try {
      const updates: UpdateTaskTemplateDTO = {};

      // Only include fields that changed
      if (formData.name !== template.name) {
        updates.name = formData.name;
      }
      if (formData.title !== template.title) {
        updates.title = formData.title;
      }
      if (formData.description !== (template.description || '')) {
        updates.description = formData.description || undefined;
      }
      if (formData.category !== (template.category || '')) {
        updates.category = formData.category || undefined;
      }
      if (formData.estimated_effort !== (template.estimated_effort || 'medium')) {
        updates.estimated_effort = formData.estimated_effort;
      }
      if (formData.user_priority !== template.user_priority) {
        updates.user_priority = formData.user_priority;
      }
      if (formData.context !== (template.context || '')) {
        updates.context = formData.context || undefined;
      }

      // Handle due_date_offset
      const newOffset = formData.due_date_offset
        ? parseInt(formData.due_date_offset, 10)
        : undefined;
      const oldOffset = template.due_date_offset;
      if (newOffset !== oldOffset) {
        updates.due_date_offset = newOffset;
      }

      await updateTemplate.mutateAsync({ id: template.id, data: updates });
      onOpenChange(false);
    } catch {
      // Error handled in hook
    }
  };

  if (!template) return null;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Pencil className="h-5 w-5" />
              Edit Template
            </DialogTitle>
            <DialogDescription>
              Make changes to your template. Tasks created from this template are not affected.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Template Name */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-name">Template Name *</Label>
              <Input
                id="edit-template-name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., Weekly Report, Bug Fix"
                required
                maxLength={100}
              />
            </div>

            {/* Task Title */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-title">Task Title *</Label>
              <Input
                id="edit-template-title"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="What needs to be done?"
                required
                maxLength={200}
              />
            </div>

            {/* Description */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-description">Description</Label>
              <Textarea
                id="edit-template-description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Add more details..."
                rows={3}
                maxLength={2000}
              />
            </div>

            {/* Category */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-category">Category</Label>
              <CategorySelect
                id="edit-template-category"
                value={formData.category}
                onChange={(value) => setFormData({ ...formData, category: value })}
              />
            </div>

            {/* Estimated Effort */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-effort">Estimated Effort</Label>
              <Select
                value={formData.estimated_effort}
                onValueChange={(value: 'small' | 'medium' | 'large' | 'xlarge') =>
                  setFormData({ ...formData, estimated_effort: value })
                }
              >
                <SelectTrigger id="edit-template-effort">
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
              <Label htmlFor="edit-template-priority">Priority (1-10)</Label>
              <Select
                value={formData.user_priority.toString()}
                onValueChange={(value) =>
                  setFormData({ ...formData, user_priority: parseInt(value, 10) })
                }
              >
                <SelectTrigger id="edit-template-priority">
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

            {/* Due Date Offset */}
            <div className="grid gap-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="edit-template-due-offset">Due Date Offset (days)</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent side="right" className="max-w-[250px]">
                      <p>
                        Set a relative due date. For example, &quot;7&quot; means the task will be
                        due 7 days from when it&apos;s created from this template.
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <Input
                id="edit-template-due-offset"
                type="number"
                min="0"
                max="365"
                value={formData.due_date_offset}
                onChange={(e) =>
                  setFormData({ ...formData, due_date_offset: e.target.value })
                }
                placeholder="e.g., 7 for 1 week"
              />
              <p className="text-xs text-muted-foreground">
                Leave empty for no due date
              </p>
            </div>

            {/* Context */}
            <div className="grid gap-2">
              <Label htmlFor="edit-template-context">Context</Label>
              <Input
                id="edit-template-context"
                value={formData.context}
                onChange={(e) => setFormData({ ...formData, context: e.target.value })}
                placeholder="e.g., From Alice - needs review"
                maxLength={500}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={updateTemplate.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={updateTemplate.isPending || !formData.name || !formData.title}
            >
              {updateTemplate.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
