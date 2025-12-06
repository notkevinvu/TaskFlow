'use client';

import { useState } from 'react';
import { useTemplates, templateToFormValues } from '@/hooks/useTemplates';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import { TaskTemplate, CreateTaskDTO } from '@/lib/api';
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
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  FileText,
  Clock,
  Tag,
  Gauge,
  Search,
  CalendarPlus,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface TemplatePickerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelectTemplate: (formValues: Partial<CreateTaskDTO>, template: TaskTemplate) => void;
}

const effortLabels: Record<string, string> = {
  small: 'Small',
  medium: 'Medium',
  large: 'Large',
  xlarge: 'X-Large',
};

export function TemplatePickerDialog({
  open,
  onOpenChange,
  onSelectTemplate,
}: TemplatePickerDialogProps) {
  const { data: templatesData, isLoading } = useTemplates();
  const [searchQuery, setSearchQuery] = useState('');

  useDialogKeyboardShortcuts(open);

  // Filter templates by search query
  const filteredTemplates = templatesData?.templates?.filter((template) => {
    const query = searchQuery.toLowerCase();
    return (
      template.name.toLowerCase().includes(query) ||
      template.title.toLowerCase().includes(query) ||
      template.category?.toLowerCase().includes(query) ||
      template.description?.toLowerCase().includes(query)
    );
  }) || [];

  const handleSelectTemplate = (template: TaskTemplate) => {
    const formValues = templateToFormValues(template);
    onSelectTemplate(formValues, template);
    onOpenChange(false);
    setSearchQuery('');
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setSearchQuery('');
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Create from Template
          </DialogTitle>
          <DialogDescription>
            Select a template to quickly create a new task with pre-filled values.
          </DialogDescription>
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        {/* Template List */}
        <ScrollArea className="h-[350px] pr-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              Loading templates...
            </div>
          ) : filteredTemplates.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {templatesData?.templates?.length === 0 ? (
                <>
                  <FileText className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p>No templates yet</p>
                  <p className="text-sm mt-1">
                    Create a template from the sidebar to get started
                  </p>
                </>
              ) : (
                <>
                  <Search className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p>No templates match your search</p>
                </>
              )}
            </div>
          ) : (
            <div className="space-y-2">
              {filteredTemplates.map((template) => (
                <button
                  key={template.id}
                  onClick={() => handleSelectTemplate(template)}
                  className={cn(
                    'w-full text-left p-4 rounded-lg border bg-card',
                    'hover:bg-accent/50 hover:border-primary/30',
                    'transition-all cursor-pointer',
                    'focus:outline-none focus:ring-2 focus:ring-primary/50'
                  )}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold truncate">{template.name}</h3>
                      <p className="text-sm text-muted-foreground truncate mt-0.5">
                        {template.title}
                      </p>
                    </div>
                    <CalendarPlus className="h-4 w-4 text-muted-foreground flex-shrink-0 mt-1" />
                  </div>

                  {/* Template metadata */}
                  <div className="flex flex-wrap items-center gap-2 mt-3">
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
                        {template.due_date_offset === 0
                          ? 'Due today'
                          : template.due_date_offset === 1
                          ? 'Due tomorrow'
                          : `Due in ${template.due_date_offset} days`}
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
                        Priority {template.user_priority}
                      </Badge>
                    )}
                  </div>

                  {/* Description preview */}
                  {template.description && (
                    <p className="text-xs text-muted-foreground mt-2 line-clamp-2">
                      {template.description}
                    </p>
                  )}
                </button>
              ))}
            </div>
          )}
        </ScrollArea>

        {/* Footer */}
        <div className="flex justify-end pt-2">
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
