# Phase 1: Frontend & Database Setup (Weeks 1-2)
## Intelligent Task Prioritization System

**Goal:** Set up Next.js frontend with PostgreSQL database and task management UI components.

**By the end of this phase, you will have:**
- ✅ Next.js 15 project with TypeScript
- ✅ PostgreSQL database running
- ✅ Shadcn/UI components integrated
- ✅ Basic authentication UI (frontend only, no backend yet)
- ✅ Task list dashboard with priority visualization
- ✅ Quick task creation form
- ✅ Docker development environment

**What you're building:** A smart task manager that captures commitments from meetings/discussions and automatically prioritizes them based on deadlines, age, and user-set importance.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Week 1: Frontend Setup](#week-1-frontend-setup)
- [Week 2: Database & UI Components](#week-2-database--ui-components)
- [Testing Your Setup](#testing-your-setup)
- [Common Issues](#common-issues)

---

## Prerequisites

### Install Required Software

1. **Node.js & npm**
   ```bash
   # Check if installed
   node --version  # Should be v20 or higher
   npm --version

   # If not installed:
   # Download from https://nodejs.org/ (LTS version)
   # OR use nvm (recommended):
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
   nvm install 20
   nvm use 20
   ```

2. **PostgreSQL**
   ```bash
   # macOS
   brew install postgresql@16

   # Ubuntu/Debian
   sudo apt update
   sudo apt install postgresql-16

   # Windows
   # Download installer from https://www.postgresql.org/download/windows/

   # Verify installation
   psql --version
   ```

3. **Docker** (recommended for easier setup)
   ```bash
   # Download Docker Desktop from https://www.docker.com/products/docker-desktop/

   # Verify installation
   docker --version
   docker-compose --version
   ```

4. **Git**
   ```bash
   git --version
   # If not installed, download from https://git-scm.com/
   ```

---

## Week 1: Frontend Setup

### Day 1: Create Next.js Project

**Step 1: Initialize Project**

```bash
# Create project root directory
mkdir web-app
cd web-app

# Create Next.js project
npx create-next-app@latest frontend

# When prompted, answer:
# ✔ Would you like to use TypeScript? › Yes
# ✔ Would you like to use ESLint? › Yes
# ✔ Would you like to use Tailwind CSS? › Yes
# ✔ Would you like to use `src/` directory? › No
# ✔ Would you like to use App Router? › Yes
# ✔ Would you like to customize the default import alias (@/*)? › No
```

**Step 2: Verify Installation**

```bash
cd frontend
npm run dev
```

Visit `http://localhost:3000` - you should see the Next.js welcome page.

**Step 3: Clean Up Default Files**

```bash
# In frontend directory
rm app/page.tsx
rm app/globals.css
```

Create new `app/globals.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 212.7 26.8% 83.9%;
  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

Create new `app/page.tsx`:

```typescript
export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <h1 className="text-4xl font-bold">TaskFlow</h1>
      <p className="mt-4 text-xl text-muted-foreground">
        Never lose track of commitments. Auto-prioritize what matters.
      </p>
    </main>
  );
}
```

---

### Day 2-3: Install Shadcn/UI

**Step 1: Initialize Shadcn**

```bash
npx shadcn@latest init

# When prompted:
# ✔ Which style would you like to use? › Default
# ✔ Which color would you like to use as base color? › Slate
# ✔ Would you like to use CSS variables for colors? › Yes
```

This creates:
- `components/ui/` directory
- `lib/utils.ts` file
- Updates `tailwind.config.ts`

**Step 2: Add Essential Components**

```bash
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add input
npx shadcn@latest add label
npx shadcn@latest add dropdown-menu
npx shadcn@latest add avatar
npx shadcn@latest add separator
```

**Step 3: Test Components**

Update `app/page.tsx`:

```typescript
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import Link from "next/link";

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <Card className="w-96">
        <CardHeader>
          <CardTitle>Welcome to TaskFlow</CardTitle>
          <CardDescription>
            Intelligent task prioritization for busy professionals
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Link href="/login">
            <Button className="w-full">Get Started</Button>
          </Link>
        </CardContent>
      </Card>
    </main>
  );
}
```

You should see a styled card with a button!

---

### Day 4-5: Create Basic Layout

**Step 1: Create Layout Components**

Create `components/layout/Header.tsx`:

```typescript
import Link from "next/link";
import { Button } from "@/components/ui/button";

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        <div className="mr-4 flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <span className="font-bold">TaskFlow</span>
          </Link>
          <nav className="flex items-center space-x-6 text-sm font-medium">
            <Link href="/dashboard">Tasks</Link>
            <Link href="/dashboard/analytics">Analytics</Link>
          </nav>
        </div>
        <div className="ml-auto flex items-center space-x-4">
          <Button variant="ghost">Sign In</Button>
          <Button>Sign Up</Button>
        </div>
      </div>
    </header>
  );
}
```

Create `components/layout/Footer.tsx`:

```typescript
export function Footer() {
  return (
    <footer className="border-t">
      <div className="container flex h-14 items-center justify-between">
        <p className="text-sm text-muted-foreground">
          © 2025 TaskFlow. All rights reserved.
        </p>
        <nav className="flex items-center space-x-4 text-sm text-muted-foreground">
          <a href="#">Privacy</a>
          <a href="#">Terms</a>
          <a href="#">Contact</a>
        </nav>
      </div>
    </footer>
  );
}
```

**Step 2: Update Root Layout**

Update `app/layout.tsx`:

```typescript
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Header } from "@/components/layout/Header";
import { Footer } from "@/components/layout/Footer";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "TaskFlow - Intelligent Task Prioritization",
  description: "Never lose track of commitments. Auto-prioritize what matters.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="flex min-h-screen flex-col">
          <Header />
          <main className="flex-1">{children}</main>
          <Footer />
        </div>
      </body>
    </html>
  );
}
```

---

### Day 6-7: Create Authentication Pages

**Step 1: Create Auth Route Group**

```bash
mkdir -p app/\(auth\)/login
mkdir -p app/\(auth\)/register
```

**Step 2: Create Login Page**

Create `app/(auth)/login/page.tsx`:

```typescript
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import Link from "next/link";

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Sign In</CardTitle>
          <CardDescription>
            Enter your email and password to access your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                required
              />
            </div>
            <Button type="submit" className="w-full">
              Sign In
            </Button>
            <div className="text-center text-sm">
              Don't have an account?{" "}
              <Link href="/register" className="underline">
                Sign up
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

