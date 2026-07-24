---
name: project-frontend
description: Insta-lite frontend React+Tailwind créé dans insta-lite-front/, connecté au back Go sur :8080
metadata:
  type: project
---

Frontend React+Tailwind pour insta-lite-back (Go/Chi/Postgres/Redis sur port 8080).

**Stack:** Vite + React 19 + TypeScript + Tailwind CSS v4 + TanStack Query v5 + Axios + Zustand + React Router v7 + react-intersection-observer

**Chemin:** `/Users/goliyaohugues/MyProjects/INSTA-LITE/insta-lite-front/`

**Why:** Couvre toutes les features du back : auth, feed, posts, likes (reactions like/love/laugh), commentaires avec replies, likes de commentaires, follow/unfollow, notifications avec badge.

**How to apply:** Lancer avec `npm run dev` depuis insta-lite-front/. Back doit tourner sur :8080.

**Structure src/:**
- `api/` : client axios + modules (auth, posts, feed, likes, comments, commentLikes, follow, notifications)
- `store/auth.ts` : Zustand store (persiste token/user dans localStorage)
- `types/index.ts` : tous les types TypeScript calqués sur les modèles Go
- `components/` : Layout, Navbar, Avatar, PostCard, CreatePostModal, CommentSection, CommentItem, ReactionPicker, FollowButton
- `pages/` : LoginPage, RegisterPage, FeedPage (feed perso), ExplorePage (tous les posts), ProfilePage (posts+followers+following), NotificationsPage
- `utils/date.ts` : formatDistanceToNow FR
