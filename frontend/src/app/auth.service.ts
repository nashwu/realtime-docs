import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

const KEY = 'jwt_token';
type TokenResp = { token: string; user: { id: string; email: string } };

@Injectable({ providedIn: 'root' })
export class AuthService {
  private base = ''; // Use relative URLs to work with ALB routing
  constructor(private http: HttpClient) {}

  getToken(): string | null { return localStorage.getItem(KEY); }
  setToken(t: string) { if (t?.trim()) localStorage.setItem(KEY, t.trim()); }
  clear() { localStorage.removeItem(KEY); }

  register(email: string, password: string) {
    return this.http.post<TokenResp>(`${this.base}/api/auth/register`, { email, password });
  }
  login(email: string, password: string) {
    return this.http.post<TokenResp>(`${this.base}/api/auth/login`, { email, password });
  }
}
