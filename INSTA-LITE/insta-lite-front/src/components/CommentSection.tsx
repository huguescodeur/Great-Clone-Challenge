import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getComments, createComment } from '../api/comments';
import CommentItem from './CommentItem';
import Avatar from './Avatar';
import { useAuthStore } from '../store/auth';

interface Props { postID: string; }

export default function CommentSection({ postID }: Props) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [content, setContent] = useState('');
  const [cursor, setCursor] = useState<string | undefined>();

  const { data, isLoading } = useQuery({
    queryKey: ['comments', postID, cursor],
    queryFn: () => getComments(postID, cursor),
  });

  const createMutation = useMutation({
    mutationFn: () => createComment(postID, content),
    onSuccess: () => { setContent(''); qc.invalidateQueries({ queryKey: ['comments', postID] }); },
  });

  return (
    <div className="border-t border-[#dbdbdb] px-3 pt-3 pb-1">
      {isLoading && <p className="text-xs text-[#8e8e8e] py-1">Chargement…</p>}

      {(data?.data ?? []).map((c) => (
        <CommentItem key={c.commentId} comment={c} postID={postID} />
      ))}

      {data?.hasMore && (
        <button onClick={() => setCursor(data.nextCursor)} className="text-sm text-[#8e8e8e] mb-2 cursor-pointer hover:text-[#262626]">
          Voir plus de commentaires
        </button>
      )}

      {user && (
        <div className="flex items-center gap-3 border-t border-[#dbdbdb] pt-3 mt-1">
          <Avatar name={user.fullName || user.username} url={user.profilePicURL} size="xs" />
          <input
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Ajouter un commentaire…"
            className="flex-1 text-sm text-[#262626] placeholder-[#8e8e8e] focus:outline-none"
            onKeyDown={(e) => e.key === 'Enter' && content.trim() && createMutation.mutate()}
          />
          {content.trim() && (
            <button
              onClick={() => createMutation.mutate()}
              disabled={createMutation.isPending}
              className="text-sm text-[#0095f6] font-semibold cursor-pointer hover:text-[#1877f2]"
            >
              Publier
            </button>
          )}
        </div>
      )}
    </div>
  );
}