**Step 3: Create Register Page**

Create `app/(auth)/register/page.tsx`:

```typescript
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import Link from "next/link";

export default function RegisterPage() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create Account</CardTitle>
          <CardDescription>
            Enter your information to create an account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Full Name</Label>
              <Input
                id="name"
                type="text"
                placeholder="John Doe"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                required
              />
            </div>
            <Button type="submit" className="w-full">
              Create Account
            </Button>
            <div className="text-center text-sm">
              Already have an account?{" "}
              <Link href="/login" className="underline">
                Sign in
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

**Test:** Visit http://localhost:3000/login and http://localhost:3000/register

---

## Week 2: Database & UI Components

### Day 8-9: Setup PostgreSQL with Docker

**Step 1: Create docker-compose.yml**

Create `docker-compose.yml` in project root (web-app directory):

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: web-app-postgres
    environment:
      POSTGRES_DB: webapp_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  pgadmin:
    image: dpage/pgadmin4:latest
    container_name: web-app-pgadmin
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@admin.com
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

**Step 2: Start Database**

```bash
# From project root
docker-compose up -d

# Check if running
docker ps

# View logs
docker-compose logs postgres
```

**Step 3: Access pgAdmin**

1. Visit http://localhost:5050
2. Login with email: `admin@admin.com`, password: `admin`
3. Add server:
   - Host: `postgres` (container name)
   - Port: `5432`
   - Database: `webapp_dev`
   - Username: `postgres`
   - Password: `postgres`

**Alternative: Use psql**

```bash
# Connect to database
docker exec -it web-app-postgres psql -U postgres -d webapp_dev

