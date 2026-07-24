import { useState, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { formatDistanceToNow } from '../utils/date';
import Avatar from './Avatar';
import ReactionPicker from './ReactionPicker';
import CommentSection from './CommentSection';
import { getMyLike, createLike, updateLike, deleteLike } from '../api/likes';
import { deletePost, updatePost } from '../api/posts';
import { useAuthStore } from '../store/auth';
import type { PostResponse, ReactionType } from '../types';

interface Props { post: PostResponse; }

export default function PostCard({ post }: Props) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [showComments, setShowComments] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editContent, setEditContent] = useState(post.content);
  const [displayContent, setDisplayContent] = useState(post.content);
  const [likesCount, setLikesCount] = useState(post.likesCount);
  const [showMenu, setShowMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!showMenu) return;
    const fn = (e: MouseEvent) => { if (!menuRef.current?.contains(e.target as Node)) setShowMenu(false); };
    document.addEventListener('mousedown', fn);
    return () => document.removeEventListener('mousedown', fn);
  }, [showMenu]);

  const { data: myLike } = useQuery({
    queryKey: ['like-me', post.postID],
    queryFn: () => getMyLike(post.postID),
    retry: false,
    enabled: !!user,
    staleTime: Infinity,
  });

  const [currentReaction, setCurrentReaction] = useState<ReactionType | null>(null);

  useEffect(() => {
    if (myLike !== undefined) setCurrentReaction(myLike.reactionType);
  }, [myLike]);

  const reactMutation = useMutation({
    mutationFn: ({ type, isUpdate }: { type: ReactionType; isUpdate: boolean }) =>
      isUpdate ? updateLike(post.postID, type) : createLike(post.postID, type),
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
    mutationFn: () => deleteLike(post.postID),
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

  const editMutation = useMutation({
    mutationFn: () => updatePost(post.postID, editContent),
    onSuccess: () => {
      setDisplayContent(editContent);
      qc.invalidateQueries({ queryKey: ['feed'] });
      qc.invalidateQueries({ queryKey: ['user-posts'] });
      setEditing(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deletePost(post.postID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['feed'] });
      qc.invalidateQueries({ queryKey: ['user-posts'] });
    },
  });

  const isOwner = user?.userID === post.userID;
  const userName = post.username;

  return (
    <article className="bg-white border border-[#dbdbdb] mb-4">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2.5">
        <Link to={`/profile/${post.userID}`} className="flex items-center gap-2.5 hover:opacity-80 transition-opacity">
          <Avatar name={userName} size="sm" ring />
          <span className="text-sm font-semibold text-[#262626]">{userName}</span>
        </Link>

        {isOwner && (
          <div ref={menuRef} className="relative">
            <button onClick={() => setShowMenu((v) => !v)} className="p-1 text-[#262626] cursor-pointer">
              <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
                <circle cx="12" cy="12" r="1.5"/><circle cx="6" cy="12" r="1.5"/><circle cx="18" cy="12" r="1.5"/>
              </svg>
            </button>
            {showMenu && (
              <div className="absolute right-0 top-8 bg-white border border-[#dbdbdb] rounded-lg shadow-lg z-20 w-36 overflow-hidden">
                <button onClick={() => { setEditing(true); setShowMenu(false); }} className="w-full text-left px-4 py-2.5 text-sm text-[#262626] hover:bg-[#fafafa] cursor-pointer">
                  Modifier
                </button>
                <button onClick={() => { deleteMutation.mutate(); setShowMenu(false); }} className="w-full text-left px-4 py-2.5 text-sm text-[#ed4956] hover:bg-[#fafafa] cursor-pointer">
                  Supprimer
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Media */}
      {post.media && post.media.length > 0 && (
        <div className={`grid gap-0.5 ${post.media.length > 1 ? 'grid-cols-2' : 'grid-cols-1'}`}>
          {post.media.map((m) =>
            m.type === 'video' ? (
              <video key={m.mediaID} src={m.url} controls className="w-full max-h-[585px] object-cover" />
            ) : (
              <img key={m.mediaID} src={m.url} alt="" className="w-full max-h-[585px] object-cover bg-[#efefef]"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
            )
          )}
        </div>
      )}

      {/* Actions */}
      <div className="px-3 pt-2.5 pb-1">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-3">
            <ReactionPicker
              current={currentReaction}
              count={0}
              onReact={(t) => reactMutation.mutate({ type: t, isUpdate: currentReaction !== null })}
              onUnreact={() => unreactMutation.mutate()}
            />
            <button onClick={() => setShowComments((v) => !v)} className="cursor-pointer text-[#262626] hover:text-[#8e8e8e] transition-colors">
              <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
              </svg>
            </button>
          </div>
        </div>

        {/* Likes count */}
        {likesCount > 0 && (
          <p className="text-sm font-semibold text-[#262626] mb-1.5">{likesCount.toLocaleString('fr')} J'aime</p>
        )}

        {/* Caption */}
        {editing ? (
          <div className="mb-2">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              rows={3}
              className="w-full text-sm border border-[#dbdbdb] rounded px-2 py-1.5 focus:outline-none resize-none"
            />
            <div className="flex gap-3 mt-1">
              <button onClick={() => editMutation.mutate()} disabled={editMutation.isPending} className="text-sm text-[#0095f6] font-semibold cursor-pointer">Sauvegarder</button>
              <button onClick={() => setEditing(false)} className="text-sm text-[#8e8e8e] cursor-pointer">Annuler</button>
            </div>
          </div>
        ) : (
          <p className="text-sm text-[#262626] mb-1.5">
            <Link to={`/profile/${post.userID}`} className="font-semibold mr-1.5 hover:opacity-70">{userName}</Link>
            {displayContent}
          </p>
        )}

        {/* Comments count */}
        {post.commentsCount > 0 && (
          <button onClick={() => setShowComments((v) => !v)} className="text-sm text-[#8e8e8e] cursor-pointer hover:text-[#262626]">
            {showComments
              ? 'Masquer les commentaires'
              : `Voir ${post.commentsCount > 1 ? `les ${post.commentsCount} commentaires` : '1 commentaire'}`}
          </button>
        )}

        {/* Timestamp */}
        <p className="text-[10px] uppercase tracking-wider text-[#8e8e8e] mt-1.5">
          {formatDistanceToNow(post.createdAt)}
        </p>
      </div>

      {showComments && <CommentSection postID={post.postID} />}
    </article>
  );
}
