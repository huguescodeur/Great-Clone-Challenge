import client from './client';
import type { User } from '../types';

export const getUserByID = (userID: string) =>
  client.get<User>(`/users/${userID}`).then((r) => r.data);
