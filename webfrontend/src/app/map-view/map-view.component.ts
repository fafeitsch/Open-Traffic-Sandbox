import {Clipboard} from '@angular/cdk/clipboard';
import {ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewEncapsulation} from '@angular/core';
import * as L from 'leaflet';
import {antPath} from 'leaflet-ant-path';
import {Subject, Subscription} from 'rxjs';
import {map, takeUntil} from 'rxjs/operators';
import {BusInfo} from '../bus-service/types';
import {MapViewStore} from './map-view.store';
import {VehicleLocationService} from './vehicle-location.service';
import {environment} from '../../environments/environment';

@Component({
  selector: 'map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss'],
  encapsulation: ViewEncapsulation.None,
  providers: [MapViewStore],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class MapViewComponent implements OnInit, OnDestroy {

  destroy$ = new Subject<boolean>();

  private busInfoSubscription = Subscription.EMPTY;

  private paths = [];

  constructor(private readonly vehicleLocationService: VehicleLocationService,
              private readonly store: MapViewStore,
              private readonly clipboard: Clipboard) {
  }

  ngOnInit() {
    const leafletMap = L.map('map').setView([49.80075, 9.93543], 16);

    L.tileLayer(environment.apiUrl + '/tile/{z}/{x}/{y}', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(leafletMap);

    const markers: { [key: string]: any } = {};

    this.vehicleLocationService.connect().pipe(
      takeUntil(this.destroy$),
    ).subscribe(location => {
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
          }, {fillOpacity: 1, icon: icon}).addTo(leafletMap);
          markers[location.id].bindPopup('Loading ...');
          markers[location.id].on('click', e => {
            const popup = e.target.getPopup();
            this.busInfoSubscription.unsubscribe();
            this.store.busSelectionChanged$(location.id);
            this.busInfoSubscription = this.store.currentBus$.pipe(
              takeUntil(this.destroy$),
              map(info => this.formatBusPopup(info))
            ).subscribe(text => popup.setContent(text));
            popup.update();
          });
          markers[location.id].busId = location.id;
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
    leafletMap.addControl(new coordinates({position: 'bottomleft'}));

    this.store.visibleLines$.pipe(takeUntil(this.destroy$)).subscribe(routes => {
      this.paths.forEach(path => leafletMap.removeLayer(path));
      this.paths = [];
      Object.entries(routes).forEach(([key, route]) => {
        this.paths.push(antPath(route.route, {
          opacity: 1.0,
          color: route.color,
          dashArray: [5, 60],
          delay: 2000
        }).addTo(leafletMap));
      });
    });
  }

  private formatBusPopup(info: BusInfo) {
    let lineMarkerCss = '';
    if (info.line !== undefined) {
      lineMarkerCss = `style="background-color:${info.line.color}"`;
    } else {
      lineMarkerCss = 'style="border: solid 1px"';
    }
    return `<div class="center-flex mb-3"><div class="line-marker mr-2" ${lineMarkerCss}></div><strong>${info.id}</strong></div>
            <div>${info.assignment}</div>`;
  }

  ngOnDestroy(): void {
    this.destroy$.next(true);
  }
}
