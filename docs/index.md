## Open Transport Sandbox

Open Transport Sandbox (OTS) aims to be an open source simulation platform/framework. As of November 2020, it is in a
very early stage of development and has only a limited set of features. At the moment, there is no official release
version.

### Features

Already implemented:

* Definition of simple timetables.
* Definition of buses with assignments (e.g. when to serve a certain line).
* Buses serve their assignments, and their locations are shown on a map.
* To define stops, real OSM data can be used (gson format).

Planned for the future:

* Incorporating passengers with routes that use the buses.
* More information in the frontend.
* More guidance for creating scenarios.

Currently out of scope:

* User management and authentication

### Developing and Running

In order to run OTS, the following prerequisites must be met at the moment.

* Go 1.14 or higher installed (needed for the backend)
* npm installed  (needed for the frontend)
* a running OSRM server (needed for route querying). A simple docker container is available for
  that ([external resource](https://hub.docker.com/r/osrm/osrm-backend/])). Alternatively, the public OSRM
  API [Demo](https://github.com/Project-OSRM/osrm-backend/wiki/Demo-server) can be used. However, please consider
  setting up your own server to reduce the load on the donation-powered demo server.
* a tile server. A simple docker container is available for
  that ([external resource](https://github.com/Overv/openstreetmap-tile-server)). Alternatively, a public OSM server can
  be used, a list can be found [here](https://wiki.openstreetmap.org/wiki/Tile_servers).

Running OTS involves the following steps:

1. Create a custom scenario or use the default scenario (see `samples` directory). Use the default scenario as blueprint
   for your own scenario.
2. Configure your OSRM server with suitable data. For the default scenario, you need `unterfanken-latest.pbf`.
3. Configure the tile server with suitable data. For the default scenario, you need `unterfanken-latest.pbf`.
4. Build the frontend with `ng build` inside the `webfrontend` directory.
5. Run the backend program located in `pkg/main/otsserver.go`. Use the `--help` flag for a documentation of that
   command. Provide the locations of your OSRM server, and your tile server with the corresponding command line flags.
6. Navigate to the appropriate localhost address (default is `localhost:9551`).



