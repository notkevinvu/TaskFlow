# TaskFlow - Intelligent Task Prioritization System

An intelligent task management system that automatically prioritizes tasks using multi-factor algorithms, helping you focus on what matters most.

## Features

- **Intelligent Prioritization** - Automatically calculates task priority based on multiple factors
- **Task Details Sidebar** - View comprehensive task information with smooth animations
- **Quick Task Creation** - Rapidly add new tasks with a streamlined modal interface
- **Analytics Dashboard** - Visualize your task patterns and completion rates
- **Responsive Design** - Works seamlessly on desktop, tablet, and mobile
- **Dark Mode Ready** - Built with light/dark mode support

## Tech Stack

### Frontend
- **Next.js 16** with App Router
- **React 19** with TypeScript
- **Tailwind CSS v4** for styling
- **shadcn/ui** component library
- **React Query** for data fetching
- **Zustand** for state management

### Backend (Phase 2+)
- **PostgreSQL 16** for data persistence
- **Python/FastAPI** for API endpoints

## Getting Started

### Prerequisites

- **Node.js** 20+ and npm
- **Docker** and Docker Compose (for database)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd TaskFlow
   ```

2. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   ```

3. **Start the database** (optional for Phase 1 - currently using mock data)
   ```bash
   # From project root
   docker compose up -d
   ```

   This starts:
   - **PostgreSQL** on `localhost:5432`
     - Database: `taskflow`
     - User: `taskflow_user`
     - Password: `taskflow_dev_password`
   - **pgAdmin** on `http://localhost:5050`
     - Email: `admin@taskflow.dev`
     - Password: `admin`

4. **Start the development server**
   ```bash
   cd frontend
   npm run dev
   ```

5. **Open your browser**
   ```
   http://localhost:3000
   ```

### Development Mode

The application runs in **development mode** by default with:
- Mock task data (6 sample tasks)
- Auto-login (no auth required)
- Mock user: `admin@taskflow.dev`

## Project Structure

```
TaskFlow/
├── frontend/                 # Next.js frontend application
│   ├── app/                 # Next.js App Router pages
│   │   ├── (auth)/         # Authentication pages
│   │   ├── (dashboard)/    # Dashboard pages
│   │   └── layout.tsx      # Root layout
│   ├── components/         # React components
│   │   ├── ui/            # shadcn/ui components
│   │   ├── CreateTaskDialog.tsx
│   │   └── TaskDetailsSidebar.tsx
│   ├── hooks/             # Custom React hooks
│   ├── lib/               # Utilities and API client
│   └── public/            # Static assets
├── docker-compose.yml      # Database containers
└── docs/                   # Documentation
```

## Available Scripts

### Frontend

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

### Docker

```bash
docker compose up -d       # Start database containers
docker compose down        # Stop containers
docker compose logs -f     # View container logs
```

## Phase 1 Completion Status

### ✅ Completed Features

- Frontend architecture with Next.js 16
- Authentication UI (login/register pages)
- Dashboard with task list and priority visualization
- Task details sidebar with smooth animations
- Quick task creation modal
- Analytics page with charts and metrics
- Mock data development workflow
- Docker setup for PostgreSQL

### 🚧 Pending (Future Phases)

- Backend API implementation
- Real database integration
- User authentication (backend)
- Task history tracking
- Priority algorithm refinement
- Design system improvements (see PRD)

## Design System

Currently using:
- **shadcn/ui** components with CSS variables
- **OKLCH color space** for modern color handling
- **Light/dark mode** ready
- **Semantic tokens** (primary, destructive, muted, etc.)

For future design system improvements and Figma integration plans, see the "Design System Improvements" section in `PRD.md`.

## Documentation

- **PRD.md** - Product Requirements Document
- **architecture-overview.md** - Technical architecture
- **data-model.md** - Database schema
- **priority-algorithm.md** - Prioritization logic
- **phase-1-weeks-1-2.md** - Implementation guide

## Contributing

This is a personal project, but feedback and suggestions are welcome!

## License

MIT

---

**Status:** Phase 1 Complete ✅
**Last Updated:** 2025-01-15
