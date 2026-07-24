import client from './client';
import type { Like, ReactionType } from '../types';

export const getLikes = (postID: string) =>
  client.get<Like[]>(`/posts/${postID}/likes`).then((r) => r.data);

export const getMyLike = (postID: string) =>
  client.get<Like>(`/posts/${postID}/likes/me`).then((r) => r.data);

export const createLike = (postID: string, reactionType: ReactionType) =>
  client.post<Like>(`/posts/${postID}/likes`, { reactionType }).then((r) => r.data);

export const updateLike = (postID: string, reactionType: ReactionType) =>
  client.patch<Like>(`/posts/${postID}/likes`, { reactionType }).then((r) => r.data);

export const deleteLike = (postID: string) =>
  client.delete(`/posts/${postID}/likes`).then((r) => r.data);
