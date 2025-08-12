import { Injectable } from '@angular/core';
import * as Y from 'yjs';
import { Awareness } from 'y-protocols/awareness';
import { encodeAwarenessUpdate, applyAwarenessUpdate } from 'y-protocols/awareness';

// ---- wire protocol codes ----
const MSG_UPDATE    = 1; // Yjs incremental
const MSG_SYNC_REQ  = 2; // ask peers for full state
const MSG_SYNC_RES  = 3; // full state as update
const MSG_AWARENESS = 4; // awareness update

type TextListener = (t: string) => void;
type StatusListener = (s: string) => void;

// What we store in Awareness per client
interface AwarenessState {
  name?: string;
  color?: string;
  cursor?: {
    anchor: Y.RelativePosition | null;
    head:   Y.RelativePosition | null;
  };
}

export type Presence = {
  clientId: number;
  name: string;
  color: string;
  anchor: number | null;
  head: number | null;
};
type PresenceListener = (p: Presence[]) => void;

@Injectable({ providedIn: 'root' })
export class WSSyncService {
  private ws?: WebSocket;
  private lastDocId?: string;

  readonly ydoc = new Y.Doc();
  readonly ytext = this.ydoc.getText('t');
  readonly awareness = new Awareness(this.ydoc);

  private textListeners: TextListener[] = [];
  private statusListeners: StatusListener[] = [];
  private presenceListeners: PresenceListener[] = [];

  private applyingRemote = false;
  private reconnectTimer?: any;
  private readonly reconnectDelay = 1500;

  private name = `user-${Math.floor(Math.random() * 1000)}`;
  private color = pickColor();

  constructor() {
    // Broadcast local Yjs ops
    this.ydoc.on('update', (update: Uint8Array, origin: unknown) => {
      if (origin === 'local') this.send(MSG_UPDATE, update);
      this.emitText();
      this.emitPresence();
      this.scheduleSnapshot();
    });

    // Broadcast awareness changes
    this.awareness.on('update', (ev: { added: number[]; updated: number[]; removed: number[] }) => {
      const changed = ev.added.concat(ev.updated, ev.removed);
      const u = encodeAwarenessUpdate(this.awareness, changed);
      this.send(MSG_AWARENESS, u);
      this.emitPresence();
    });

    // Init my awareness (name/color; cursor added later)
    this.awareness.setLocalState({
      name: this.name,
      color: this.color,
      cursor: { anchor: null, head: null }
    } as AwarenessState);
  }

  // public API used by your component
  onText(cb: TextListener) { this.textListeners.push(cb); }
  onStatus(cb: StatusListener) { this.statusListeners.push(cb); }
  onPresence(cb: PresenceListener) { this.presenceListeners.push(cb); }

  connect(docId: string): void {
    this.lastDocId = docId;
    this.emitStatus('connecting');

    try {
      this.ws?.close();
      this.ws = new WebSocket(`ws://localhost:8080/ws?docId=${encodeURIComponent(docId)}`);
      this.ws.binaryType = 'arraybuffer';

      this.ws.onopen = () => {
        this.emitStatus('connected');
        // Ask peers for full state, and announce my current awareness
        this.send(MSG_SYNC_REQ, new Uint8Array(0));
        const u = encodeAwarenessUpdate(this.awareness, [this.ydoc.clientID]);
        this.send(MSG_AWARENESS, u);
      };

      this.ws.onclose = () => { this.emitStatus('reconnecting'); this.scheduleReconnect(); };
      this.ws.onerror = () => this.emitStatus('error');

      this.ws.onmessage = (ev) => {
        const bytes = new Uint8Array(ev.data as ArrayBuffer);
        if (bytes.length === 0) return;
        const type = bytes[0];
        const payload = bytes.subarray(1);

        switch (type) {
          case MSG_UPDATE:
          case MSG_SYNC_RES: {
            this.applyingRemote = true;
            try { Y.applyUpdate(this.ydoc, payload); }
            finally { this.applyingRemote = false; }
            break;
          }
          case MSG_SYNC_REQ: {
            const full = Y.encodeStateAsUpdate(this.ydoc);
            this.send(MSG_SYNC_RES, full);
            break;
          }
          case MSG_AWARENESS: {
            applyAwarenessUpdate(this.awareness, payload, this);
            this.emitPresence();
            break;
          }
        }
      };
    } catch {
      this.emitStatus('error');
      this.scheduleReconnect();
    }
  }

  updateLocal(next: string): void {
    if (this.applyingRemote) return;
    const cur = this.ytext.toString();
    if (cur === next) return;

    this.ydoc.transact(() => {
      this.ytext.delete(0, this.ytext.length);
      this.ytext.insert(0, next);
    }, 'local');
  }

  /** Update the caret/selection presence from absolute indices in current text */
  updateMyCursor(anchor: number, head: number) {
    const relAnchor = fromAbs(this.ytext, anchor);
    const relHead   = fromAbs(this.ytext, head);
    const cur = (this.awareness.getLocalState() as AwarenessState) ?? {};
    this.awareness.setLocalState({
      ...cur,
      cursor: { anchor: relAnchor, head: relHead }
    });
  }

  // internals
  private emitText() { this.textListeners.forEach(cb => cb(this.ytext.toString())); }

  private emitStatus(s: string) { this.statusListeners.forEach(cb => cb(s)); }

  private emitPresence() {
    const peers: Presence[] = [];
    const states = this.awareness.getStates(); // Map<number, AwarenessState>
    states.forEach((st: AwarenessState | undefined, clientId) => {
      const name  = st?.name  ?? `user-${clientId}`;
      const color = st?.color ?? '#999';
      const anchor = toAbs(this.ydoc, st?.cursor?.anchor ?? null);
      const head   = toAbs(this.ydoc, st?.cursor?.head   ?? null);
      peers.push({ clientId, name, color, anchor, head });
    });
    this.presenceListeners.forEach(cb => cb(peers));
  }

  private send(type: number, payload: Uint8Array) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    const buf = new Uint8Array(1 + payload.length);
    buf[0] = type;
    buf.set(payload, 1);
    this.ws.send(buf);
  }

  private scheduleReconnect() {
    if (this.reconnectTimer || !this.lastDocId) return;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = undefined;
      this.connect(this.lastDocId!);
    }, this.reconnectDelay);
  }

  private snapshotTimer: any = null;

  private scheduleSnapshot() {
    clearTimeout(this.snapshotTimer);
    this.snapshotTimer = setTimeout(() => {
      const full = Y.encodeStateAsUpdate(this.ydoc);
      // reuse the existing send() to prepend the 1-byte type (3)
      this.send(MSG_SYNC_RES, full);
    }, 300); // debounce window; tweak as you like
  }

}



// relative/absolute conversions for cursor positions
function fromAbs(text: Y.Text, index: number | null): Y.RelativePosition | null {
  if (index == null) return null;
  return Y.createRelativePositionFromTypeIndex(text, index);
}
function toAbs(doc: Y.Doc, rel: Y.RelativePosition | null): number | null {
  if (!rel) return null;
  const abs = Y.createAbsolutePositionFromRelativePosition(rel, doc);
  return abs?.index ?? null;
}
function pickColor() {
  const colors = ['#F87171','#34D399','#60A5FA','#FBBF24','#A78BFA','#F472B6','#4ADE80','#22D3EE'];
  return colors[Math.floor(Math.random() * colors.length)];
}
