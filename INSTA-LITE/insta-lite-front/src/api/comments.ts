import client from './client';
import type { Comment, PaginatedCommentResponse } from '../types';

export const getComments = (postID: string, cursor?: string, limit = 10) =>
  client
    .get<PaginatedCommentResponse>(`/posts/${postID}/comments`, { params: { cursor, limit } })
    .then((r) => r.data);

export const getReplies = (commentID: string) =>
  client.get<Comment[]>(`/comments/${commentID}/replies`).then((r) => r.data);

export const createComment = (postID: string, content: string, parentCommentId?: string) =>
  client
    .post<Comment>(`/posts/${postID}/comments`, { content, parentCommentId: parentCommentId ?? null })
    .then((r) => r.data);

export const updateComment = (commentID: string, content: string) =>
  client.patch<Comment>(`/comments/${commentID}`, { content }).then((r) => r.data);

export const deleteComment = (commentID: string) =>
  client.delete(`/comments/${commentID}`).then((r) => r.data);
