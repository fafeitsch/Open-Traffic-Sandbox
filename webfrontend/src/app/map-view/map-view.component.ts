import { Component, OnInit } from '@angular/core';
declare let L;
import '../../../node_modules/leaflet/dist/leaflet';
import {VehicleLocationService} from '../vehicle-location.service';

@Component({
  selector: 'app-map-view',
  templateUrl: './map-view.component.html',
  styleUrls: ['./map-view.component.css']
})
export class MapViewComponent implements OnInit {

  constructor(private vehicleLocationService: VehicleLocationService) { }

  ngOnInit() {
    const map = L.map('map').setView([51.505, -0.09], 13);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const markers = {};

    this.vehicleLocationService.locations.subscribe(location => {
        if (!markers[location.vehicleId]) {
          map.setView(location.coordinate);
          const marker = L.circleMarker({lat: location.coordinate[0], lon: location.coordinate[1]}, {fillOpacity: 1}).addTo(map);
          markers[location.vehicleId] = marker;
        }
        markers[location.vehicleId].setLatLng({lat: location.coordinate[0], lon: location.coordinate[1]});
      }
    );
  }

}
