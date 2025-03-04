"use client";

import { useEffect, useState } from 'react';

export default function ProtectedPage() {
  const [message, setMessage] = useState('');

  useEffect(() => {
    const baseUrl = process.env.NEXT_PUBLIC_BASE_URL;
    fetch(`${baseUrl}/api/v1/protected`, {
      method: 'POST',
      credentials: 'include',
    })
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
      })
      .then((data) => setMessage(data.message))
      .catch(() => setMessage('Unauthorized or error'));
  }, []);

  return (
    <div className="p-4">
      <h1>Protected Content</h1>
      <p>{message}</p>
    </div>
  );
}
