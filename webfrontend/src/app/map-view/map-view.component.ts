import {Component, OnDestroy, OnInit} from '@angular/core';
import '../../../node_modules/leaflet/dist/leaflet';
import {VehicleLocationService} from '../vehicle-location.service';
import {Subscription} from 'rxjs';

declare let L;

@Component({
  selector: 'app-map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.css']
})
export class MapViewComponent implements OnInit, OnDestroy {

  private locationSubscription = Subscription.EMPTY;

  constructor(private vehicleLocationService: VehicleLocationService) {
  }

  ngOnInit() {
    const map = L.map('map').setView([51.505, -0.09], 13);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const markers = {};

    this.locationSubscription = this.vehicleLocationService.connect().subscribe(location => {
        if (!markers[location.id]) {
          map.setView(location.loc);
          markers[location.id] = L.circleMarker({
            lat: location.loc[0],
            lon: location.loc[1]
          }, {fillOpacity: 1}).addTo(map);
        }
        markers[location.id].setLatLng({lat: location.loc[0], lon: location.loc[1]});
      }
    );
  }

  ngOnDestroy(): void {
    this.locationSubscription.unsubscribe();
  }
}
