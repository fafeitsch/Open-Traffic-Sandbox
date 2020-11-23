import {Clipboard} from '@angular/cdk/clipboard';
import {Component, OnDestroy, OnInit, ViewEncapsulation} from '@angular/core';
import * as L from 'leaflet';
import {antPath} from 'leaflet-ant-path';
import {Subscription} from 'rxjs';
import {MapViewStore} from './map-view.store';
import {VehicleLocationService} from './vehicle-location.service';

@Component({
  selector: 'map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss'],
  encapsulation: ViewEncapsulation.None,
  providers: [MapViewStore]
})
export class MapViewComponent implements OnInit, OnDestroy {

  private locationSubscription = Subscription.EMPTY;
  private viewLineSubscription = Subscription.EMPTY;

  private paths = [];

  constructor(private readonly vehicleLocationService: VehicleLocationService,
              private readonly store: MapViewStore,
              private readonly clipboard: Clipboard) {
  }

  ngOnInit() {
    const map = L.map('map').setView([49.80075, 9.93543], 16);

    L.tileLayer('http://localhost:8080/tile/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const markers: { [key: string]: any } = {};

    this.locationSubscription = this.vehicleLocationService.connect().subscribe(location => {
        if (!markers[location.id]) {
          const icon = L.divIcon({
            className: 'map-marker',
            iconAnchor: [10, 10],
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

    const coordinates = L.Control.extend({
      onAdd: overlay => {
        const container = L.DomUtil.create('div');
        overlay.addEventListener('click', e => {
          this.clipboard.copy(`${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`);
          container.innerHTML = `${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`;
        });
        return container;
      }
    });
    map.addControl(new coordinates({position: 'bottomleft'}));

    this.viewLineSubscription = this.store.visibleLines$.subscribe(routes => {
      this.paths.forEach(path => map.removeLayer(path));
      this.paths = [];
      Object.entries(routes).forEach(([key, route]) => {
        this.paths.push(antPath(route.route, {opacity: 1.0, color: route.color, dashArray: [5, 60], delay: 2000}).addTo(map));
      });
    });
  }

  ngOnDestroy(): void {
    this.locationSubscription.unsubscribe();
    this.viewLineSubscription.unsubscribe();
  }
}
