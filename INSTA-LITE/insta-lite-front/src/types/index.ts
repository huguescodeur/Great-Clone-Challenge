export interface User {
  userID: string;
  username: string;
  email: string;
  fullName: string;
  bio: string;
  profilePicURL: string;
  followersCount: number;
  followingCount: number;
  postsCount: number;
  createdAt: string;
}

export interface PostMedia {
  mediaID: string;
  postID: string;
  url: string;
  type: 'image' | 'video';
  createdAt: string;
}

export interface Post {
  postID: string;
  userID: string;
  username: string;
  content: string;
  likesCount: number;
  commentsCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface PostResponse extends Post {
  media: PostMedia[];
}

export interface PaginatedPostResponse {
  data: PostResponse[];
  next_cursor: string;
  has_more: boolean;
}

export type ReactionType = 'like' | 'love' | 'laugh';

export interface Like {
  postID: string;
  userID: string;
  reactionType: ReactionType;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  commentId: string;
  postId: string;
  userId: string;
  username: string;
  parentCommentId?: string | null;
  content: string;
  likesCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedCommentResponse {
  data: Comment[];
  nextCursor: string;
  hasMore: boolean;
}

export interface CommentLike {
  commentId: string;
  userId: string;
  reactionType: ReactionType;
  createdAt: string;
  updatedAt: string;
}

export type FollowStatus = 'pending' | 'accepted';

export interface Follow {
  followerId: string;
  followeeId: string;
  followerUsername?: string;
  followeeUsername?: string;
  status: FollowStatus;
  createdAt: string;
}

export type NotificationType = 'like_post' | 'comment_post' | 'like_comment' | 'new_follower';

export interface Notification {
  notificationId: string;
  recipientId: string;
  actorId: string;
  actorUsername: string;
  type: NotificationType;
  entityId?: string | null;
  read: boolean;
  createdAt: string;
}

export interface PaginatedNotificationResponse {
  data: Notification[];
  nextCursor: string;
  hasMore: boolean;
  unreadCount: number;
}

export interface FollowersResponse {
  followers: Follow[];
  total: number;
  page: number;
  limit: number;
}

export interface FollowingResponse {
  following: Follow[];
  total: number;
  page: number;
  limit: number;
}

export interface SignedURLResponse {
  signature: string;
  timestamp: number;
  apiKey: string;
  cloudName: string;
  uploadUrl: string;
  folder: string;
}
