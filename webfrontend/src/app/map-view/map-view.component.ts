import { Clipboard } from '@angular/cdk/clipboard';
import { ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewEncapsulation } from '@angular/core';
import { Control, divIcon, DomUtil, map as createLeaflet, marker as createMarker, tileLayer } from 'leaflet';
import { antPath } from 'leaflet-ant-path';
import { Subject, Subscription } from 'rxjs';
import { filter, map, takeUntil } from 'rxjs/operators';
import { BusInfo } from '../bus-service/types';
import { MapViewStore } from './map-view.store';
import { VehicleLocationService } from './vehicle-location.service';
import { environment } from '../../environments/environment';

@Component({
  selector: 'map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.scss'],
  encapsulation: ViewEncapsulation.None,
  providers: [MapViewStore],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class MapViewComponent implements OnInit, OnDestroy {
  destroy$ = new Subject<boolean>();

  private busInfoSubscription = Subscription.EMPTY;

  private paths: any[] = [];

  constructor(
    private readonly vehicleLocationService: VehicleLocationService,
    private readonly store: MapViewStore,
    private readonly clipboard: Clipboard
  ) {}

  ngOnInit() {
    const leafletMap = createLeaflet('map').setView([49.80075, 9.93543], 16);

    tileLayer(environment.apiUrl + '/tile/{z}/{x}/{y}', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(leafletMap);

    const markers: { [key: string]: any } = {};

    this.vehicleLocationService
      .connect()
      .pipe(takeUntil(this.destroy$))
      .subscribe(location => {
        let marker = markers[location.id];
        if (!marker) {
          marker = createMarker(
            {
              lat: location.loc[0],
              lng: location.loc[1],
            },
            { opacity: 1 }
          ).addTo(leafletMap);
          markers[location.id] = marker;
          marker.bindPopup('Loading ...');
          marker.on('click', (e: any) => {
            const popup = e.target.getPopup();
            this.busInfoSubscription.unsubscribe();
            this.store.busSelectionChanged$(location.id);
            this.busInfoSubscription = this.store.currentBus$
              .pipe(
                takeUntil(this.destroy$),
                filter(info => !!info),
                map(info => this.formatBusPopup(info!))
              )
              .subscribe(text => popup.setContent(text));
            popup.update();
          });
          marker.busId = location.id;
        }
        let className = 'driving-bus';
        if (location.stopId !== undefined) {
          className = 'waiting-bus';
        }
        const icon = divIcon({
          className: className,
          iconAnchor: [10, 10],
          iconSize: undefined,
          html: `${location.id}`,
        });
        marker.setIcon(icon);
        marker.setLatLng({ lat: location.loc[0], lon: location.loc[1] });
      });

    const coordinates = Control.extend({
      onAdd: (overlay: any) => {
        const container = DomUtil.create('div');
        overlay.addEventListener('click', (e: any) => {
          this.clipboard.copy(`${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`);
          container.innerHTML = `${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`;
        });
        return container;
      },
    });
    leafletMap.addControl(new coordinates({ position: 'bottomleft' }));

    this.store.visibleLines$.pipe(takeUntil(this.destroy$)).subscribe(routes => {
      this.paths.forEach(path => leafletMap.removeLayer(path));
      this.paths = [];
      Object.entries(routes).forEach(([key, route]) => {
        this.paths.push(
          antPath(route.route, {
            opacity: 1.0,
            color: route.color,
            dashArray: [5, 60],
            delay: 2000,
          }).addTo(leafletMap)
        );
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
