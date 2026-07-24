import client from './client';
import type { PaginatedPostResponse } from '../types';

export const getFeed = (cursor?: string, limit = 10) =>
  client
    .get<PaginatedPostResponse>('/feed', { params: { cursor, limit } })
    .then((r) => r.data);
