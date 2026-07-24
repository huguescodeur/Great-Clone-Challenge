import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';
import { getFeed } from '../api/feed';
import PostCard from '../components/PostCard';
import { useAuthStore } from '../store/auth';
import { useNavigate } from 'react-router-dom';
import type { PostResponse } from '../types';

export default function FeedPage() {
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const [posts, setPosts] = useState<PostResponse[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(true);
  const { ref, inView } = useInView({ threshold: 0.1 });

  useEffect(() => { if (!user) navigate('/login'); }, [user, navigate]);

  const { data, isFetching } = useQuery({
    queryKey: ['feed', cursor],
    queryFn: () => getFeed(cursor),
    enabled: !!user,
  });

  useEffect(() => {
    if (!data) return;
    if (!cursor) { setPosts(data.data ?? []); }
    else { setPosts((prev) => [...prev, ...(data.data ?? [])]); }
    setHasMore(data.has_more);
  }, [data, cursor]);

  useEffect(() => {
    if (inView && hasMore && !isFetching && data?.next_cursor) setCursor(data.next_cursor);
  }, [inView, hasMore, isFetching, data]);

  if (!user) return null;

  return (
    <div>
      {posts.length === 0 && !isFetching && (
        <div className="text-center py-16">
          <div className="w-16 h-16 mx-auto mb-4 border-2 border-[#dbdbdb] rounded-full flex items-center justify-center">
            <svg width="30" height="30" fill="none" viewBox="0 0 24 24" stroke="#8e8e8e" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"/><circle cx="12" cy="13" r="3"/>
            </svg>
          </div>
          <h2 className="text-[18px] font-semibold text-[#262626] mb-1">Ton fil est vide</h2>
          <p className="text-sm text-[#8e8e8e]">Suis des personnes pour voir leurs photos ici</p>
        </div>
      )}

      {posts.map((post) => <PostCard key={post.postID} post={post} />)}

      {isFetching && (
        <div className="flex justify-center py-5">
          <div className="w-7 h-7 border-2 border-[#dbdbdb] border-t-[#262626] rounded-full animate-spin" />
        </div>
      )}

      {hasMore && <div ref={ref} className="h-4" />}
      {!hasMore && posts.length > 0 && (
        <div className="text-center py-8">
          <p className="text-sm text-[#8e8e8e]">Tu es à jour ✓</p>
        </div>
      )}
    </div>
  );
}
