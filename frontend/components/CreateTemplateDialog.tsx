'use client';

import { useState, useRef } from 'react';
import { useCreateTemplate } from '@/hooks/useTemplates';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import { CreateTaskTemplateDTO } from '@/lib/api';
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
import { FileText, Info } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface CreateTemplateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // Optional: Pre-fill form with task data (for "Save as Template" feature)
  initialData?: Partial<CreateTaskTemplateDTO>;
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

export function CreateTemplateDialog({
  open,
  onOpenChange,
  initialData,
}: CreateTemplateDialogProps) {
  const createTemplate = useCreateTemplate();
  const prevOpenRef = useRef(false);

  useDialogKeyboardShortcuts(open);

  const getEmptyFormData = (): FormData => ({
    name: '',
    title: '',
    description: '',
    category: '',
    estimated_effort: 'medium',
    user_priority: 5,
    due_date_offset: '',
    context: '',
  });

  const [formData, setFormData] = useState<FormData>(getEmptyFormData);

  // Handle dialog open/close transitions and apply initialData
  const handleOpenChange = (newOpen: boolean) => {
    if (newOpen && !prevOpenRef.current) {
      // Dialog is opening - reset form and apply initial data
      if (initialData) {
        setFormData({
          name: initialData.name || '',
          title: initialData.title || '',
          description: initialData.description || '',
          category: initialData.category || '',
          estimated_effort: initialData.estimated_effort || 'medium',
          user_priority: initialData.user_priority || 5,
          due_date_offset: initialData.due_date_offset?.toString() || '',
          context: initialData.context || '',
        });
      } else {
        setFormData(getEmptyFormData());
      }
    }
    prevOpenRef.current = newOpen;
    onOpenChange(newOpen);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const templateData: CreateTaskTemplateDTO = {
        name: formData.name,
        title: formData.title,
        description: formData.description || undefined,
        category: formData.category || undefined,
        estimated_effort: formData.estimated_effort,
        user_priority: formData.user_priority,
        context: formData.context || undefined,
        due_date_offset: formData.due_date_offset
          ? parseInt(formData.due_date_offset, 10)
          : undefined,
      };

      await createTemplate.mutateAsync(templateData);

      // Reset form and close dialog
      setFormData(getEmptyFormData());
      onOpenChange(false);
    } catch {
      // Error handled in hook
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Create Template
            </DialogTitle>
            <DialogDescription>
              Create a reusable template for quickly adding similar tasks.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Template Name */}
            <div className="grid gap-2">
              <Label htmlFor="template-name">Template Name *</Label>
              <Input
                id="template-name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., Weekly Report, Bug Fix"
                required
                maxLength={100}
              />
              <p className="text-xs text-muted-foreground">
                A short name to identify this template
              </p>
            </div>

            {/* Task Title */}
            <div className="grid gap-2">
              <Label htmlFor="template-title">Task Title *</Label>
              <Input
                id="template-title"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="What needs to be done?"
                required
                maxLength={200}
              />
            </div>

            {/* Description */}
            <div className="grid gap-2">
              <Label htmlFor="template-description">Description</Label>
              <Textarea
                id="template-description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Add more details..."
                rows={3}
                maxLength={2000}
              />
            </div>

            {/* Category */}
            <div className="grid gap-2">
              <Label htmlFor="template-category">Category</Label>
              <CategorySelect
                id="template-category"
                value={formData.category}
                onChange={(value) => setFormData({ ...formData, category: value })}
              />
            </div>

            {/* Estimated Effort */}
            <div className="grid gap-2">
              <Label htmlFor="template-effort">Estimated Effort</Label>
              <Select
                value={formData.estimated_effort}
                onValueChange={(value: 'small' | 'medium' | 'large' | 'xlarge') =>
                  setFormData({ ...formData, estimated_effort: value })
                }
              >
                <SelectTrigger id="template-effort">
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
              <Label htmlFor="template-priority">Priority (1-10)</Label>
              <Select
                value={formData.user_priority.toString()}
                onValueChange={(value) =>
                  setFormData({ ...formData, user_priority: parseInt(value, 10) })
                }
              >
                <SelectTrigger id="template-priority">
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
                <Label htmlFor="template-due-offset">Due Date Offset (days)</Label>
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
                id="template-due-offset"
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
              <Label htmlFor="template-context">Context</Label>
              <Input
                id="template-context"
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
              disabled={createTemplate.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createTemplate.isPending || !formData.name || !formData.title}
            >
              {createTemplate.isPending ? 'Creating...' : 'Create Template'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
