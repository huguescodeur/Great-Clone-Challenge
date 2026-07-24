import { useState, type ChangeEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { login } from '../api/auth';
import { useAuthStore } from '../store/auth';

export default function LoginPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [form, setForm] = useState({ identifier: '', password: '' });
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: () => login(form),
    onMutate: () => setError(''),
    onSuccess: (data) => { setAuth(data.user, data.token); navigate('/feed'); },
    onError: () => setError('Les informations que tu as saisies ne correspondent à aucun compte.'),
  });

  const setField = (k: keyof typeof form) => (e: ChangeEvent<HTMLInputElement>) => {
    setForm((f) => ({ ...f, [k]: e.target.value }));
    setError('');
  };

  return (
    <div className="min-h-screen bg-[#fafafa] flex items-center justify-center p-4">
      <div className="w-full max-w-[350px]">

        {/* Card */}
        <div className="bg-white border border-[#dbdbdb] px-10 py-10">
          <h1 className="text-[36px] text-center mb-6" style={{ fontFamily: "'Billabong', cursive, serif" }}>
            InstaLite
          </h1>

          <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="space-y-2">
            <input
              type="text"
              placeholder="Numéro de téléphone, nom d'utilisateur ou email"
              value={form.identifier}
              onChange={setField('identifier')}
              className="w-full px-2.5 py-2 bg-[#fafafa] border border-[#dbdbdb] rounded-[3px] text-xs text-[#262626] placeholder-[#8e8e8e] focus:outline-none focus:border-[#a8a8a8]"
              required
            />
            <input
              type="password"
              placeholder="Mot de passe"
              value={form.password}
              onChange={setField('password')}
              className="w-full px-2.5 py-2 bg-[#fafafa] border border-[#dbdbdb] rounded-[3px] text-xs text-[#262626] placeholder-[#8e8e8e] focus:outline-none focus:border-[#a8a8a8]"
              required
            />

            {error && (
              <p className="text-xs text-[#ed4956] text-center py-1">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={mutation.isPending || !form.identifier || !form.password}
              className="w-full py-1.5 mt-2 bg-[#0095f6] text-white rounded-lg text-sm font-semibold hover:bg-[#1877f2] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer transition-colors"
            >
              {mutation.isPending ? 'Connexion…' : 'Se connecter'}
            </button>
          </form>

          <div className="flex items-center gap-3 mt-4">
            <div className="flex-1 h-px bg-[#dbdbdb]" />
            <span className="text-xs font-semibold text-[#8e8e8e]">OU</span>
            <div className="flex-1 h-px bg-[#dbdbdb]" />
          </div>
        </div>

        {/* Register card */}
        <div className="bg-white border border-[#dbdbdb] mt-2.5 py-4 text-center">
          <p className="text-sm text-[#262626]">
            Tu n'as pas de compte ?{' '}
            <Link to="/register" className="text-[#0095f6] font-semibold hover:text-[#1877f2]">
              Inscris-toi
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
