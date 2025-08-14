import { Component, ElementRef, OnInit, ViewChild, AfterViewInit, NgZone } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { WSSyncService } from './ws-sync.service';
import { DocsService, DocItem } from './docs.service';
import { AuthService } from './auth.service';
import * as Y from 'yjs';

type PeerCursor = { clientId: number; name: string; color: string; anchor: number|null; head: number|null };

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit, AfterViewInit {
  @ViewChild('ta', { static: true }) taRef!: ElementRef<HTMLTextAreaElement>;

  // Auth UI
  email = '';
  password = '';
  userEmail = localStorage.getItem('user_email') ?? '';

  // Docs UX state
  docs: DocItem[] = [];
  docId = '';
  newTitle = '';
  loading = false;
  creating = false;

  // Editor state
  status = 'disconnected';
  text = '';
  peers: PeerCursor[] = [];

  constructor(
    public sync: WSSyncService,
    private zone: NgZone,
    private api: DocsService,
    private auth: AuthService
  ) {}

  // lifecycle
  ngOnInit(): void {
    // bind sync events
    this.sync.onStatus(s => this.status = s);
    this.sync.onText(t => { if (this.text !== t) this.text = t; this.renderCursors(); });
    this.sync.onPresence(p => {
      this.peers = p.filter(x => x.clientId !== this.sync.ydoc.clientID);
      this.renderCursors();
    });

    // load docs (and auto-select/connect below)
    this.refreshDocs();
  }

  ngAfterViewInit(): void {
    const ta = this.taRef.nativeElement;
    const bump = () => this.zone.runOutsideAngular(() => {
      requestAnimationFrame(() => this.zone.run(() => this.sendMyCursor()));
    });
    ta.addEventListener('keyup', bump);
    ta.addEventListener('keydown', bump);
    ta.addEventListener('mouseup', bump);
    ta.addEventListener('input', bump);
    document.addEventListener('selectionchange', () => {
      if (document.activeElement === ta) bump();
    });
    ta.addEventListener('scroll', () => this.renderCursors());
    window.addEventListener('resize', () => this.renderCursors());
  }

  // auth
  authed() {
    const hasToken = !!localStorage.getItem('jwt_token');
    return hasToken && !!this.userEmail;
  }

  doRegister() {
    const e = this.email.trim(), p = this.password;
    if (!e || p.length < 8) { alert('Enter a valid email and 8+ char password'); return; }
    this.auth.register(e, p).subscribe({
      next: res => {
        this.auth.setToken(res.token);
        this.userEmail = res.user.email;
        localStorage.setItem('user_email', this.userEmail);
        this.refreshDocs(true); // force refresh + maybe auto-select
      },
      error: err => { alert('Register failed'); console.error(err); }
    });
  }

  doLogin() {
    const e = this.email.trim(), p = this.password;
    if (!e || !p) return;
    this.auth.login(e, p).subscribe({
      next: res => {
        this.auth.setToken(res.token);
        this.userEmail = res.user.email;
        localStorage.setItem('user_email', this.userEmail);
        this.refreshDocs(true);
      },
      error: err => { alert('Login failed'); console.error(err); }
    });
  }

  doLogout() {
    this.auth.clear();
    localStorage.removeItem('user_email');
    this.userEmail = '';
    this.docs = [];
    this.docId = '';
    this.text = '';
    this.status = 'disconnected';
  }

  // docs
  refreshDocs(autoPick: boolean = false) {
    this.loading = true;
    this.api.list().subscribe({
      next: d => {
        this.docs = d;
        this.loading = false;

        // Auto-select the current doc if it exists else first doc if asked
        if (!this.docId && autoPick && this.docs.length > 0) {
          this.docId = this.docs[0].id;
          this.onDocChange(this.docId); // auto-connect
        }
      },
      error: _ => { this.loading = false; }
    });
  }

  createDoc() {
    const title = this.newTitle?.trim();
    if (!title) return;
    this.creating = true;
    this.api.create(title).subscribe({
      next: d => {
        this.creating = false;
        this.newTitle = '';
        this.docs = [d, ...this.docs];
        
        // Auto-select & auto-connect newly created doc
        this.docId = d.id;
        this.onDocChange(this.docId);
      },
      error: _ => { this.creating = false; }
    });
  }

  // Auto-connect whenever the dropdown changes
  onDocChange(nextId: string) {
    this.docId = nextId;
    if (!this.docId) return;

    // Reset document FIRST to ensure clean state
    if (this.sync.getLastDocId() !== nextId) {
      this.sync.resetDocument();
      this.text = ''; // Clear UI immediately
    }

    // Load snapshot THEN open WS
    this.api.getBlob(this.docId).subscribe({
      next: async (blob) => {
        try {
          const buf = new Uint8Array(await blob.arrayBuffer());
          if (buf.length > 0) {
            Y.applyUpdate(this.sync.ydoc, buf); // pure Yjs bytes
          }
          
          // Update the local text state to reflect the new document
          this.text = this.sync.ytext.toString();
        } catch (e) {
          console.warn('Snapshot apply failed; continuing without it', e);
          this.text = '';
        }
        this.sync.connect(this.docId);
      },
      error: () => {
        // Start fresh on error
        this.text = '';
        this.sync.connect(this.docId);
      }
    });
  }

  // editor
  onInput(): void {
    this.sync.updateLocal(this.text);
    this.sendMyCursor();
  }
  onSelect(): void { this.sendMyCursor(); }

  private sendMyCursor() {
    const ta = this.taRef.nativeElement;
    this.sync.updateMyCursor(ta.selectionStart, ta.selectionEnd);
    this.renderCursors();
  }

  // cursors render
  private renderCursors() {
    const ta = this.taRef?.nativeElement;
    if (!ta) return;

    const wrap = ta.closest('.editor-wrap') as HTMLElement | null;
    if (!wrap) return;

    const overlay = wrap.querySelector<HTMLElement>('.cursor-overlay');
    if (!overlay) return;

    overlay.innerHTML = '';

    for (const p of this.peers) {
      if (p.head == null) continue;
      const caret = caretRectFor(ta, overlay, p.head);
      if (!caret) continue;

      const box = document.createElement('div');
      box.style.position = 'absolute';
      box.style.left = `${caret.left - 6}px`;
      box.style.top = `${caret.top}px`;
      box.style.width = `12px`;
      box.style.height = `${caret.height}px`;
      box.style.pointerEvents = 'auto';
      box.style.background = 'transparent';
      overlay.appendChild(box);

      const bar = document.createElement('div');
      bar.className = 'caret';
      bar.style.position = 'absolute';
      bar.style.left = `6px`;
      bar.style.top = `0px`;
      bar.style.height = `${caret.height}px`;
      bar.style.width = '0px';
      bar.style.borderLeft = `2px solid ${p.color}`;
      bar.style.pointerEvents = 'none';
      box.appendChild(bar);

      const label = document.createElement('div');
      label.className = 'label';
      label.textContent = p.name;
      label.style.position = 'absolute';
      label.style.left = `8px`;
      label.style.top = `${Math.max(0, caret.top - 25)}px`;
      label.style.background = p.color;
      label.style.color = '#fff';
      label.style.padding = '1px 4px';
      label.style.borderRadius = '3px';
      label.style.lineHeight = '1.2';
      label.style.whiteSpace = 'nowrap';
      label.style.opacity = '0';
      label.style.transition = 'opacity .12s ease-in-out, font-size .12s ease-in-out';
      label.style.pointerEvents = 'none';
      label.style.fontSize = '11px';
      box.addEventListener('mouseenter', () => { label.style.opacity = '1'; label.style.fontSize = '10px'; });
      box.addEventListener('mouseleave', () => { label.style.opacity = '0'; label.style.fontSize = '11px'; });
      overlay.appendChild(label);
    }
  }
}

