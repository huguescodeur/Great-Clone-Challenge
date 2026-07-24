import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';
import { getNotifications, markAsRead } from '../api/notifications';
import { formatDistanceToNow } from '../utils/date';
import { useAuthStore } from '../store/auth';
import { useNavigate } from 'react-router-dom';
import type { Notification } from '../types';

const notifConfig: Record<string, { label: string }> = {
  like_post:    { label: 'a aimé ta publication.' },
  comment_post: { label: 'a commenté ta publication :' },
  like_comment: { label: 'a aimé ton commentaire.' },
  new_follower: { label: 'a commencé à te suivre.' },
};

export default function NotificationsPage() {
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [notifs, setNotifs] = useState<Notification[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(true);
  const { ref, inView } = useInView({ threshold: 0.1 });

  useEffect(() => { if (!user) navigate('/login'); }, [user, navigate]);

  const { data, isFetching } = useQuery({
    queryKey: ['notifications', cursor],
    queryFn: () => getNotifications(cursor),
    enabled: !!user,
  });

  useEffect(() => {
    if (!data) return;
    if (!cursor) { setNotifs(data.data ?? []); }
    else { setNotifs((prev) => [...prev, ...(data.data ?? [])]); }
    setHasMore(data.hasMore);
  }, [data, cursor]);

  useEffect(() => {
    if (inView && hasMore && !isFetching && data?.nextCursor) setCursor(data.nextCursor);
  }, [inView, hasMore, isFetching, data]);

  const readMutation = useMutation({
    mutationFn: (id: string) => markAsRead(id),
    onSuccess: (_, id) => {
      setNotifs((prev) => prev.map((n) => n.notificationId === id ? { ...n, read: true } : n));
      qc.invalidateQueries({ queryKey: ['notifications-count'] });
    },
  });

  if (!user) return null;

  const grouped = groupByDate(notifs);

  return (
    <div className="bg-white border border-[#dbdbdb]">
      <div className="px-4 py-4 border-b border-[#dbdbdb]">
        <h1 className="text-base font-semibold text-[#262626]">Notifications</h1>
      </div>

      {notifs.length === 0 && !isFetching && (
        <div className="text-center py-16">
          <p className="text-sm text-[#8e8e8e]">Aucune notification</p>
        </div>
      )}

      {grouped.map(({ label, items }) => (
        <div key={label}>
          <p className="px-4 py-2 text-sm font-semibold text-[#262626] bg-[#fafafa]">{label}</p>
          {items.map((n) => {
            const cfg = notifConfig[n.type] ?? { label: n.type };
            return (
              <div
                key={n.notificationId}
                onClick={() => !n.read && readMutation.mutate(n.notificationId)}
                className={`flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-[#fafafa] transition-colors border-b border-[#f0f0f0] ${!n.read ? 'bg-[#eff7ff]' : ''}`}
              >
                <div className="w-11 h-11 rounded-full flex-shrink-0 flex items-center justify-center"
                  style={{ background: 'linear-gradient(45deg,#f09433,#e6683c,#dc2743,#cc2366,#bc1888)' }}>
                  <span className="text-white text-xs font-bold">{n.actorUsername.slice(0, 2).toUpperCase()}</span>
                </div>
                <p className="flex-1 text-sm text-[#262626] leading-snug">
                  <span className="font-semibold">{n.actorUsername}</span>
                  {' '}{cfg.label}
                  {' '}
                  <span className="text-[#8e8e8e] font-normal">{formatDistanceToNow(n.createdAt)}</span>
                </p>
                {!n.read && <div className="w-2 h-2 bg-[#0095f6] rounded-full flex-shrink-0" />}
              </div>
            );
          })}
        </div>
      ))}

      {isFetching && (
        <div className="flex justify-center py-5">
          <div className="w-7 h-7 border-2 border-[#dbdbdb] border-t-[#262626] rounded-full animate-spin" />
        </div>
      )}

      {hasMore && <div ref={ref} className="h-4" />}
    </div>
  );
}

function groupByDate(notifs: Notification[]) {
  const groups: { label: string; items: Notification[] }[] = [];
  const now = Date.now();
  const day = 86400000;

  for (const n of notifs) {
    const diff = now - new Date(n.createdAt).getTime();
    const label = diff < day ? "Aujourd'hui" : diff < 7 * day ? 'Cette semaine' : 'Plus ancien';
    const g = groups.find((x) => x.label === label);
    if (g) g.items.push(n);
    else groups.push({ label, items: [n] });
  }
  return groups;
}
