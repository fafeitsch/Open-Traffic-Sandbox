import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {WebsocketService} from './websocket.service';

export interface VehicleLocation {
  vehicleId: string;
  coordinate: number[];
}

@Injectable({
  providedIn: 'root'
})
export class VehicleLocationService {

  constructor(private wsService: WebsocketService) {
  }

  listenToLocations(): Observable<VehicleLocation> {
    return this.wsService.connect('ws://localhost:8000/sockets').pipe(map(
      (response: MessageEvent): VehicleLocation => {
        const data = JSON.parse(response.data);
        return {
          vehicleId: data.id,
          coordinate: data.loc
        };
      }));
  }
}
