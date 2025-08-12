import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

const KEY = 'jwt_token';
type TokenResp = { token: string; user: { id: string; email: string } };

@Injectable({ providedIn: 'root' })
export class AuthService {
  private base = 'http://localhost:8080';
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
