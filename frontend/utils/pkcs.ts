export function generateRandomString(length = 43) {
    const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
    let result = '';
    for (let i = 0; i < length; i++) {
      const randomIndex = Math.floor(Math.random() * charset.length);
      result += charset.charAt(randomIndex);
    }
    return result;
  }
  
  async function sha256(plain: string) {
    const encoder = new TextEncoder();
    const data = encoder.encode(plain);
    return crypto.subtle.digest('SHA-256', data);
  }
  
  function base64UrlEncode(buffer: ArrayBuffer) {
    const bytes = new Uint8Array(buffer);
    let str = '';
    bytes.forEach((b) => (str += String.fromCharCode(b)));
    return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }
  
  export async function createPKCECodes() {
    const codeVerifier = generateRandomString();
    const hashed = await sha256(codeVerifier);
    const codeChallenge = base64UrlEncode(hashed);
    return { codeVerifier, codeChallenge };
  }
  