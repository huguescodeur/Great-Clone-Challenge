import { useState, useRef, useEffect } from 'react';
import type { ReactionType } from '../types';

interface Props {
  current: ReactionType | null;
  count: number;
  onReact: (type: ReactionType) => void;
  onUnreact: () => void;
}

const reactions: { type: ReactionType; emoji: string }[] = [
  { type: 'like', emoji: '❤️' },
  { type: 'love', emoji: '😍' },
  { type: 'laugh', emoji: '😂' },
];

export default function ReactionPicker({ current, count, onReact, onUnreact }: Props) {
  const [show, setShow] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!show) return;
    const fn = (e: MouseEvent) => { if (!ref.current?.contains(e.target as Node)) setShow(false); };
    document.addEventListener('mousedown', fn);
    return () => document.removeEventListener('mousedown', fn);
  }, [show]);

  const toggle = () => {
    if (current) setShow((v) => !v);
    else onReact('like');
  };

  const pick = (type: ReactionType) => {
    setShow(false);
    onReact(type);
  };

  return (
    <div ref={ref} className="relative flex items-center gap-1.5">
      {show && (
        <div className="absolute bottom-9 left-0 bg-white border border-[#dbdbdb] rounded-2xl shadow-lg flex items-center gap-1 px-2 py-1.5 z-10">
          {reactions.map((r) => (
            <button key={r.type} onClick={() => pick(r.type)} className="text-xl hover:scale-125 transition-transform p-1 cursor-pointer">
              {r.emoji}
            </button>
          ))}
          <button
            onClick={() => { setShow(false); onUnreact(); }}
            className="ml-1 w-6 h-6 flex items-center justify-center rounded-full bg-[#efefef] text-[#8e8e8e] hover:bg-[#dbdbdb] text-xs font-bold cursor-pointer"
            title="Retirer la réaction"
          >
            ✕
          </button>
        </div>
      )}

      <button
        onClick={toggle}
        onContextMenu={(e) => { e.preventDefault(); setShow((v) => !v); }}
        className="flex items-center gap-0 cursor-pointer"
        title={current ? 'Changer la réaction' : 'Aimer'}
      >
        {current === 'like' ? (
          <svg width="24" height="24" fill="#ed4956" viewBox="0 0 24 24"><path d="M16.792 3.904A4.989 4.989 0 0121.5 9.122c0 3.072-2.652 4.959-5.197 7.222-2.512 2.243-3.865 3.469-4.303 3.752-.477-.309-2.143-1.823-4.303-3.752C5.141 14.072 2.5 12.167 2.5 9.122a4.989 4.989 0 014.708-5.218 4.21 4.21 0 013.675 1.941c.84 1.175.98 1.763 1.12 1.763s.278-.588 1.11-1.766a4.17 4.17 0 013.679-1.938m0-2a6.04 6.04 0 00-4.797 2.127 6.052 6.052 0 00-4.787-2.127A6.985 6.985 0 00.5 9.122c0 3.61 2.55 5.827 5.015 7.97.283.246.569.494.853.747l1.027.918a44.998 44.998 0 003.518 3.018 2 2 0 002.174 0 45.263 45.263 0 003.626-3.115l.922-.824c.293-.26.59-.519.885-.774 2.334-2.025 4.98-4.32 4.98-7.94a6.985 6.985 0 00-6.708-7.218z"/></svg>
        ) : current === 'love' ? (
          <span className="text-2xl leading-none" style={{ lineHeight: '24px' }}>😍</span>
        ) : current === 'laugh' ? (
          <span className="text-2xl leading-none" style={{ lineHeight: '24px' }}>😂</span>
        ) : (
          <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8} className="text-[#262626]"><path strokeLinecap="round" strokeLinejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
        )}
      </button>

      {!current && (
        <button onClick={() => setShow((v) => !v)} className="text-[11px] text-[#8e8e8e] cursor-pointer hover:text-[#262626]">▾</button>
      )}

      {count > 0 && (
        <span className="text-sm font-semibold text-[#262626]">{count}</span>
      )}
    </div>
  );
}
