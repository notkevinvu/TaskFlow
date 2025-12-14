# TaskFlow Learning Curriculum

Welcome to the TaskFlow Learning Curriculum! This comprehensive guide teaches modern full-stack development by studying a real-world project that was built from scratch over 28 days.

## What You'll Learn

By studying TaskFlow, you'll understand:

- **Full-Stack Architecture** - How frontend and backend work together
- **Clean Architecture in Go** - Domain-driven design with proper layering
- **Modern React Patterns** - React Query, Zustand, optimistic updates
- **Database Design** - Schema evolution, migrations, indexing strategies
- **Real Development Journey** - How software actually evolves (not just final state)
- **Refactoring Decisions** - When and why to refactor code
- **Technical Tradeoffs** - How to evaluate and document decisions

---

## About TaskFlow

TaskFlow is an intelligent task prioritization system that automatically calculates task priority using a multi-factor algorithm. It was built as a full-stack application with:

| Layer | Technology |
|-------|------------|
| **Frontend** | Next.js 16, React 19, TypeScript, Tailwind CSS, shadcn/ui |
| **Backend** | Go 1.24, Gin framework, Clean Architecture |
| **Database** | PostgreSQL 16 via Supabase |
| **State** | React Query (server state), Zustand (client state) |
| **Auth** | JWT with bcrypt password hashing |

**Development Stats:**
- 237 commits over 28 days
- 12 database migrations
- 84 merged pull requests
- ~6,000 lines of Go, ~11,000 lines of TypeScript

---

## Curriculum Structure

The curriculum is organized into 9 modules that build on each other:

### Foundation Modules (Start Here)

| Module | Title | Description |
|--------|-------|-------------|
| [01](./01-project-overview.md) | **Project Overview** | Tech stack rationale and architecture introduction |
| [02](./02-backend-architecture.md) | **Backend Architecture** | Clean Architecture layers and dependency injection |
| [03](./03-priority-algorithm.md) | **Priority Algorithm** | Business logic case study with explainable scoring |

### Deep Dive Modules

| Module | Title | Description |
|--------|-------|-------------|
| [04](./04-frontend-architecture.md) | **Frontend Architecture** | React Query patterns and optimistic updates |
| [05](./05-database-design.md) | **Database Design** | Schema evolution and migration strategies |
| [06](./06-development-journey.md) | **Development Journey** | The 0→1 story: how the project evolved |

### Advanced Modules

| Module | Title | Description |
|--------|-------|-------------|
| [07](./07-refactoring-case-studies.md) | **Refactoring Case Studies** | Before/after examples with impact analysis |
| [08](./08-technical-decisions.md) | **Technical Decisions** | Tradeoffs explained with rationale |
| [09](./09-lessons-learned.md) | **Lessons Learned** | Best practices and reusable patterns |

---

## Study Paths

Choose a path based on your experience level:

### Beginner Path (2-3 weeks)

If you're new to full-stack development, follow this order:

```
Week 1: Foundation
├── Module 01: Project Overview (2 days)
├── Module 03: Priority Algorithm (2-3 days)
└── Module 06: Development Journey (1 day)

Week 2: Frontend Focus
├── Module 04: Frontend Architecture (3-4 days)
└── Module 05: Database Design (2-3 days)

Week 3: Synthesis
├── Module 09: Lessons Learned (1 day)
└── Review and practice exercises
```

**Focus on:** Understanding patterns, running the code, completing exercises.

### Intermediate Path (1-2 weeks)

If you're familiar with web development basics:

```
Days 1-3: Architecture
├── Module 02: Backend Architecture (1.5 days)
├── Module 04: Frontend Architecture (1.5 days)

Days 4-7: Real-World Patterns
├── Module 06: Development Journey (1 day)
├── Module 07: Refactoring Case Studies (2 days)

Days 8-10: Decisions & Synthesis
├── Module 08: Technical Decisions (1.5 days)
└── Module 09: Lessons Learned (0.5 days)
```

**Focus on:** Architecture patterns, refactoring techniques, tradeoff analysis.

### Experienced Developer Path (3-5 days)

If you want to quickly extract key insights:

```
Day 1: Architecture Deep Dive
├── Module 02: Backend Architecture (skim structure, study DI patterns)
└── Module 04: Frontend Architecture (focus on React Query patterns)

Day 2: Refactoring & Decisions
├── Module 07: Refactoring Case Studies (especially N+1 elimination)
└── Module 08: Technical Decisions (tradeoff tables)

Day 3: Synthesis
└── Module 09: Lessons Learned (extract reusable patterns)

Days 4-5: Deep dive into specific areas of interest
```

