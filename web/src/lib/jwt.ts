// Utility to decode JWT tokens (without verification, for client-side use only)

export interface JWTPayload {
  user_id: number;
  username: string;
  exp: number;
  iat: number;
}

export function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) {
      return null;
    }

    const payload = parts[1];
    if (payload === undefined) {
      return null;
    }
    const decoded = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    const parsed = JSON.parse(decoded);

    // Check if token is expired
    if (parsed.exp && parsed.exp * 1000 < Date.now()) {
      return null;
    }

    return parsed as JWTPayload;
  } catch (error) {
    console.error("Failed to decode JWT:", error);
    return null;
  }
}
