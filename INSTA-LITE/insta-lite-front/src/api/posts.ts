import client from './client';
import type { PaginatedPostResponse, PostResponse, PostMedia } from '../types';

export const getAllPosts = (cursor?: string, limit = 10) =>
  client
    .get<PaginatedPostResponse>('/posts', { params: { cursor, limit } })
    .then((r) => r.data);

export const getPostsByUser = (userID: string, cursor?: string, limit = 10) =>
  client
    .get<PaginatedPostResponse>(`/posts/${userID}`, { params: { cursor, limit } })
    .then((r) => r.data);

export const createPost = (content: string, medias: Omit<PostMedia, 'mediaID' | 'postID' | 'createdAt'>[]) =>
  client
    .post<PostResponse>('/posts', { content, medias })
    .then((r) => r.data);

export const updatePost = (postID: string, content: string) =>
  client
    .patch<PostResponse>('/posts', { postID, content })
    .then((r) => r.data);

export const deletePost = (postID: string) =>
  client.delete('/posts', { data: { postID } }).then((r) => r.data);
