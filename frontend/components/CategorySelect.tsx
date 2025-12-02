'use client';

import { useState, useMemo } from 'react';
import { useTasks } from '@/hooks/useTasks';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface CategorySelectProps {
  value: string;
  onChange: (value: string) => void;
  id?: string;
}

const COMMON_CATEGORIES = [
  'Work',
  'Personal',
  'Meeting',
  'Code Review',
  'Bug Fix',
  'Documentation',
  'Planning',
  'Testing',
];

export function CategorySelect({ value, onChange, id }: CategorySelectProps) {
  const { data: tasksData } = useTasks();
  const [isCustomMode, setIsCustomMode] = useState(false);
  const [customInputValue, setCustomInputValue] = useState('');

  // Extract unique categories from existing tasks
  const existingCategories = useMemo(() => {
    if (!tasksData?.tasks) return [];

    const categories = new Set<string>();
    tasksData.tasks.forEach(task => {
      if (task.category && task.category.trim() !== '') {
        categories.add(task.category);
      }
    });

    return Array.from(categories).sort();
  }, [tasksData]);

  // Combine common categories with existing ones (remove duplicates)
  const allCategories = useMemo(() => {
    const combined = new Set([...COMMON_CATEGORIES, ...existingCategories]);
    return Array.from(combined).sort();
  }, [existingCategories]);

  // Check if current value is in the list - derived during render
  const valueInList = allCategories.includes(value);

  // Show custom input if user explicitly chose custom mode OR if the value isn't in the list
  const isCustom = isCustomMode || (value !== '' && !valueInList);

  // The displayed custom value is either user input or the prop value (for external custom values)
  const displayedCustomValue = isCustomMode ? customInputValue : value;

  const handleSelectChange = (newValue: string) => {
    if (newValue === '__custom__') {
      setIsCustomMode(true);
      setCustomInputValue('');
      onChange('');
    } else {
      setIsCustomMode(false);
      setCustomInputValue('');
      onChange(newValue);
    }
  };

  const handleCustomInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    // Limit to 50 characters (matching backend validation)
    if (newValue.length <= 50) {
      setCustomInputValue(newValue);
      onChange(newValue);
    }
  };

  const handleBackToSelect = () => {
    setIsCustomMode(false);
    setCustomInputValue('');
    onChange('');
  };

  if (isCustom) {
    return (
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Label htmlFor={id || 'custom-category'} className="text-sm">
            Custom Category
          </Label>
          <button
            type="button"
            onClick={handleBackToSelect}
            className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 underline"
          >
            Choose from list
          </button>
        </div>
        <Input
          id={id || 'custom-category'}
          value={displayedCustomValue}
          onChange={handleCustomInputChange}
          placeholder="Enter category name..."
          maxLength={50}
          autoFocus
        />
        <p className="text-xs text-muted-foreground">
          {displayedCustomValue.length}/50 characters
        </p>
      </div>
    );
  }

  return (
    <Select value={value} onValueChange={handleSelectChange}>
      <SelectTrigger id={id}>
        <SelectValue placeholder="Select category..." />
      </SelectTrigger>
      <SelectContent>
        {/* Show "Create custom" option */}
        <SelectItem value="__custom__" className="text-blue-600 dark:text-blue-400 font-medium">
          ✏️ Create custom category...
        </SelectItem>

        {/* Divider */}
        <div className="h-px bg-gray-200 dark:bg-gray-700 my-1" />

        {/* All categories */}
        {allCategories.map((category) => (
          <SelectItem key={category} value={category}>
            {category}
          </SelectItem>
        ))}

        {/* Empty state */}
        {allCategories.length === 0 && (
          <div className="px-2 py-6 text-center text-sm text-muted-foreground">
            No categories yet
          </div>
        )}
      </SelectContent>
    </Select>
  );
}
