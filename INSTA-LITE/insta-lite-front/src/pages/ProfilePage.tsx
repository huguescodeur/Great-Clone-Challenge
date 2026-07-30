import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';
import { getPostsByUser } from '../api/posts';
import { getFollowers, getFollowing } from '../api/follow';
import { getUserByUsername } from '../api/users';
import Avatar from '../components/Avatar';
import FollowButton from '../components/FollowButton';
import PostCard from '../components/PostCard';
import { useAuthStore } from '../store/auth';
import type { Follow, PostResponse } from '../types';

export default function ProfilePage() {
  const { username } = useParams<{ username: string }>();
  const { user: me } = useAuthStore();
  const [posts, setPosts] = useState<PostResponse[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(true);
  const [tab, setTab] = useState<'posts' | 'followers' | 'following'>('posts');
  const { ref, inView } = useInView({ threshold: 0.1 });

  const isMe = me?.username === username;

  const { data: profileUser } = useQuery({
    queryKey: ['user', username],
    queryFn: () => getUserByUsername(username!),
    enabled: !!username,
    staleTime: 30_000,
    placeholderData: isMe ? (me as typeof profileUser) : undefined,
  });

  const profileUserID = profileUser?.userID;

  const { data: postsData, isFetching } = useQuery({
    queryKey: ['user-posts', profileUserID, cursor],
    queryFn: () => getPostsByUser(profileUserID!, cursor),
    enabled: !!profileUserID && tab === 'posts',
  });

  const { data: followersData } = useQuery({
    queryKey: ['followers', profileUserID],
    queryFn: () => getFollowers(profileUserID!),
    enabled: !!profileUserID && tab === 'followers',
  });

  const { data: followingData } = useQuery({
    queryKey: ['following', profileUserID],
    queryFn: () => getFollowing(profileUserID!),
    enabled: !!profileUserID && tab === 'following',
  });

  useEffect(() => {
    if (!postsData) return;
    if (!cursor) {
      setPosts(postsData.data ?? []);
    } else {
      setPosts((prev) => [...prev, ...(postsData.data ?? [])]);
    }
    setHasMore(postsData.has_more);
  }, [postsData, cursor]);

  useEffect(() => {
    if (inView && hasMore && !isFetching && postsData?.next_cursor) setCursor(postsData.next_cursor);
  }, [inView, hasMore, isFetching, postsData]);

  useEffect(() => {
    setPosts([]); setCursor(undefined); setHasMore(true);
  }, [username]);

  if (!username) return null;

  const displayName = profileUser?.username || username;

  return (
    <div>
      {/* Header */}
      <div className="bg-white border-b border-[#dbdbdb] px-4 pt-8 pb-5 mb-0">
        <div className="flex items-start gap-8 mb-6">
          <Avatar name={displayName} url={profileUser?.profilePicURL} size="xl" ring />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-4 mb-4 flex-wrap">
              <h1 className="text-xl font-light text-[#262626]">{displayName}</h1>
              {!isMe && profileUserID && <FollowButton targetUserID={profileUserID} targetUsername={username} />}
              {isMe && (
                <button className="px-4 py-1.5 border border-[#dbdbdb] rounded-lg text-sm font-semibold text-[#262626] hover:bg-[#fafafa] cursor-pointer">
                  Modifier le profil
                </button>
              )}
            </div>

            {/* Stats */}
            <div className="flex gap-8 mb-3">
              <button onClick={() => setTab('posts')} className="text-center cursor-pointer hover:opacity-70">
                <span className="font-semibold text-[#262626]">{profileUser?.postsCount ?? posts.length}</span>
                {' '}
                <span className="text-[#262626]">publication{(profileUser?.postsCount ?? 0) > 1 ? 's' : ''}</span>
              </button>
              <button onClick={() => setTab('followers')} className="cursor-pointer hover:opacity-70 text-[#262626]">
                <span className="font-semibold">{profileUser?.followersCount ?? followersData?.total ?? '—'}</span>
                {' '}abonné{(profileUser?.followersCount ?? 0) > 1 ? 's' : ''}
              </button>
              <button onClick={() => setTab('following')} className="cursor-pointer hover:opacity-70 text-[#262626]">
                <span className="font-semibold">{profileUser?.followingCount ?? followingData?.total ?? '—'}</span>
                {' '}abonnement{(profileUser?.followingCount ?? 0) > 1 ? 's' : ''}
              </button>
            </div>

            {profileUser?.fullName && (
              <p className="text-sm font-semibold text-[#262626]">{profileUser.fullName}</p>
            )}
            {profileUser?.bio && (
              <p className="text-sm text-[#262626] whitespace-pre-line">{profileUser.bio}</p>
            )}
          </div>
        </div>

        {/* Tab bar */}
        <div className="flex border-t border-[#dbdbdb] -mx-4">
          {[
            { key: 'posts' as const, icon: <svg width="12" height="12" fill="currentColor" viewBox="0 0 24 24"><path d="M2 2h8.99v8.99H2zm0 11.01h8.99V22H2zM13.01 2H22v8.99h-8.99zm0 11.01H22V22h-8.99z"/></svg>, label: 'PUBLICATIONS' },
          ].map(({ key, icon, label }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex-1 flex items-center justify-center gap-1 py-3 text-[11px] font-semibold tracking-wider cursor-pointer transition-colors ${
                tab === key
                  ? 'text-[#262626] border-t border-[#262626] -mt-px'
                  : 'text-[#8e8e8e] hover:text-[#262626]'
              }`}
            >
              {icon}{label}
            </button>
          ))}
        </div>
      </div>

      {/* Posts grid */}
      {tab === 'posts' && (
        <div>
          {posts.length === 0 && !isFetching && (
            <div className="text-center py-16 bg-white border-b border-[#dbdbdb]">
              <svg width="62" height="62" fill="none" viewBox="0 0 24 24" stroke="#262626" strokeWidth={0.8} className="mx-auto mb-4">
                <path strokeLinecap="round" strokeLinejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"/><circle cx="12" cy="13" r="3"/>
              </svg>
              <h3 className="text-xl font-light text-[#262626] mb-1">Aucune publication</h3>
              {isMe && <p className="text-sm text-[#8e8e8e]">Partage ta première photo</p>}
            </div>
          )}
          {posts.map((p) => <PostCard key={p.postID} post={p} />)}
          {isFetching && (
            <div className="flex justify-center py-5">
              <div className="w-7 h-7 border-2 border-[#dbdbdb] border-t-[#262626] rounded-full animate-spin" />
            </div>
          )}
          {hasMore && <div ref={ref} className="h-4" />}
        </div>
      )}

      {/* Followers list */}
      {tab === 'followers' && (
        <div className="bg-white">
          {!followersData && (
            <div className="flex justify-center py-8">
              <div className="w-7 h-7 border-2 border-[#dbdbdb] border-t-[#262626] rounded-full animate-spin" />
            </div>
          )}
          {(followersData?.followers?.length ?? 0) === 0 && followersData && (
            <p className="text-center text-[#8e8e8e] py-10 text-sm">Aucun abonné</p>
          )}
          {(followersData?.followers ?? []).map((f, i, arr) => (
            <FollowRow key={f.followerId + i} follow={f} displayID={f.followerId} isLast={i === arr.length - 1} />
          ))}
        </div>
      )}

      {/* Following list */}
      {tab === 'following' && (
        <div className="bg-white">
          {!followingData && (
            <div className="flex justify-center py-8">
              <div className="w-7 h-7 border-2 border-[#dbdbdb] border-t-[#262626] rounded-full animate-spin" />
            </div>
          )}
          {(followingData?.following?.length ?? 0) === 0 && followingData && (
            <p className="text-center text-[#8e8e8e] py-10 text-sm">Aucun abonnement</p>
          )}
          {(followingData?.following ?? []).map((f, i, arr) => (
            <FollowRow key={f.followeeId + i} follow={f} displayID={f.followeeId} isLast={i === arr.length - 1} />
          ))}
        </div>
      )}
    </div>
  );
}

function FollowRow({ follow, displayID, isLast }: { follow: Follow; displayID: string; isLast: boolean }) {
  return (
    <div className={`flex items-center gap-3 px-4 py-3 ${!isLast ? 'border-b border-[#dbdbdb]' : ''}`}>
      <Avatar name={displayID} size="md" />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-[#262626] truncate">{displayID.slice(0, 12)}</p>
        <p className="text-xs text-[#8e8e8e]">
          {follow.status === 'accepted' ? 'Abonné(e)' : 'En attente'}
        </p>
      </div>
      <FollowButton targetUserID={displayID} />
    </div>
  );
}
