'use client';

import { useState } from 'react';
import { useTemplates, useDeleteTemplate } from '@/hooks/useTemplates';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import { TaskTemplate } from '@/lib/api';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
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
import {
  FileText,
  Pencil,
  Trash2,
  Tag,
  Clock,
  Gauge,
  Plus,
  Settings,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface ManageTemplatesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEditTemplate: (template: TaskTemplate) => void;
  onCreateTemplate: () => void;
}

const effortLabels: Record<string, string> = {
  small: 'Small',
  medium: 'Medium',
  large: 'Large',
  xlarge: 'X-Large',
};

export function ManageTemplatesDialog({
  open,
  onOpenChange,
  onEditTemplate,
  onCreateTemplate,
}: ManageTemplatesDialogProps) {
  const { data: templatesData, isLoading } = useTemplates();
  const deleteTemplate = useDeleteTemplate();
  const [templateToDelete, setTemplateToDelete] = useState<TaskTemplate | null>(null);

  useDialogKeyboardShortcuts(open);

  const handleDelete = async () => {
    if (!templateToDelete) return;

    try {
      await deleteTemplate.mutateAsync(templateToDelete.id);
      setTemplateToDelete(null);
    } catch {
      // Error is handled in hook
    }
  };

  const templates = templatesData?.templates || [];

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              Manage Templates
            </DialogTitle>
            <DialogDescription>
              View, edit, or delete your task templates.
            </DialogDescription>
          </DialogHeader>

          {/* Create New Button */}
          <div className="flex justify-end">
            <Button
              onClick={() => {
                onCreateTemplate();
              }}
              size="sm"
              className="gap-1"
            >
              <Plus className="h-4 w-4" />
              New Template
            </Button>
          </div>

          {/* Template List */}
          <ScrollArea className="h-[400px] pr-4">
            {isLoading ? (
              <div className="flex items-center justify-center py-8 text-muted-foreground">
                Loading templates...
              </div>
            ) : templates.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <FileText className="h-12 w-12 mx-auto mb-3 opacity-50" />
                <p>No templates yet</p>
                <p className="text-sm mt-1">
                  Create your first template to quickly add common tasks
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {templates.map((template) => (
                  <div
                    key={template.id}
                    className="p-4 rounded-lg border bg-card hover:bg-accent/30 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                          <h3 className="font-semibold truncate">{template.name}</h3>
                        </div>
                        <p className="text-sm text-muted-foreground truncate mt-1 ml-6">
                          {template.title}
                        </p>

                        {/* Metadata badges */}
                        <div className="flex flex-wrap items-center gap-2 mt-2 ml-6">
                          {template.category && (
                            <Badge variant="secondary" className="text-xs gap-1">
                              <Tag className="h-3 w-3" />
                              {template.category}
                            </Badge>
                          )}
                          {template.estimated_effort && (
                            <Badge variant="outline" className="text-xs gap-1">
                              <Gauge className="h-3 w-3" />
                              {effortLabels[template.estimated_effort]}
                            </Badge>
                          )}
                          {template.due_date_offset !== undefined && template.due_date_offset !== null && (
                            <Badge variant="outline" className="text-xs gap-1">
                              <Clock className="h-3 w-3" />
                              +{template.due_date_offset}d
                            </Badge>
                          )}
                          {template.user_priority !== 5 && (
                            <Badge
                              variant="outline"
                              className={cn(
                                'text-xs',
                                template.user_priority >= 8 && 'border-red-500 text-red-600',
                                template.user_priority <= 3 && 'border-blue-500 text-blue-600'
                              )}
                            >
                              P{template.user_priority}
                            </Badge>
                          )}
                        </div>
                      </div>

                      {/* Action buttons */}
                      <div className="flex items-center gap-1 flex-shrink-0">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => onEditTemplate(template)}
                          className="transition-all hover:scale-105 cursor-pointer"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setTemplateToDelete(template)}
                          className="transition-all hover:scale-105 cursor-pointer text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={!!templateToDelete}
        onOpenChange={(open) => !open && setTemplateToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Template</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the template &quot;{templateToDelete?.name}&quot;?
              <br />
              <br />
              This action cannot be undone. Tasks created from this template will not be affected.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteTemplate.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleteTemplate.isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {deleteTemplate.isPending ? 'Deleting...' : 'Delete Template'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
