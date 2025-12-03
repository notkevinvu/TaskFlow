'use client';

import { Button } from "@/components/ui/button";
import { Trash2, RotateCcw, X } from "lucide-react";

interface BulkActionsBarProps {
  selectedCount: number;
  onDelete: () => void;
  onRestore: () => void;
  onClearSelection: () => void;
  isDeleting?: boolean;
  isRestoring?: boolean;
}

export function BulkActionsBar({
  selectedCount,
  onDelete,
  onRestore,
  onClearSelection,
  isDeleting,
  isRestoring,
}: BulkActionsBarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className="flex items-center justify-between p-3 bg-muted rounded-lg">
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium">
          {selectedCount} task{selectedCount !== 1 ? 's' : ''} selected
        </span>
        <Button
          variant="ghost"
          size="sm"
          onClick={onClearSelection}
          className="h-7 px-2"
        >
          <X className="h-4 w-4 mr-1" />
          Clear
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onRestore}
          disabled={isRestoring || isDeleting}
        >
          <RotateCcw className="h-4 w-4 mr-2" />
          {isRestoring ? 'Restoring...' : 'Restore to Active'}
        </Button>
        <Button
          variant="destructive"
          size="sm"
          onClick={onDelete}
          disabled={isDeleting || isRestoring}
        >
          <Trash2 className="h-4 w-4 mr-2" />
          {isDeleting ? 'Deleting...' : 'Delete'}
        </Button>
      </div>
    </div>
  );
}
