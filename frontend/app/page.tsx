"use client"; // We need client-side code to do the PKCE generation & redirect

// Assuming the import error is due to a missing or incorrect path
// If the file exists, ensure it's in the correct location and the path is correct
// If the file doesn't exist, create it or replace the import with a correct one
// For demonstration, let's assume the correct path is './utils/pkce'
import { createPKCECodes } from '../utils/pkcs';

export default function Home() {
  const handleLogin = async () => {
    const { codeVerifier, codeChallenge } = await createPKCECodes();

    // Store codeVerifier in session storage
    sessionStorage.setItem('pkce_code_verifier', codeVerifier);

    const ssoUrl = process.env.NEXT_PUBLIC_SSO_AUTHORIZE_URL;
    const clientId = process.env.NEXT_PUBLIC_CLIENT_ID;
    const redirectUri = process.env.NEXT_PUBLIC_REDIRECT_URI;
    const authUrl = `${ssoUrl}?response_type=code&client_id=${encodeURIComponent(clientId!)}&redirect_uri=${encodeURIComponent(redirectUri!)}&scope=openid%20email%20profile&code_challenge_method=S256&code_challenge=${codeChallenge}`;
    console.log(authUrl);
    window.location.href = authUrl;
  };

  const handleLogout = async () => {
    try {
      const response = await fetch('http://localhost:8000/api/admin/v1/logout', {
        method: 'POST',
      });

      if (response.ok) {
        // Handle successful logout, e.g., redirect to login page
        window.location.href = '/';
      } else {
        console.error('Logout failed');
      }
    } catch (error) {
      console.error('Error during logout:', error);
    }
  };

  return (
    <div className="p-4">
      <button className="bg-blue-500 text-white px-4 py-2 rounded" onClick={handleLogin}>
        Login
      </button>
      <button className="bg-red-500 text-white px-4 py-2 rounded ml-2" onClick={handleLogout}>
        Logout
      </button>
    </div>
  );
}
