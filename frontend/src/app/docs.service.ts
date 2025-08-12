import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

export type DocItem = {
  id: string;
  title: string;
  version: number;
  updatedAt: string; // ISO timestamp
};

@Injectable({ providedIn: 'root' })
export class DocsService {
  private base = 'http://localhost:8080'; // API root

  constructor(private http: HttpClient) {}

  // Fetch list of docs (protected route)
  list() {
    return this.http.get<DocItem[]>(`${this.base}/api/docs`);
  }

  // Create a new doc
  create(title: string) {
    return this.http.post<DocItem>(`${this.base}/api/docs`, { title });
  }

  // Get the raw Yjs doc snapshot as a Blob
  getBlob(id: string) {
    return this.http.get(`${this.base}/api/docs/${id}`, { responseType: 'blob' });
  }
}
