import client from './client';
import type { Follow, FollowersResponse, FollowingResponse } from '../types';

export const createFollow = (userID: string) =>
  client.post<Follow>(`/users/${userID}/follow`).then((r) => r.data);

export const deleteFollow = (userID: string) =>
  client.delete(`/users/${userID}/follow`).then((r) => r.data);

export const getFollowStatus = (userID: string) =>
  client.get<Follow>(`/users/${userID}/follow/status`).then((r) => r.data);

export const getFollowers = (userID: string, page = 1, limit = 20) =>
  client
    .get<FollowersResponse>(`/users/${userID}/followers`, { params: { page, limit } })
    .then((r) => r.data);

export const getFollowing = (userID: string, page = 1, limit = 20) =>
  client
    .get<FollowingResponse>(`/users/${userID}/following`, { params: { page, limit } })
    .then((r) => r.data);
