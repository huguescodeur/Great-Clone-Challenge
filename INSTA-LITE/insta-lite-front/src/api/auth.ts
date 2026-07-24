import client from './client';
import type { User } from '../types';

interface AuthResponse {
  user: User;
  token: string;
}

export const register = (data: {
  username: string;
  email: string;
  fullName: string;
  password: string;
}) => client.post<AuthResponse>('/auth/register', data).then((r) => r.data);

export const login = (data: { identifier: string; password: string }) =>
  client.post<AuthResponse>('/auth/login', data).then((r) => r.data);

export const logout = () =>
  client.post('/auth/logout').then((r) => r.data);
