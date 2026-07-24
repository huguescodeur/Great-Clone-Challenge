import { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';
import { getAllPosts } from '../api/posts';
import PostCard from '../components/PostCard';
import type { PostResponse } from '../types';

export default function ExplorePage() {
  const [posts, setPosts] = useState<PostResponse[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(true);
  const initialized = useRef(false);
  const { ref, inView } = useInView({ threshold: 0.1 });

  const { data, isFetching } = useQuery({
    queryKey: ['posts', cursor],
    queryFn: () => getAllPosts(cursor),
  });

  useEffect(() => {
    if (!data) return;
    if (!initialized.current) { setPosts(data.data ?? []); initialized.current = true; }
    else if (cursor) { setPosts((prev) => [...prev, ...(data.data ?? [])]); }
    setHasMore(data.has_more);
  }, [data, cursor]);

  useEffect(() => {
    if (inView && hasMore && !isFetching && data?.next_cursor) setCursor(data.next_cursor);
  }, [inView, hasMore, isFetching, data]);

  return (
    <div>
      {posts.length === 0 && !isFetching && (
        <div className="text-center py-16">
          <p className="text-sm text-[#8e8e8e]">Aucune publication à explorer</p>
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
