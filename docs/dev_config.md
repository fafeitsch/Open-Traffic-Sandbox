## Development Configuration

OTS requires two third-party services to be available:
1. an OSRM server. The OSRM server provides OTS with the necessary route information.
It is possible to use a public OSRM server, but this is discouraged for development reasons
in order to minimize the load and traffic on the public servers. 
2. an OSM Tile Server. For the same reason as we are using our own OSRM server, we also 
our own Tile Server to spare the public OSM server. At the moment, the OSM Tile Server
is hardcoded in the `map-view` component of the frontend to `localhost:8080`. If you wan to
use the public server nonetheless, or you have a tile server available at a different address, please change the URL there.

To be continued.
