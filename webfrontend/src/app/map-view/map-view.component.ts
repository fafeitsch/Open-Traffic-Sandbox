import {Component, OnDestroy, OnInit, ViewEncapsulation} from '@angular/core';
import '../../../node_modules/leaflet/dist/leaflet';
import {VehicleLocationService} from '../vehicle-location.service';
import {Subscription} from 'rxjs';

declare let L;

@Component({
  selector: 'app-map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss'],
  encapsulation: ViewEncapsulation.None
})
export class MapViewComponent implements OnInit, OnDestroy {

  private locationSubscription = Subscription.EMPTY;

  constructor(private vehicleLocationService: VehicleLocationService) {
  }

  ngOnInit() {
    const map = L.map('map').setView([49.4805, 9.5608], 13);

    L.tileLayer('http://localhost:8080/tile/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const markers = {};

    this.locationSubscription = this.vehicleLocationService.connect().subscribe(location => {
        if (!markers[location.id]) {
          map.setView(location.loc);
          const icon = L.divIcon({
            className: 'map-marker',
            iconSize: null,
            html: `${location.id}`
          });
          markers[location.id] = L.marker({
            lat: location.loc[0],
            lon: location.loc[1]
          }, {fillOpacity: 1, icon: icon}).addTo(map);
        }
        markers[location.id].setLatLng({lat: location.loc[0], lon: location.loc[1]});
      }
    );
  }

  ngOnDestroy(): void {
    this.locationSubscription.unsubscribe();
  }
}
