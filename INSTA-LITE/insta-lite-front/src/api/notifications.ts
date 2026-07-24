import client from './client';
import type { PaginatedNotificationResponse } from '../types';

export const getNotifications = (cursor?: string, limit = 20) =>
  client
    .get<PaginatedNotificationResponse>('/notifications', { params: { cursor, limit } })
    .then((r) => r.data);

export const markAsRead = (notificationID: string) =>
  client.patch(`/notifications/${notificationID}/read`).then((r) => r.data);
