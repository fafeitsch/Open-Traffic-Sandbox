Open Traffic Toolbox
=====================

What is _Open Traffic Toolbox_?
-------------------------------

At the moment it is just a curious side project of mine. I simply wanted to learn the
Go programming language as well as Angular. The project is about maps and traffic, but
as of now, I do not know where it is headed.

What will happen if I run some code in the repository?
------------------------------------

As of now, the whole thing is just a proof of concept, nothing more. The code is not
clean, there is no documentation, no possibility to set some hard coded constants.
At the moment, you need an [OSRM server](https://github.com/Project-OSRM/osrm-backend/wiki/Running-OSRM) running at ```localhost:5000``` with the
street data of at least [lower franconia](http://download.geofabrik.de/europe/germany/bayern/unterfranken.html) loaded. If you then ran ```go run otrserver.go``` in package ```main```,
then you can navigate to ```http://localhost:8000``` and see two dots wandering
through the city of Würzburg.

Most probably there are some broken test cases and other bugs, I am working on clearing them out :).

License
------------------------------------
See the LICENSE file.

This repository contains data derived from Open Street Map (stops.geojson). The data in this file is licensed under: © OpenStreetMap contributors.