# Inside psql:
\dt   # List tables
\q    # Quit
```

---

### Day 10-11: Create Dashboard Layout

**Step 1: Add More Shadcn Components**

```bash
cd frontend
npx shadcn@latest add badge
npx shadcn@latest add table
npx shadcn@latest add tabs
npx shadcn@latest add sheet
```

**Step 2: Create Dashboard Directory**

```bash
mkdir -p app/dashboard/analytics
mkdir -p app/dashboard/settings
```

**Step 3: Create Dashboard Layout**

Create `app/dashboard/layout.tsx`:

```typescript
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import Link from "next/link";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-64 border-r bg-muted/40">
        <div className="flex h-14 items-center border-b px-4">
          <Link href="/dashboard" className="font-semibold">
            TaskFlow
          </Link>
        </div>
        <nav className="space-y-1 p-4">
          <Link
            href="/dashboard"
            className="flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-accent"
          >
            My Tasks
          </Link>
          <Link
            href="/dashboard/analytics"
            className="flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-accent"
          >
            Analytics
          </Link>
          <Link
            href="/dashboard/settings"
            className="flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-accent"
          >
            Settings
          </Link>
        </nav>
      </aside>

      {/* Main Content */}
      <div className="flex-1">
        <header className="flex h-14 items-center justify-between border-b px-6">
          <h1 className="text-lg font-semibold">My Tasks</h1>
          <div className="flex items-center gap-4">
            <Button size="sm">+ Quick Add</Button>
            <Avatar>
              <AvatarFallback>JD</AvatarFallback>
            </Avatar>
          </div>
        </header>
        <main className="p-6">{children}</main>
      </div>
    </div>
  );
}
```

**Step 4: Create Task Dashboard Page**

Create `app/dashboard/page.tsx`:

```typescript
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export default function DashboardPage() {
  // Mock data - will be replaced with real data from API in Phase 2
  const mockTasks = [
    {
      id: 1,
      title: "Review design doc for auth flow",
      priority: 95,
      dueDate: "Tomorrow",
      category: "Code Review",
      context: "From Alice - needs feedback by Friday",
      bumpCount: 0,
    },
    {
      id: 2,
      title: "Update README with deployment instructions",
      priority: 78,
      dueDate: "In 3 days",
      category: "Documentation",
      context: "Tech debt item",
      bumpCount: 2,
    },
    {
      id: 3,
      title: "Fix bug in user authentication",
      priority: 85,
      dueDate: "Today",
      category: "Bug Fix",
      context: "Production issue",
      bumpCount: 0,
    },
  ];

  const atRiskTasks = mockTasks.filter(t => t.bumpCount >= 2);
  const quickWins = mockTasks.filter(t => t.priority < 80);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Today's Priorities</h2>
        <p className="text-muted-foreground">
          Focus on what matters most. {mockTasks.length} tasks waiting.
        </p>
      </div>

      {/* Quick Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">At Risk</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{atRiskTasks.length}</div>
            <p className="text-xs text-muted-foreground">
              Tasks bumped 2+ times
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Quick Wins</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{quickWins.length}</div>
            <p className="text-xs text-muted-foreground">
              Small tasks, easy to complete
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Completion Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">68%</div>
            <p className="text-xs text-muted-foreground">
              Last 7 days
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Priority Tasks */}
      <Card>
        <CardHeader>
          <CardTitle>Priority Tasks</CardTitle>
          <CardDescription>
            Sorted by intelligent priority score
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {mockTasks.map((task) => (
              <div
                key={task.id}
                className="flex items-start gap-4 rounded-lg border p-4 hover:bg-accent/50 transition-colors"
              >
                <div className="flex-1 space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge variant={task.priority >= 90 ? "destructive" : task.priority >= 75 ? "default" : "secondary"}>
                      {task.priority}
                    </Badge>
                    <h3 className="font-medium">{task.title}</h3>
                    {task.bumpCount > 0 && (
                      <Badge variant="outline" className="ml-auto">
                        Bumped {task.bumpCount}x
                      </Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground">{task.context}</p>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span>Due: {task.dueDate}</span>
                    <span>•</span>
                    <span>{task.category}</span>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline">Bump</Button>
                  <Button size="sm">Complete</Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
```

---

### Day 12-13: Add Task Analytics Page

Create `app/dashboard/analytics/page.tsx`:

```typescript
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Task Analytics</h2>
        <p className="text-muted-foreground">
          Understand your work patterns and improve task completion.
        </p>
      </div>

      <Tabs defaultValue="delays" className="space-y-4">
        <TabsList>
          <TabsTrigger value="delays">Delay Analysis</TabsTrigger>
          <TabsTrigger value="velocity">Velocity</TabsTrigger>
          <TabsTrigger value="categories">By Category</TabsTrigger>
        </TabsList>

        <TabsContent value="delays" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle>Avg Bump Count</CardTitle>
                <CardDescription>Per task (last 30 days)</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">2.3</div>
                <p className="text-xs text-muted-foreground">
                  +0.5 from last month
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>At-Risk Tasks</CardTitle>
                <CardDescription>Bumped 3+ times</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">12</div>
                <p className="text-xs text-muted-foreground">
                  8% of active tasks
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Avg Time to Complete</CardTitle>
                <CardDescription>From creation to done</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">5.2 days</div>
                <p className="text-xs text-muted-foreground">
                  Target: 3 days
                </p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Delay Patterns</CardTitle>
              <CardDescription>Tasks you tend to postpone (Coming in Phase 3)</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64 flex items-center justify-center border-2 border-dashed rounded-md">
                <p className="text-muted-foreground">Bump frequency chart will appear here</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Most Bumped Categories</CardTitle>
              <CardDescription>Which types of tasks do you avoid?</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm">Documentation</span>
                  <span className="text-sm font-medium">Avg 4.2 bumps</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">Code Review</span>
                  <span className="text-sm font-medium">Avg 2.8 bumps</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">Bug Fix</span>
                  <span className="text-sm font-medium">Avg 1.3 bumps</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="velocity">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Tasks Completed</CardTitle>
                <CardDescription>Last 7 days</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">23 tasks</div>
                <p className="text-xs text-muted-foreground">
                  68% completion rate
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Average Velocity</CardTitle>
                <CardDescription>Tasks per week</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">15.5</div>
                <p className="text-xs text-muted-foreground">
                  +3 from last month
                </p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Completion Trend</CardTitle>
              <CardDescription>Chart coming in Phase 3</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64 flex items-center justify-center border-2 border-dashed rounded-md">
                <p className="text-muted-foreground">Line chart: tasks completed over time</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="categories">
          <Card>
            <CardHeader>
              <CardTitle>Task Breakdown by Category</CardTitle>
              <CardDescription>Where your time goes</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Code Review</span>
                    <span className="text-sm text-muted-foreground">40%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full">
                    <div className="h-2 bg-primary rounded-full" style={{width: '40%'}}></div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Bug Fix</span>
                    <span className="text-sm text-muted-foreground">30%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full">
                    <div className="h-2 bg-primary rounded-full" style={{width: '30%'}}></div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Documentation</span>
                    <span className="text-sm text-muted-foreground">20%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full">
                    <div className="h-2 bg-primary rounded-full" style={{width: '20%'}}></div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Meeting Follow-up</span>
                    <span className="text-sm text-muted-foreground">10%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full">
                    <div className="h-2 bg-primary rounded-full" style={{width: '10%'}}></div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

---

### Day 14: Environment Variables & Configuration

**Step 1: Create Environment File**

Create `frontend/.env.local`:

```bash
# API Configuration (we'll use this in Phase 2)
NEXT_PUBLIC_API_URL=http://localhost:8080

# Database (for reference, won't connect from frontend directly)
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/webapp_dev
```

Create `frontend/.env.example`:

```bash
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080

# Database URL
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/webapp_dev
```

**Step 2: Create API Configuration**

Create `frontend/lib/api.ts`:

```typescript
import axios from 'axios';

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for adding auth token
api.interceptors.request.use(
  (config) => {
    // We'll add token logic in Phase 2
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for handling errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

---

## Testing Your Setup

### Checklist

- [ ] Next.js dev server runs: `npm run dev`
- [ ] Can access http://localhost:3000
- [ ] Login page works: http://localhost:3000/login
- [ ] Register page works: http://localhost:3000/register
- [ ] Dashboard works: http://localhost:3000/dashboard
- [ ] Analytics page works: http://localhost:3000/dashboard/analytics
- [ ] PostgreSQL running: `docker ps` shows postgres container
- [ ] Can access pgAdmin: http://localhost:5050

### Run Full Test

```bash
# Terminal 1: Start database
docker-compose up -d

# Terminal 2: Start frontend
cd frontend
npm run dev

# Open browser
# Visit all pages and verify they load correctly
```

---

## Common Issues

### Issue: Port 3000 already in use

**Solution:**
```bash
# Find process using port 3000
lsof -i :3000  # macOS/Linux
netstat -ano | findstr :3000  # Windows

# Kill the process or change Next.js port
# In package.json, change: "dev": "next dev -p 3001"
```

### Issue: PostgreSQL won't start

**Solution:**
```bash
# Check Docker logs
docker-compose logs postgres

# Remove and recreate
docker-compose down -v
docker-compose up -d
```

### Issue: Shadcn components not styling correctly

**Solution:**
1. Check `tailwind.config.ts` was updated by shadcn
2. Verify `globals.css` has the CSS variables
3. Restart dev server: `Ctrl+C` then `npm run dev`

### Issue: Module not found errors

**Solution:**
```bash
# Delete and reinstall
rm -rf node_modules package-lock.json
npm install
```

---

## Phase 1 Summary

**What you've built:**
- ✅ Next.js 15 with TypeScript and Tailwind CSS
- ✅ Shadcn/UI component library integrated
- ✅ Authentication UI (login/register pages)
- ✅ **Task dashboard** with priority-sorted task list
- ✅ **Priority badges** showing computed scores (0-100)
- ✅ **At-risk indicators** for tasks bumped multiple times
- ✅ **Analytics dashboard** with delay analysis and velocity metrics
- ✅ PostgreSQL database running in Docker
- ✅ pgAdmin for database management

**Key Features Displayed:**
- Priority-sorted task cards with context
- Visual indicators: red (90+), blue (75-89), gray (<75)
- Bump count tracking
- Quick stats: At Risk, Quick Wins, Completion Rate
- Analytics tabs: Delay Analysis, Velocity, Category Breakdown

**What's next (Phase 2):**
- Build Go backend with Gin framework
- Implement **priority calculation algorithm**
- Connect backend to PostgreSQL with sqlc
- Create task CRUD API endpoints
- Implement bump tracking
- Implement JWT authentication
- Connect frontend to live backend data

**You're ready for Phase 2!** 🎉

You now have a complete task management UI. In Phase 2, we'll bring it to life with the intelligent priority algorithm and real data.

See `phase-2-weeks-3-4.md` for backend implementation with priority calculation.
