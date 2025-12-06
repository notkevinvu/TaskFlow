'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  templateAPI,
  getApiErrorMessage,
  CreateTaskTemplateDTO,
  UpdateTaskTemplateDTO,
  TaskTemplate,
  CreateTaskDTO,
} from '@/lib/api';
import { toast } from 'sonner';

/**
 * Hook for fetching all templates for the current user
 */
export function useTemplates() {
  return useQuery({
    queryKey: ['templates'],
    queryFn: async () => {
      const response = await templateAPI.list();
      return response.data;
    },
  });
}

/**
 * Hook for fetching a single template by ID
 */
export function useTemplate(templateId: string) {
  return useQuery({
    queryKey: ['templates', templateId],
    queryFn: async () => {
      const response = await templateAPI.getById(templateId);
      return response.data;
    },
    enabled: !!templateId,
  });
}

/**
 * Hook for creating a new template
 */
export function useCreateTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTaskTemplateDTO) => templateAPI.create(data),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      toast.success(`Template "${response.data.name}" created!`);
    },
    onError: (err: unknown) => {
      const message = getApiErrorMessage(err, 'Failed to create template', 'CreateTemplate');
      if (message.includes('already exists')) {
        toast.error('A template with this name already exists');
      } else {
        toast.error(message);
      }
    },
  });
}

/**
 * Hook for updating a template
 */
export function useUpdateTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTaskTemplateDTO }) =>
      templateAPI.update(id, data),
    onSuccess: (response, variables) => {
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      queryClient.invalidateQueries({ queryKey: ['templates', variables.id] });
      toast.success(`Template "${response.data.name}" updated!`);
    },
    onError: (err: unknown) => {
      const message = getApiErrorMessage(err, 'Failed to update template', 'UpdateTemplate');
      if (message.includes('already exists')) {
        toast.error('A template with this name already exists');
      } else {
        toast.error(message);
      }
    },
  });
}

/**
 * Hook for deleting a template
 */
export function useDeleteTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => templateAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      toast.success('Template deleted');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to delete template', 'DeleteTemplate'));
    },
  });
}

/**
 * Hook for using a template (get pre-filled CreateTaskDTO)
 * Does not show toast since this is just for fetching data to display in form
 */
export function useTemplateForTaskCreation() {
  return useMutation({
    mutationFn: ({ id, overrides }: { id: string; overrides?: Partial<CreateTaskDTO> }) =>
      templateAPI.use(id, overrides),
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to load template', 'UseTemplate'));
    },
  });
}

/**
 * Hook for saving a task as a new template
 * Accepts task data and creates a template from it
 */
export function useSaveAsTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      templateName,
      task,
    }: {
      templateName: string;
      task: {
        title: string;
        description?: string;
        category?: string;
        estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
        user_priority?: number;
        context?: string;
        related_people?: string[];
        due_date?: string;
      };
    }) => {
      // Convert task to template DTO
      // Calculate due_date_offset from due_date if present
      let dueDateOffset: number | undefined;
      if (task.due_date) {
        const dueDate = new Date(task.due_date);
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        dueDate.setHours(0, 0, 0, 0);
        const diffDays = Math.round((dueDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
        // Only set offset if positive and within valid range
        if (diffDays >= 0 && diffDays <= 365) {
          dueDateOffset = diffDays;
        }
      }

      const templateData: CreateTaskTemplateDTO = {
        name: templateName,
        title: task.title,
        description: task.description,
        category: task.category,
        estimated_effort: task.estimated_effort,
        user_priority: task.user_priority,
        context: task.context,
        related_people: task.related_people,
        due_date_offset: dueDateOffset,
      };

      return templateAPI.create(templateData);
    },
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['templates'] });
      toast.success(`Saved as template "${response.data.name}"`);
    },
    onError: (err: unknown) => {
      const message = getApiErrorMessage(err, 'Failed to save as template', 'SaveAsTemplate');
      if (message.includes('already exists')) {
        toast.error('A template with this name already exists');
      } else {
        toast.error(message);
      }
    },
  });
}

/**
 * Utility to convert a template to initial form values for CreateTaskDialog
 */
export function templateToFormValues(template: TaskTemplate): Partial<CreateTaskDTO> {
  // Calculate due date from offset
  let dueDate: string | undefined;
  if (template.due_date_offset !== undefined && template.due_date_offset !== null) {
    const date = new Date();
    date.setDate(date.getDate() + template.due_date_offset);
    dueDate = date.toISOString();
  }

  return {
    title: template.title,
    description: template.description,
    category: template.category,
    estimated_effort: template.estimated_effort,
    user_priority: template.user_priority !== 5 ? template.user_priority : undefined,
    context: template.context,
    related_people: template.related_people,
    due_date: dueDate,
  };
}