// helpers
function caretRectFor(ta: HTMLTextAreaElement, overlay: HTMLElement, index: number) {
  const pre = mirror(ta);
  const CARET_BIAS_PX = 10;
  pre.textContent = ta.value.slice(0, index);
  const marker = document.createElement('span');
  marker.textContent = '\u200b';
  pre.appendChild(marker);
  document.body.appendChild(pre);

  const overlayRect = overlay.getBoundingClientRect();
  const markRect    = marker.getBoundingClientRect();

  let left = (markRect.left - overlayRect.left) + ta.scrollLeft;
  let top  = (markRect.top  - overlayRect.top)  + ta.scrollTop;

  const cs = getComputedStyle(ta);
  const height   = parseFloat(cs.lineHeight || '16');
  const fontSize = parseFloat(cs.fontSize || '14');
  const extraSpace     = Math.max(0, height - fontSize);
  const verticalAdjust = extraSpace / 2;
  const biasPx         = CARET_BIAS_PX || Math.round(extraSpace * 0.25);
  top = top - verticalAdjust - biasPx;

  document.body.removeChild(pre);
  return { left, top, height };
}

function mirror(ta: HTMLTextAreaElement) {
  const s = getComputedStyle(ta);
  const taRect = ta.getBoundingClientRect();
  const pre = document.createElement('pre');
  pre.className = 'ta-mirror';
  pre.style.position = 'fixed';
  pre.style.left = `${taRect.left}px`;
  pre.style.top  = `${taRect.top}px`;
  pre.style.visibility   = 'hidden';
  pre.style.whiteSpace   = 'pre-wrap';
  pre.style.wordBreak    = 'break-word';
  pre.style.overflow     = 'hidden';
  pre.style.font          = s.font;
  pre.style.lineHeight    = s.lineHeight;
  pre.style.letterSpacing = s.letterSpacing;
  (pre.style as any).tabSize = (s as any).tabSize || '4';
  pre.style.textAlign     = s.textAlign;
  pre.style.padding   = s.padding;
  pre.style.border    = s.border;
  pre.style.boxSizing = s.boxSizing;
  pre.style.width  = `${ta.clientWidth}px`;
  pre.style.height = `${ta.clientHeight}px`;
  return pre;
}
