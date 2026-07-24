import client from './client';
import type { SignedURLResponse } from '../types';

export const getSignedURL = () =>
  client.get<SignedURLResponse>('/upload/signed-url').then((r) => r.data);

export async function uploadFile(file: File): Promise<string> {
  const { signature, timestamp, apiKey, uploadUrl, folder } = await getSignedURL();

  const formData = new FormData();
  formData.append('file', file);
  formData.append('api_key', apiKey);
  formData.append('timestamp', String(timestamp));
  formData.append('signature', signature);
  formData.append('folder', folder);

  const res = await fetch(uploadUrl, { method: 'POST', body: formData });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error?.message ?? `Upload échoué (${res.status})`);
  }
  const data = await res.json();
  return data.secure_url as string;
}
