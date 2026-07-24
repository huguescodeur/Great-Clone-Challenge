import client from './client';
import type { CommentLike, ReactionType } from '../types';

export const getCommentLikes = (commentID: string) =>
  client.get<CommentLike[]>(`/comments/${commentID}/likes`).then((r) => r.data);

export const getMyCommentLike = (commentID: string) =>
  client.get<CommentLike>(`/comments/${commentID}/likes/me`).then((r) => r.data);

export const createCommentLike = (commentID: string, reactionType: ReactionType) =>
  client
    .post<CommentLike>(`/comments/${commentID}/likes`, { reactionType })
    .then((r) => r.data);

export const updateCommentLike = (commentID: string, reactionType: ReactionType) =>
  client
    .patch<CommentLike>(`/comments/${commentID}/likes`, { reactionType })
    .then((r) => r.data);

export const deleteCommentLike = (commentID: string) =>
  client.delete(`/comments/${commentID}/likes`).then((r) => r.data);
