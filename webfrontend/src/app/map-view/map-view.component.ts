import {Component, OnDestroy, OnInit, ViewEncapsulation} from '@angular/core';
import '../../../node_modules/leaflet/dist/leaflet';
import {VehicleLocationService} from '../vehicle-location.service';
import {Subscription} from 'rxjs';
import {Clipboard} from '@angular/cdk/clipboard';

declare let L;

@Component({
  selector: 'map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss'],
  encapsulation: ViewEncapsulation.None
})
export class MapViewComponent implements OnInit, OnDestroy {

  private locationSubscription = Subscription.EMPTY;

  constructor(private readonly vehicleLocationService: VehicleLocationService,
              private readonly clipboard: Clipboard) {
  }

  ngOnInit() {
    const map = L.map('map').setView([49.80075, 9.93543], 16);

    L.tileLayer('http://localhost:8080/tile/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const markers = {};

    this.locationSubscription = this.vehicleLocationService.connect().subscribe(location => {
        if (!markers[location.id]) {
          map.setView(location.loc);
          const icon = L.divIcon({
            className: 'map-marker',
            iconAnchor: [20, 20],
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

    const Coordinates = L.Control.extend({
      onAdd: map => {
        const container = L.DomUtil.create('div');
        map.addEventListener('click', e => {
          this.clipboard.copy(`${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`);
          container.innerHTML = `${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`;
        });
        return container;
      }
    });
    map.addControl(new Coordinates({position: 'bottomleft'}));
  }

  ngOnDestroy(): void {
    this.locationSubscription.unsubscribe();
  }
}
