import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { map, withLatestFrom } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { LineServiceModule } from './line-service.module';
import { Line, LineRoute } from './types';

@Injectable({
  providedIn: LineServiceModule,
})
export class LineService {
  constructor(private http: HttpClient) {}

  public getLines(): Observable<Line[]> {
    return this.http.get<Line[]>(environment.apiUrl + '/api/lines');
  }

  public getLine(key: string): Observable<Line> {
    return this.http.get<Line>(environment.apiUrl + '/api/lines/' + key);
  }

  public getLineRoute(key: string): Observable<LineRoute> {
    return this.http.get<number[][]>(environment.apiUrl + '/api/lines/' + key + '/route').pipe(
      withLatestFrom(this.getLine(key)),
      map(([route, line]) => ({ key: key, route: route, color: line.color }))
    );
  }
}
