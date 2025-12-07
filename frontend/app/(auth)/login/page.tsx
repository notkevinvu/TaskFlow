'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import Link from "next/link";
import { useAuth } from '@/hooks/useAuth';
import { getApiErrorMessage } from '@/lib/api';
import { tokens } from '@/lib/tokens';
import { Sparkles } from 'lucide-react';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, startGuest, isLoading } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      await login(email, password);
      router.push('/dashboard');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Login failed. Please check your credentials.', 'Login'));
    }
  };

  const handleTryAsGuest = async () => {
    setError('');
    try {
      await startGuest();
      router.push('/dashboard');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to start guest session. Please try again.', 'GuestStart'));
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4" style={{ background: tokens.gradient.surface }}>
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold">Sign in to TaskFlow</CardTitle>
          <CardDescription>
            Enter your credentials to access your intelligent task manager
          </CardDescription>
        </CardHeader>
        <CardContent>
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
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
                disabled={isLoading}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                disabled={isLoading}
              />
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign In'}
            </Button>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">or</span>
              </div>
            </div>

            <Button
              type="button"
              variant="outline"
              className="w-full gap-2"
              onClick={handleTryAsGuest}
              disabled={isLoading}
            >
              <Sparkles className="h-4 w-4" />
              Try as Guest
            </Button>
            <p className="text-xs text-center text-muted-foreground">
              No sign-up required. Your data is saved for 30 days.
            </p>

            <div className="text-center text-sm text-muted-foreground">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="text-primary underline hover:text-primary/80">
                Sign up
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
