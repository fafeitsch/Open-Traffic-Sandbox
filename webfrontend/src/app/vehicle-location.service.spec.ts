import { TestBed } from '@angular/core/testing';

import { VehicleLocationService } from './vehicle-location.service';

describe('VehicleLocationService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: VehicleLocationService = TestBed.get(VehicleLocationService);
    expect(service).toBeTruthy();
  });
});
