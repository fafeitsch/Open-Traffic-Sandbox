import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {WebSocketSubject} from 'rxjs/internal-compatibility';
import {delay, filter, map, retryWhen, switchMap} from 'rxjs/operators';
import {webSocket} from 'rxjs/webSocket';
import {environment} from '../../environments/environment';

export interface VehicleLocation {
  id: string;
  loc: number[];
  departure?: number;
  stopId?: string;
}

@Injectable({
  providedIn: 'root'
})
export class VehicleLocationService {

  private connection$: WebSocketSubject<VehicleLocation>;

  connect(): Observable<VehicleLocation> {
    let url = location.origin;
    if (!environment.production) {
      url = environment.apiUrl;
    }
    return of(url).pipe(
      filter(apiUrl => !!apiUrl),
      map(apiUrl => apiUrl.replace(/^http/, 'ws') + '/sockets'),
      switchMap(wsUrl => {
        if (this.connection$) {
          return this.connection$;
        } else {
          this.connection$ = webSocket(wsUrl);
          return this.connection$;
        }
      }),
      retryWhen((errors) => errors.pipe(delay(10))));
  }
}
