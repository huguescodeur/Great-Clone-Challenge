import client from './client';
import type { User } from '../types';

export const getUserByID = (userID: string) =>
  client.get<User>(`/users/${userID}`).then((r) => r.data);

export const getUserByUsername = (username: string) =>
  client.get<User>(`/users/by-username/${username}`).then((r) => r.data);
