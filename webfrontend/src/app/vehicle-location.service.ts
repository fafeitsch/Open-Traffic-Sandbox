import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {delay, filter, map, retryWhen, switchMap} from 'rxjs/operators';
import {environment} from '../environments/environment';
import {WebSocketSubject} from 'rxjs/internal-compatibility';
import {webSocket} from 'rxjs/webSocket';

export interface VehicleLocation {
  id: string;
  loc: number[];
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
    console.log(url);
    return of(url).pipe(
      filter(apiUrl => !!apiUrl),
      map(apiUrl => apiUrl.replace(/^http/, 'ws') + '/sockets'),
      switchMap(wsUrl => {
        console.log(wsUrl);
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
