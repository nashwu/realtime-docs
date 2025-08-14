import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AppComponent } from './app.component';
import { WSSyncService } from './ws-sync.service';
import { DocsService } from './docs.service';
import { AuthService } from './auth.service';
import { of } from 'rxjs';

// Mock services
class MockWSSyncService {
  connect() {}
  disconnect() {}
  ydoc = { 
    getText: () => ({ 
      toString: () => '',
      observe: () => {}
    }) 
  };
  setPresence() {}
  onText() {}
  onStatus() {}
  onPresence() {}
}

class MockDocsService {
  list() { 
    return of([]); 
  }
  create(title: string) { 
    return of({ id: '1', title, version: 1, updatedAt: new Date().toISOString() }); 
  }
  getBlob(id: string) { 
    return of(new Blob()); 
  }
}

class MockAuthService {
  login() { return Promise.resolve('test-token'); }
  logout() {}
}

describe('AppComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent, HttpClientTestingModule],
      providers: [
        { provide: WSSyncService, useClass: MockWSSyncService },
        { provide: DocsService, useClass: MockDocsService },
        { provide: AuthService, useClass: MockAuthService }
      ]
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(AppComponent);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });

  it('should render textarea', () => {
    const fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector('textarea')).toBeTruthy();
  });
});
