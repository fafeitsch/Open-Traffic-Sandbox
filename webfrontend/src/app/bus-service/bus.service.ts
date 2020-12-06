import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';
import {environment} from '../../environments/environment';
import {BusInfo} from './types';

@Injectable()
export class BusService {

  constructor(private http: HttpClient) {
  }

  public getBusInfo(id: string): Observable<BusInfo> {
    return this.http.get<BusInfo>(environment.apiUrl + '/api/buses/' + id + '/info');
  }

  public getBusRoute(key: string): Observable<number[][]> {
    return this.http.get<number[][]>(environment.apiUrl + '/api/buses' + key + '/route');
  }
}
