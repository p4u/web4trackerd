const endpoint = "/api/locations";
let locations = [];

fetch(endpoint)
  .then(response => response.json())
  .then(data => {
    // Store the location data in an array
    for (let key in data.locations) {
      locations.push(data.locations[key]);
    }

    // Sort the location data by index
    locations.sort((a, b) => {
      return parseInt(a.index) - parseInt(b.index);
    });

    // Create a map using OpenStreetMap
    const map = L.map("map").setView([locations[0].latitude, locations[0].longitude], 13);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: 'Map data © <a href="https://openstreetmap.org">OpenStreetMap</a> contributors',
      maxZoom: 20
    }).addTo(map);

    // Add markers and connect them
    let previousLatLng = null;
    for (let i = 0; i < locations.length; i++) {
      const latLng = [locations[i].latitude, locations[i].longitude];

      // Add a marker for the current location
      let marker;
      if (i === locations.length - 1) {
        // Last marker should be a different color
        marker = L.marker(latLng, {
          icon: L.icon({
            iconUrl: "/web/last-marker.jpg",
            iconSize: [40, 40],
            iconAnchor: [12, 41]
          })
        }).addTo(map);
      } else {
        marker = L.marker(latLng).addTo(map);
      }

      // Show the timestamp and index when the marker is clicked
      marker.bindPopup(`${i}: ${new Date(locations[i].timestamp).toLocaleString()}`);

      // Connect the marker with the previous one
      if (previousLatLng) {
        L.polyline([previousLatLng, latLng], { color: "red" }).addTo(map);
      }
      previousLatLng = latLng;
    }
  })
  .catch(error => {
    console.error(error);
  });







/*
const endpoint = "/api/locations";
let locations = [];

fetch(endpoint)
  .then(response => response.json())
  .then(data => {
    // Store the location data in an array
    for (let key in data.locations) {
      locations.push(data.locations[key]);
    }

    // Sort the location data by index
    locations.sort((a, b) => {
      return parseInt(a.index) - parseInt(b.index);
    });

    // Create a map using OpenStreetMap
    const map = L.map("map").setView([locations[0].latitude, locations[0].longitude], 13);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: 'Map data © <a href="https://openstreetmap.org">OpenStreetMap</a> contributors',
      maxZoom: 20
    }).addTo(map);

    // Add markers and connect them
    let previousLatLng = null;
    for (let i = 0; i < locations.length; i++) {
      const latLng = [locations[i].latitude, locations[i].longitude];

      // Add a marker for the current location
      let marker;
      if (i === locations.length - 1) {
        // Last marker should be a different color
        marker = L.marker(latLng, {
          icon: L.icon({
            iconUrl: "/web/last-marker.jpg",
            iconSize: [40, 40],
            iconAnchor: [12, 41]
          })
        }).addTo(map);
      } else {
        marker = L.marker(latLng).addTo(map);
      }

      // Show the timestamp when the marker is clicked
      marker.bindPopup(new Date(locations[i].timestamp).toLocaleString());

      // Connect the marker with the previous one
      if (previousLatLng) {
        L.polyline([previousLatLng, latLng], { color: "red" }).addTo(map);
      }
      previousLatLng = latLng;
    }
  })
  .catch(error => {
    console.error(error);
  });




/*
// initialize the map and set its view to a given place and zoom
var map = L.map('map').setView([41.7092, 2.4550], 14);

// Add a tile layer to the map
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
  attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, <a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>',
  maxZoom: 18
}).addTo(map);

// Query the API endpoint every minute
setInterval(function() {
  // Make a GET request to the API endpoint
  fetch('/api/locations')
    .then(response => response.json())
    .then(data => {
      // Get the locations data
      var locations = data.locations;

      // Remove all existing markers from the map
      map.eachLayer(function(layer) {
        if (layer instanceof L.Marker) {
          map.removeLayer(layer);
        }
      });

      // Add markers to the map for each location
      var previousLatLng;
      for (var key in locations) {
        var location = locations[key];
        var latLng = [location.latitude, location.longitude];

        // Create a marker for the location
        var marker = L.marker(latLng);

        // If this is the last location, set the marker color to red
        if (key == Object.keys(locations).length) {
          marker.setIcon(L.icon({
            iconUrl: 'red-marker.png',
            iconSize: [25, 41],
            iconAnchor: [12, 41],
            popupAnchor: [1, -34],
            shadowSize: [41, 41]
          }));
        }

        // Add the marker to the map
        marker.addTo(map);

        // Connect the markers with a line, if there is a previous location
        if (previousLatLng) {
          var polyline = L.polyline([previousLatLng, latLng], {
            color: 'blue',
            weight: 3,
            opacity: 0.5,
            smoothFactor: 1
          }).addTo(map);
        }

        previousLatLng = latLng;
      }
    });
}, 60000);
*/