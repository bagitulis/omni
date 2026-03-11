/**
 * Authentication Types
 * API types use snake_case to match backend JSON response
 */

export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

export interface LoginResponse {
  token: string;
  access_token: string;
  tenant_id: string;
  user: User;
  expires_in: number;
  requiresCaptcha?: boolean;
  isLocked?: boolean;
  lockMinutesRemaining?: number;
}

export interface LoginErrorResponse {
  error: string;
  code?: string;
  requiresCaptcha?: boolean;
  isLocked?: boolean;
  lockMinutesRemaining?: number;
}

export type RegisterResponse = LoginResponse;
