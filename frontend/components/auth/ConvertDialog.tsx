'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuth } from '@/hooks/useAuth';
import { getApiErrorMessage } from '@/lib/api';
import { tokens } from '@/lib/tokens';
import { Check, Sparkles } from 'lucide-react';

interface ConvertDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ConvertDialog({ open, onOpenChange }: ConvertDialogProps) {
  const { convertToRegistered, isLoading } = useAuth();
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      await convertToRegistered({ email, name, password });
      onOpenChange(false);
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to create account. Please try again.', 'ConvertGuest'));
    }
  };

  const benefits = [
    'Keep all your tasks forever',
    'Access advanced features like subtasks, dependencies, and templates',
    'Track your achievements and streaks',
    'Use recurring tasks for habits',
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Sparkles className="h-5 w-5" style={{ color: tokens.highlight.strong }} />
            Create Your Account
          </DialogTitle>
          <DialogDescription>
            Keep your tasks and unlock all features
          </DialogDescription>
        </DialogHeader>

        <div className="mb-4 rounded-lg p-3" style={{ backgroundColor: tokens.highlight.muted }}>
          <p className="text-sm font-medium mb-2" style={{ color: tokens.highlight.strong }}>
            Benefits of registering:
          </p>
          <ul className="space-y-1">
            {benefits.map((benefit, i) => (
              <li key={i} className="flex items-center gap-2 text-sm text-muted-foreground">
                <Check className="h-4 w-4" style={{ color: tokens.status.success.default }} />
                {benefit}
              </li>
            ))}
          </ul>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div
              className="p-3 text-sm rounded border"
              style={{
                color: tokens.status.error.default,
                backgroundColor: tokens.status.error.muted,
                borderColor: tokens.status.error.default,
              }}
            >
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="convert-email">Email</Label>
            <Input
              id="convert-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              required
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="convert-name">Name</Label>
            <Input
              id="convert-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              required
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="convert-password">Password</Label>
            <Input
              id="convert-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Create a password"
              required
              disabled={isLoading}
              minLength={8}
            />
            <p className="text-xs text-muted-foreground">
              At least 8 characters
            </p>
          </div>

          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              className="flex-1"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Maybe Later
            </Button>
            <Button type="submit" className="flex-1" disabled={isLoading}>
              {isLoading ? 'Creating...' : 'Create Account'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
