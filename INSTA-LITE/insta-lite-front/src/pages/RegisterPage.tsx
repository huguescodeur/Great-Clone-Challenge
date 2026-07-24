import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { register } from '../api/auth';
import { useAuthStore } from '../store/auth';

export default function RegisterPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [form, setForm] = useState({ username: '', email: '', fullName: '', password: '' });
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: () => register(form),
    onMutate: () => setError(''),
    onSuccess: (data) => { setAuth(data.user, data.token); navigate('/feed'); },
    onError: (e: Error) => setError(e.message || 'Une erreur est survenue'),
  });

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((f) => ({ ...f, [k]: e.target.value }));
    setError('');
  };

  return (
    <div className="min-h-screen bg-[#fafafa] flex items-center justify-center p-4">
      <div className="w-full max-w-[350px]">

        <div className="bg-white border border-[#dbdbdb] px-10 py-10">
          <h1 className="text-[36px] text-center mb-2" style={{ fontFamily: "'Billabong', cursive, serif" }}>
            InstaLite
          </h1>
          <p className="text-[#8e8e8e] text-sm text-center font-semibold mb-5">
            Inscris-toi pour voir les photos et vidéos de tes amis.
          </p>

          <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="space-y-2">
            {[
              { k: 'email' as const, t: 'email', p: 'Adresse e-mail' },
              { k: 'fullName' as const, t: 'text', p: 'Nom complet' },
              { k: 'username' as const, t: 'text', p: "Nom d'utilisateur" },
              { k: 'password' as const, t: 'password', p: 'Mot de passe' },
            ].map(({ k, t, p }) => (
              <input
                key={k}
                type={t}
                placeholder={p}
                value={form[k]}
                onChange={set(k)}
                className="w-full px-2.5 py-2 bg-[#fafafa] border border-[#dbdbdb] rounded-[3px] text-xs text-[#262626] placeholder-[#8e8e8e] focus:outline-none focus:border-[#a8a8a8]"
                required
              />
            ))}

            {error && (
              <p className="text-xs text-[#ed4956] text-center py-1">
                {error}
              </p>
            )}

            <p className="text-[11px] text-[#8e8e8e] text-center mt-2">
              En t'inscrivant, tu acceptes nos Conditions générales.
            </p>

            <button
              type="submit"
              disabled={mutation.isPending || !form.email || !form.username || !form.password}
              className="w-full py-1.5 bg-[#0095f6] text-white rounded-lg text-sm font-semibold hover:bg-[#1877f2] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer transition-colors"
            >
              {mutation.isPending ? 'Inscription…' : "S'inscrire"}
            </button>
          </form>
        </div>

        <div className="bg-white border border-[#dbdbdb] mt-2.5 py-4 text-center">
          <p className="text-sm text-[#262626]">
            Tu as déjà un compte ?{' '}
            <Link to="/login" className="text-[#0095f6] font-semibold hover:text-[#1877f2]">
              Connecte-toi
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
