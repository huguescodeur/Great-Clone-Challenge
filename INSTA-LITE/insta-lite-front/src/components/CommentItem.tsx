import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Avatar from './Avatar';
import ReactionPicker from './ReactionPicker';
import { getReplies, createComment, updateComment, deleteComment } from '../api/comments';
import { getMyCommentLike, createCommentLike, updateCommentLike, deleteCommentLike } from '../api/commentLikes';
import { useAuthStore } from '../store/auth';
import { formatDistanceToNow } from '../utils/date';
import type { Comment, ReactionType } from '../types';

interface Props { comment: Comment; postID: string; depth?: number; }

export default function CommentItem({ comment, postID, depth = 0 }: Props) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [showReplies, setShowReplies] = useState(false);
  const [replyContent, setReplyContent] = useState('');
  const [replying, setReplying] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [likesCount, setLikesCount] = useState(comment.likesCount);

  const { data: myLike } = useQuery({
    queryKey: ['comment-like-me', comment.commentId],
    queryFn: () => getMyCommentLike(comment.commentId),
    retry: false,
    enabled: !!user,
    staleTime: Infinity,
  });

  const [currentReaction, setCurrentReaction] = useState<ReactionType | null>(null);

  useEffect(() => {
    if (myLike !== undefined) setCurrentReaction(myLike.reactionType);
  }, [myLike]);

  const { data: replies } = useQuery({
    queryKey: ['replies', comment.commentId],
    queryFn: () => getReplies(comment.commentId),
    enabled: showReplies,
  });

  const replyMutation = useMutation({
    mutationFn: () => createComment(postID, replyContent, comment.commentId),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['replies', comment.commentId] }); qc.invalidateQueries({ queryKey: ['comments', postID] }); setReplyContent(''); setReplying(false); setShowReplies(true); },
  });

  const editMutation = useMutation({
    mutationFn: () => updateComment(comment.commentId, editContent),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['comments', postID] }); setEditing(false); },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteComment(comment.commentId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['comments', postID] }),
  });

  const reactMutation = useMutation({
    mutationFn: ({ type, isUpdate }: { type: ReactionType; isUpdate: boolean }) =>
      isUpdate ? updateCommentLike(comment.commentId, type) : createCommentLike(comment.commentId, type),
    onMutate: ({ type, isUpdate }) => {
      const prev = currentReaction;
      setCurrentReaction(type);
      if (!isUpdate) setLikesCount((c) => c + 1);
      return { prev, isUpdate };
    },
    onError: (_, __, ctx) => {
      setCurrentReaction(ctx?.prev ?? null);
      if (!ctx?.isUpdate) setLikesCount((c) => Math.max(0, c - 1));
    },
  });

  const unreactMutation = useMutation({
    mutationFn: () => deleteCommentLike(comment.commentId),
    onMutate: () => {
      const prev = currentReaction;
      setCurrentReaction(null);
      setLikesCount((c) => Math.max(0, c - 1));
      return { prev };
    },
    onError: (_, __, ctx) => {
      setCurrentReaction(ctx?.prev ?? null);
      setLikesCount((c) => c + 1);
    },
  });

  const isOwner = user?.userID === comment.userId;
  const name = comment.username;

  return (
    <div className={`flex gap-2.5 ${depth > 0 ? 'ml-10 mt-2' : 'mb-3'}`}>
      <Avatar name={name} size="xs" />
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            {editing ? (
              <div>
                <input
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  className="w-full text-sm border border-[#dbdbdb] rounded px-2 py-1 focus:outline-none"
                  onKeyDown={(e) => e.key === 'Enter' && editMutation.mutate()}
                />
                <div className="flex gap-2 mt-1">
                  <button onClick={() => editMutation.mutate()} className="text-xs text-[#0095f6] font-semibold cursor-pointer">Sauvegarder</button>
                  <button onClick={() => setEditing(false)} className="text-xs text-[#8e8e8e] cursor-pointer">Annuler</button>
                </div>
              </div>
            ) : (
              <p className="text-sm text-[#262626]">
                <span className="font-semibold mr-1.5">{name}</span>
                {comment.content}
              </p>
            )}
            <div className="flex items-center gap-3 mt-0.5">
              <span className="text-[10px] text-[#8e8e8e]">{formatDistanceToNow(comment.createdAt)}</span>
              {likesCount > 0 && <span className="text-[10px] font-semibold text-[#8e8e8e]">{likesCount} J'aime</span>}
              {depth === 0 && (
                <button onClick={() => setReplying((v) => !v)} className="text-[10px] font-semibold text-[#8e8e8e] cursor-pointer hover:text-[#262626]">Répondre</button>
              )}
              {isOwner && (
                <>
                  <button onClick={() => setEditing(true)} className="text-[10px] text-[#8e8e8e] cursor-pointer hover:text-[#262626]">Modifier</button>
                  <button onClick={() => deleteMutation.mutate()} className="text-[10px] text-[#8e8e8e] cursor-pointer hover:text-[#ed4956]">Supprimer</button>
                </>
              )}
            </div>
          </div>
          <div className="mt-1">
            <ReactionPicker current={currentReaction} count={0} onReact={(t) => reactMutation.mutate({ type: t, isUpdate: currentReaction !== null })} onUnreact={() => unreactMutation.mutate()} />
          </div>
        </div>

        {replying && (
          <div className="flex gap-2 mt-2 items-center">
            <input
              value={replyContent}
              onChange={(e) => setReplyContent(e.target.value)}
              placeholder={`Répondre à ${name}…`}
              className="flex-1 text-sm text-[#262626] placeholder-[#8e8e8e] focus:outline-none"
              onKeyDown={(e) => e.key === 'Enter' && replyContent.trim() && replyMutation.mutate()}
            />
            {replyContent.trim() && (
              <button onClick={() => replyMutation.mutate()} className="text-sm text-[#0095f6] font-semibold cursor-pointer">Publier</button>
            )}
          </div>
        )}

        {depth === 0 && (
          <button onClick={() => setShowReplies((v) => !v)} className="flex items-center gap-1.5 mt-1.5 cursor-pointer">
            <span className="inline-block w-5 h-px bg-[#8e8e8e]" />
            <span className="text-[11px] font-semibold text-[#8e8e8e] hover:text-[#262626]">
              {showReplies ? 'Masquer les réponses' : replies ? `Voir ${replies.length} réponse${replies.length > 1 ? 's' : ''}` : 'Voir les réponses'}
            </span>
          </button>
        )}

        {showReplies && replies?.map((r) => (
          <CommentItem key={r.commentId} comment={r} postID={postID} depth={depth + 1} />
        ))}
      </div>
    </div>
  );
}
