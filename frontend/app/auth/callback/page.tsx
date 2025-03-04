"use client";

import { useSearchParams, useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function CallbackPage() {
  const searchParams = useSearchParams();
  const router = useRouter();

  useEffect(() => {
    const code = searchParams.get('code');
    if (!code) return;

    const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
    if (!codeVerifier) return;
    const baseUrl = process.env.NEXT_PUBLIC_BASE_URL;
    const redirectUri = process.env.NEXT_PUBLIC_REDIRECT_URI;

    // Call backend to exchange the code.
    fetch(`${baseUrl}/api/v1/token-exchange`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: code, codeVerifier: codeVerifier, redirectUri: redirectUri }),
    })
      .then((res) => {
        if (!res.ok) throw new Error('Token exchange failed');
        return;
      })
      .then((res) => {
        // Once exchange is successful, go to protected page
        router.push('/protected');
      })
      .catch((err) => console.error(err));
  }, [searchParams, router]);

  return <div className="p-4">Authenticating...</div>;
}
