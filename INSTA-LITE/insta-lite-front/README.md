# insta-lite — frontend

React + TypeScript client for the Insta-Lite social platform.

## Tech stack

| Layer | Choice |
|---|---|
| Framework | React 19 |
| Language | TypeScript 6 |
| Build tool | Vite 8 |
| Styling | Tailwind CSS v4 |
| Routing | React Router v7 |
| Server state | TanStack Query v5 |
| Client state | Zustand v5 |
| HTTP client | Axios |
| Scroll detection | react-intersection-observer |
| Linter | oxlint |

## Prerequisites

- Node.js 20+
- The backend API running (see `../insta-lite-back/README.md`)

## Getting started

```bash
npm install
npm run dev        # starts the dev server on http://localhost:5173
```

## Environment variables

The API base URL is configured in `src/api/client.ts`:

```ts
baseURL: 'http://localhost:8080/api/v1'
```

If you need to point to a different backend (e.g. a staging server or a custom port), update this value directly or expose it via a Vite env variable:

1. Create a `.env.local` file at the root of this project:
   ```
   VITE_API_URL=http://localhost:8080/api/v1
   ```
2. Update `src/api/client.ts`:
   ```ts
   baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'
   ```

## Available scripts

| Script | Description |
|---|---|
| `npm run dev` | Start dev server with HMR |
| `npm run build` | Type-check and build for production |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run oxlint |

## Project structure

```
src/
  api/              Axios functions per domain (auth, posts, likes, comments, follow, feed, notifications, upload, users)
  components/       Reusable UI components (PostCard, CommentSection, ReactionPicker, FollowButton, Avatar, …)
  pages/            Route-level page components (FeedPage, ProfilePage, NotificationsPage, LoginPage, RegisterPage)
  store/            Zustand stores (auth — persisted to localStorage)
  types/            Shared TypeScript interfaces and types
  utils/            Helpers (date formatting, …)
```

## Key design decisions

### Authentication

The JWT token is stored in `localStorage` and attached to every request by an Axios interceptor. A 401 response on any non-auth endpoint clears the token and redirects to `/login`.

### Server state

TanStack Query manages all server state. Key patterns used:

- **Cursor-based infinite scroll**: `cursor` local state drives the query key; `react-intersection-observer` triggers the next page when the sentinel element enters the viewport.
- **Optimistic updates**: reactions (post likes, comment likes) use `onMutate` to update local state immediately with `onError` rollback, avoiding race conditions with the query cache.
- **Cache invalidation after follow**: following or unfollowing a user invalidates both the target's profile query (`['user', targetUserID]`) and the logged-in user's own profile query (`['user', me.userID]`) so follower/following counts update without a manual refresh.

### Reactions

Clicking an empty heart → direct `like` reaction.  
Clicking an already-liked heart → opens a picker showing all reaction types plus a ✕ button to unlike.  
Right-clicking the heart always opens the picker.

### Profile page

Always fetches the profile from `GET /users/{userID}` (TanStack Query, 30-second stale time) rather than relying on the auth store for display counts. The auth store value is only used as a fallback while the initial fetch is loading.
