export type LoginRequest = {
  username: string;
  password: string;
}

export type RegisterRequest = {
  username: string;
  password: string;
  email?: string;
}

export type User = {
  id: number;
  username: string;
  email?: string;
  created_at?: string;
  updated_at?: string;
}

export type AuthResponse = {
  token: string;
  user: User;
}