**Focus on:** Reusable patterns, decision frameworks, case studies.

---

## Prerequisites

To get the most from this curriculum:

### Required Knowledge
- Basic programming experience (any language)
- Understanding of HTTP and REST APIs
- Familiarity with command line

### Helpful (but not required)
- React basics (components, hooks)
- SQL fundamentals (SELECT, INSERT, JOIN)
- Go basics (packages, functions)

### Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/notkevinvu/TaskFlow.git
   cd TaskFlow
   ```

2. **Explore the codebase:**
   ```bash
   # Backend structure
   ls backend/internal/

   # Frontend structure
   ls frontend/app/
   ls frontend/hooks/
   ```

3. **Optional: Run the application** (see main README.md for setup instructions)

---

## How to Use This Curriculum

### For Each Module

1. **Read the learning objectives** - Know what you're trying to learn
2. **Study the inline code snippets** - Understand the patterns
3. **Explore the referenced files** - See the full context
4. **Complete the exercises** - Reinforce your learning
5. **Reflect on the questions** - Connect concepts to your own experience

### Code References

Each module references specific files in the codebase. The format is:

```
backend/internal/domain/task.go:45-60
```

This means: file `backend/internal/domain/task.go`, lines 45-60.

### Exercises

Exercises are marked with:
- 🔰 **Beginner** - Straightforward, follows the examples
- 🎯 **Intermediate** - Requires some exploration
- 🚀 **Advanced** - Open-ended, requires creativity

---

## Key Concepts Preview

Before diving in, here are the core concepts you'll encounter:

### Clean Architecture (Backend)

```
┌─────────────────────────────────────────┐
│          HTTP Handlers (Adapters)        │  ← HTTP/JSON
├─────────────────────────────────────────┤
│          Services (Business Logic)       │  ← Orchestration
├─────────────────────────────────────────┤
│          Domain (Entities & Rules)       │  ← Pure business
├─────────────────────────────────────────┤
│          Repositories (Data Access)      │  ← Database
└─────────────────────────────────────────┘
         Dependencies flow inward →
```

### Server State Separation (Frontend)

```
┌─────────────────┐    ┌─────────────────┐
│   React Query   │    │     Zustand     │
│  (Server State) │    │  (Client State) │
├─────────────────┤    ├─────────────────┤
│ - Tasks         │    │ - Auth token    │
│ - Analytics     │    │ - UI state      │
│ - User data     │    │ - Theme         │
│ - Cached API    │    │ - Preferences   │
└─────────────────┘    └─────────────────┘
```

### Priority Algorithm

```
Score = (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
```

---

## Learning Tips

1. **Don't just read - explore the code.** Open the files and trace through the logic.

2. **Ask "why?"** For every pattern you see, ask why it was chosen over alternatives.

3. **Compare before/after.** The refactoring case studies show how code evolved - this is where real learning happens.

4. **Build something.** The best way to internalize patterns is to use them in your own project.

5. **Take notes.** Write down patterns you want to reuse in your own work.

---

## Quick Reference

### Key Backend Files
| Purpose | Path |
|---------|------|
| Entry point | `backend/cmd/server/main.go` |
| Domain entities | `backend/internal/domain/task.go` |
| Interfaces | `backend/internal/ports/` |
| Services | `backend/internal/service/task_service.go` |
| Handlers | `backend/internal/handler/task_handler.go` |
| Priority algorithm | `backend/internal/domain/priority/calculator.go` |

### Key Frontend Files
| Purpose | Path |
|---------|------|
| Root layout | `frontend/app/layout.tsx` |
| Dashboard | `frontend/app/(dashboard)/dashboard/page.tsx` |
| API client | `frontend/lib/api.ts` |
| Task hooks | `frontend/hooks/useTasks.ts` |
| Query keys | `frontend/lib/queryKeys.ts` |

### Key Documentation
| Purpose | Path |
|---------|------|
| Project status | `docs/PROJECT_STATUS.md` |
| Priority algorithm spec | `docs/product/priority-algorithm.md` |
| N+1 query fix | `docs/optimizations/n1-query-fix-explained.md` |
| Security audit | `docs/audits/findings-security.md` |

---

## Ready to Start?

Begin with **[Module 01: Project Overview](./01-project-overview.md)** to understand the tech stack and architecture philosophy.

Happy learning! 🎓
