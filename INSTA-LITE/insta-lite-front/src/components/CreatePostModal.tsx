import { useState, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createPost } from '../api/posts';
import { uploadFile } from '../api/upload';
import Avatar from './Avatar';
import { useAuthStore } from '../store/auth';

interface Props { onClose: () => void; }

interface MediaItem {
  file: File;
  preview: string;
  type: 'image' | 'video';
}

type UploadStep =
  | { kind: 'uploading'; index: number; total: number }
  | { kind: 'publishing' }
  | { kind: 'done' };

export default function CreatePostModal({ onClose }: Props) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [content, setContent] = useState('');
  const [medias, setMedias] = useState<MediaItem[]>([]);
  const [error, setError] = useState('');
  const [step, setStep] = useState<UploadStep | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const onFiles = (files: FileList | null) => {
    if (!files) return;
    const items: MediaItem[] = Array.from(files).map((f) => ({
      file: f,
      preview: URL.createObjectURL(f),
      type: f.type.startsWith('video') ? 'video' : 'image',
    }));
    setMedias((prev) => [...prev, ...items]);
  };

  const removeMedia = (i: number) => {
    URL.revokeObjectURL(medias[i].preview);
    setMedias((prev) => prev.filter((_, idx) => idx !== i));
  };

  const createMutation = useMutation({
    mutationFn: async () => {
      if (medias.length === 0) throw new Error('Ajoute au moins une photo ou vidéo');

      const uploaded: { url: string; type: 'image' | 'video' }[] = [];
      for (let i = 0; i < medias.length; i++) {
        setStep({ kind: 'uploading', index: i + 1, total: medias.length });
        const url = await uploadFile(medias[i].file);
        uploaded.push({ url, type: medias[i].type });
      }

      setStep({ kind: 'publishing' });
      const result = await createPost(content, uploaded);
      setStep({ kind: 'done' });
      return result;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['feed'] });
      qc.invalidateQueries({ queryKey: ['user-posts', user?.userID] });
      medias.forEach((m) => URL.revokeObjectURL(m.preview));
      setTimeout(onClose, 2000);
    },
    onError: (e: unknown) => {
      setStep(null);
      if (e instanceof Error) {
        const axiosMsg = (e as { response?: { status?: number } }).response?.status;
        setError(axiosMsg ? `Erreur serveur (${axiosMsg}) — vérifie que le serveur est démarré` : e.message);
      } else {
        setError('Une erreur inattendue est survenue');
      }
    },
  });

  const isPending = createMutation.isPending || step !== null;

  return (
    <div
      className="fixed inset-0 bg-black/65 z-50 flex items-center justify-center p-4"
      onClick={(e) => !isPending && e.target === e.currentTarget && onClose()}
    >
      <div className="bg-white rounded-xl w-full max-w-lg overflow-hidden shadow-2xl relative">

        {/* Progress overlay */}
        {step && (
          <div className="absolute inset-0 bg-white z-10 flex flex-col items-center justify-center gap-4 rounded-xl">
            {step.kind === 'done' ? (
              <>
                <div className="w-14 h-14 rounded-full border-2 border-[#0095f6] flex items-center justify-center">
                  <svg width="28" height="28" fill="none" viewBox="0 0 24 24" stroke="#0095f6" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <p className="text-sm font-semibold text-[#262626]">Publication partagée !</p>
              </>
            ) : (
              <>
                <div className="w-14 h-14 rounded-full border-2 border-[#dbdbdb] border-t-[#0095f6] animate-spin" />
                <div className="text-center">
                  {step.kind === 'uploading' ? (
                    <>
                      <p className="text-sm font-semibold text-[#262626]">
                        Envoi {step.index}/{step.total}…
                      </p>
                      <p className="text-xs text-[#8e8e8e] mt-1">Upload de la photo</p>
                    </>
                  ) : (
                    <>
                      <p className="text-sm font-semibold text-[#262626]">Publication en cours…</p>
                      <p className="text-xs text-[#8e8e8e] mt-1">Presque terminé</p>
                    </>
                  )}
                </div>
                {medias.length > 1 && step.kind === 'uploading' && (
                  <div className="flex gap-1.5 mt-1">
                    {medias.map((_, i) => (
                      <div
                        key={i}
                        className={`w-2 h-2 rounded-full transition-colors ${i < step.index ? 'bg-[#0095f6]' : 'bg-[#dbdbdb]'}`}
                      />
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        )}

        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-[#dbdbdb]">
          <button onClick={onClose} disabled={isPending} className="cursor-pointer text-[#262626] disabled:opacity-40">
            <svg width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <span className="text-sm font-semibold text-[#262626]">Créer une publication</span>
          <button
            onClick={() => createMutation.mutate()}
            disabled={!content.trim() || medias.length === 0 || isPending}
            className="text-sm font-semibold text-[#0095f6] disabled:text-[#b2dffc] cursor-pointer disabled:cursor-not-allowed"
          >
            Partager
          </button>
        </div>

        {/* Body */}
        <div className="flex" style={{ minHeight: 300 }}>
          {/* Left: media */}
          <div className="w-56 flex-shrink-0 bg-black flex flex-col">
            {medias.length === 0 ? (
              <div
                onClick={() => fileRef.current?.click()}
                onDrop={(e) => { e.preventDefault(); onFiles(e.dataTransfer.files); }}
                onDragOver={(e) => e.preventDefault()}
                className="flex-1 flex flex-col items-center justify-center gap-2 cursor-pointer"
              >
                <svg width="64" height="64" fill="none" viewBox="0 0 24 24" stroke="#fff" strokeWidth={0.8}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                </svg>
                <p className="text-white text-sm text-center px-4">Fais glisser tes photos et vidéos ici</p>
                <button className="mt-1 px-3 py-1.5 bg-[#0095f6] text-white rounded-lg text-sm font-semibold cursor-pointer hover:bg-[#1877f2]">
                  Sélectionner sur l'appareil
                </button>
              </div>
            ) : (
              <div className={`flex-1 grid ${medias.length > 1 ? 'grid-cols-2' : 'grid-cols-1'} gap-0.5`}>
                {medias.map((m, i) => (
                  <div key={i} className="relative bg-[#111]">
                    {m.type === 'video'
                      ? <video src={m.preview} className="w-full h-full object-cover" muted />
                      : <img src={m.preview} alt="" className="w-full h-full object-cover" />}
                    {!isPending && (
                      <button
                        onClick={() => removeMedia(i)}
                        className="absolute top-1 right-1 bg-black/60 text-white rounded-full w-5 h-5 flex items-center justify-center cursor-pointer text-xs font-bold"
                      >×</button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Right: caption */}
          <div className="flex-1 flex flex-col p-3">
            {user && (
              <div className="flex items-center gap-2 mb-3">
                <Avatar name={user.fullName || user.username} url={user.profilePicURL} size="xs" />
                <span className="text-sm font-semibold text-[#262626]">{user.username}</span>
              </div>
            )}
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Rédigez une légende…"
              maxLength={2200}
              rows={8}
              disabled={isPending}
              className="flex-1 resize-none text-sm text-[#262626] placeholder-[#8e8e8e] focus:outline-none disabled:opacity-60"
            />
            <div className="flex items-center justify-between mt-2 pt-2 border-t border-[#dbdbdb]">
              <span className="text-[12px] text-[#8e8e8e]">{content.length}/2 200</span>
              {medias.length > 0 && !isPending && (
                <button onClick={() => fileRef.current?.click()} className="text-[12px] text-[#0095f6] cursor-pointer font-semibold">
                  + Ajouter des médias
                </button>
              )}
            </div>

            {error && <p className="text-xs text-[#ed4956] mt-2">{error}</p>}
            {medias.length === 0 && (
              <p className="text-xs text-[#8e8e8e] mt-1">Une photo ou vidéo est requise</p>
            )}
          </div>
        </div>
      </div>

      <input ref={fileRef} type="file" accept="image/*,video/*" multiple className="hidden"
        onChange={(e) => onFiles(e.target.files)} />
    </div>
  );
}
