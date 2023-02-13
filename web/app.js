// initialize the map and set its view to a given place and zoom
var map = L.map('map').setView([41.7092, 2.4550], 14);

// add an OpenStreetMap tile layer
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
  attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
}).addTo(map);

// create a layer group to hold the markers
var markerLayer = L.layerGroup().addTo(map);

// create an array to store the markers
var markers = [];

// create an array to store the polyline
var polyline;

// function to query the API and update the markers
function updateMarkers() {
  // make the API request
  fetch('/api/locations')
    .then(response => response.json())
    .then(data => {
      // remove all markers from the marker layer
      markerLayer.clearLayers();

      // reset the markers array
      markers = [];

      // loop through the data and add a marker for each location
      data.forEach(location => {
        var marker = L.marker([location.latitude, location.longitude]).addTo(markerLayer);

        // add a popup with the timestamp
        marker.bindPopup(new Date(location.timestamp).toLocaleString());

        // add the marker to the markers array
        markers.push(marker);
      });

      // remove the previous polyline if it exists
      if (polyline) {
        map.removeLayer(polyline);
      }

      // create a new polyline connecting the markers
      polyline = L.polyline(markers.map(marker => marker.getLatLng()), {
        color: 'red'
      }).addTo(map);
    });
}

// call the updateMarkers function every minute
setInterval(updateMarkers, 60000);

// call the updateMarkers function when the page loads
updateMarkers();
