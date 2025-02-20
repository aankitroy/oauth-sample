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

    // Call backend to exchange the code.
    fetch('http://localhost:8000/api/admin/v1/token-exchange', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ code: code, codeVerifier: codeVerifier }),
    })
      .then((res) => {
        if (!res.ok) throw new Error('Token exchange failed');
        return res.json();
      })
      .then((res) => {
        // Once exchange is successful, go to protected page
        router.push('/protected');
      })
      .catch((err) => console.error(err));
  }, [searchParams, router]);

  return <div className="p-4">Authenticating...</div>;
}
