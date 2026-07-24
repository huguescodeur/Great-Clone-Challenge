import { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import { logout } from '../api/auth';
import { getNotifications } from '../api/notifications';
import { useAuthStore } from '../store/auth';
import Avatar from './Avatar';
import CreatePostModal from './CreatePostModal';

export default function Navbar() {
  const { user, clearAuth } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [showCreate, setShowCreate] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  const { data: notifData } = useQuery({
    queryKey: ['notifications-count'],
    queryFn: () => getNotifications(undefined, 1),
    enabled: !!user,
    refetchInterval: 30000,
  });

  const logoutMutation = useMutation({
    mutationFn: logout,
    onSettled: () => { clearAuth(); navigate('/login'); },
  });

  const unread = notifData?.unreadCount ?? 0;
  const at = (p: string) => location.pathname === p;

  return (
    <>
      <nav className="fixed top-0 left-0 right-0 z-50 bg-white border-b border-[#dbdbdb]">
        <div className="max-w-[975px] mx-auto px-4 h-[60px] flex items-center justify-between">

          {/* Logo Instagram-style */}
          <Link to="/feed" className="text-[22px] font-semibold" style={{ fontFamily: "'Billabong', cursive, serif", letterSpacing: '-0.5px' }}>
            InstaLite
          </Link>

          {/* Search placeholder (style Instagram) */}
          <div className="hidden sm:flex items-center bg-[#efefef] rounded-lg px-3 py-1.5 w-64">
            <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#8e8e8e" strokeWidth={2} className="flex-shrink-0">
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <span className="ml-2 text-sm text-[#8e8e8e]">Rechercher</span>
          </div>

          {/* Right icons */}
          <div className="flex items-center gap-1">
            <Link to="/feed" title="Accueil" className={`p-2 rounded-lg transition-colors ${at('/feed') ? 'text-black' : 'text-[#8e8e8e] hover:text-black'}`}>
              {at('/feed') ? (
                <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24"><path d="M9.005 16.545a2.997 2.997 0 012.997-2.997A2.997 2.997 0 0115 16.545V22h7V11.543L12 2 2 11.543V22h7.005z"/></svg>
              ) : (
                <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>
              )}
            </Link>

            {user && (
              <button onClick={() => setShowCreate(true)} title="Créer" className="p-2 text-[#8e8e8e] hover:text-black transition-colors cursor-pointer">
                <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
                  <rect x="3" y="3" width="18" height="18" rx="3"/>
                  <path strokeLinecap="round" d="M12 8v8M8 12h8"/>
                </svg>
              </button>
            )}

            <Link to="/explore" title="Explorer" className={`p-2 transition-colors ${at('/explore') ? 'text-black' : 'text-[#8e8e8e] hover:text-black'}`}>
              {at('/explore') ? (
                <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.486 2 2 6.486 2 12s4.486 10 10 10 10-4.486 10-10S17.514 2 12 2zm0 18c-4.411 0-8-3.589-8-8s3.589-8 8-8 8 3.589 8 8-3.589 8-8 8z"/><path d="M9.535 9.535L7.5 16.5l6.965-2.035 2.035-6.965-6.965 2.035zM12 13.5a1.5 1.5 0 110-3 1.5 1.5 0 010 3z"/></svg>
              ) : (
                <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>
              )}
            </Link>

            {user && (
              <Link to="/notifications" title="Notifications" className={`relative p-2 transition-colors ${at('/notifications') ? 'text-black' : 'text-[#8e8e8e] hover:text-black'}`}>
                {at('/notifications') ? (
                  <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24"><path d="M16.792 3.904A4.989 4.989 0 0121.5 9.122c0 3.072-2.652 4.959-5.197 7.222-2.512 2.243-3.865 3.469-4.303 3.752-.477-.309-2.143-1.823-4.303-3.752C5.141 14.072 2.5 12.167 2.5 9.122a4.989 4.989 0 014.708-5.218 4.21 4.21 0 013.675 1.941c.84 1.175.98 1.763 1.12 1.763s.278-.588 1.11-1.766a4.17 4.17 0 013.679-1.938m0-2a6.04 6.04 0 00-4.797 2.127 6.052 6.052 0 00-4.787-2.127A6.985 6.985 0 00.5 9.122c0 3.61 2.55 5.827 5.015 7.97.283.246.569.494.853.747l1.027.918a44.998 44.998 0 003.518 3.018 2 2 0 002.174 0 45.263 45.263 0 003.626-3.115l.922-.824c.293-.26.59-.519.885-.774 2.334-2.025 4.98-4.32 4.98-7.94a6.985 6.985 0 00-6.708-7.218z"/></svg>
                ) : (
                  <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
                )}
                {unread > 0 && <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full" />}
              </Link>
            )}

            {user ? (
              <div className="relative">
                <button onClick={() => setShowMenu((v) => !v)} className="p-1 cursor-pointer">
                  <div className={`rounded-full overflow-hidden ${at(`/profile/${user.userID}`) ? 'ring-2 ring-black ring-offset-1' : ''}`}>
                    <Avatar name={user.fullName || user.username} url={user.profilePicURL} size="xs" />
                  </div>
                </button>
                {showMenu && (
                  <div className="absolute right-0 top-10 bg-white border border-[#dbdbdb] rounded-lg shadow-lg z-50 w-52 py-1 overflow-hidden">
                    <Link to={`/profile/${user.userID}`} className="flex items-center gap-3 px-4 py-3 text-sm text-[#262626] hover:bg-[#fafafa]" onClick={() => setShowMenu(false)}>
                      <Avatar name={user.fullName || user.username} url={user.profilePicURL} size="xs" />
                      <span className="font-semibold">{user.username}</span>
                    </Link>
                    <div className="border-t border-[#dbdbdb] my-1" />
                    <button
                      onClick={() => { logoutMutation.mutate(); setShowMenu(false); }}
                      className="w-full text-left px-4 py-2.5 text-sm text-[#262626] hover:bg-[#fafafa] cursor-pointer"
                    >
                      Déconnexion
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <Link to="/login" className="ml-1 px-4 py-1.5 bg-[#0095f6] text-white rounded-lg text-sm font-semibold hover:bg-[#1877f2] transition-colors">
                Connexion
              </Link>
            )}
          </div>
        </div>
      </nav>

      {showCreate && <CreatePostModal onClose={() => setShowCreate(false)} />}
    </>
  );
}
