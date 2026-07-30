import { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { createFollow, deleteFollow, getFollowStatus } from '../api/follow';
import { useAuthStore } from '../store/auth';

interface Props { targetUserID: string; targetUsername?: string; }

export default function FollowButton({ targetUserID, targetUsername }: Props) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [following, setFollowing] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user || user.userID === targetUserID) { setLoading(false); return; }
    getFollowStatus(targetUserID)
      .then(() => setFollowing(true))
      .catch(() => setFollowing(false))
      .finally(() => setLoading(false));
  }, [targetUserID, user]);

  if (!user || user.userID === targetUserID) return null;

  const toggle = async () => {
    setLoading(true);
    try {
      if (following) { await deleteFollow(targetUserID); setFollowing(false); }
      else { await createFollow(targetUserID); setFollowing(true); }
      qc.invalidateQueries({ queryKey: ['user', targetUsername ?? targetUserID] });
      if (user) qc.invalidateQueries({ queryKey: ['user', user.username] });
    } finally { setLoading(false); }
  };

  if (loading) return <div className="w-16 h-7 bg-[#efefef] rounded-lg animate-pulse" />;

  return (
    <button
      onClick={toggle}
      className={`px-4 py-1.5 rounded-lg text-sm font-semibold transition-colors cursor-pointer ${
        following
          ? 'bg-[#efefef] text-[#262626] hover:bg-[#e0e0e0]'
          : 'bg-[#0095f6] text-white hover:bg-[#1877f2]'
      }`}
    >
      {following ? 'Abonné(e)' : 'Suivre'}
    </button>
  );
}
