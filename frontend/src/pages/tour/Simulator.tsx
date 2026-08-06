import { useEffect, useRef, useState } from 'react';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import { positionAPI } from '../../api';
import './Tour.scss';

interface Location {
  lat: number;
  lng: number;
}

const createCurrentLocationIcon = () => {
  return L.divIcon({
    html: `
      <svg width="32" height="40" viewBox="0 0 32 40" xmlns="http://www.w3.org/2000/svg">
        <g>
          <path d="M16 0C8.268 0 2 6.268 2 14c0 7.732 14 26 14 26s14-18.268 14-26c0-7.732-6.268-14-14-14z" 
                fill="#F45F00" stroke="white" stroke-width="2"/>
          <circle cx="16" cy="14" r="5" fill="white"/>
        </g>
      </svg>
    `,
    iconSize: [32, 40],
    iconAnchor: [16, 40],
    popupAnchor: [0, -40],
    className: 'current-location-marker'
  });
};

export default function Simulator() {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markerRef = useRef<L.Marker | null>(null);
  const [currentLocation, setCurrentLocation] = useState<Location | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return;

    const savedLocation = localStorage.getItem('simulatorLocation');
    let initialLat = 44.8176;
    let initialLng = 20.4569;
    let initialZoom = 6;

    if (savedLocation) {
      try {
        const location = JSON.parse(savedLocation);
        initialLat = location.lat;
        initialLng = location.lng;
        initialZoom = 10;
        setCurrentLocation(location);
      } catch (e) {
        console.error('Error parsing saved location:', e);
      }
    }

    const map = L.map(mapContainerRef.current).setView([initialLat, initialLng], initialZoom);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
      maxZoom: 19,
    }).addTo(map);

    if (savedLocation) {
      try {
        const location = JSON.parse(savedLocation);
        const marker = L.marker([location.lat, location.lng], {
          icon: createCurrentLocationIcon(),
        }).addTo(map);
        marker.bindPopup(
          `<div><strong>Current Location</strong><br/>Lat: ${location.lat.toFixed(6)}<br/>Lng: ${location.lng.toFixed(6)}</div>`,
        );
        markerRef.current = marker;
      } catch (e) {
        console.error('Error parsing saved location:', e);
      }
    }

    const handleMapClick = async (e: L.LeafletMouseEvent) => {
      const { lat, lng } = e.latlng;
      const newLocation = { lat, lng };

      // Sacuvaj u localStorage
      localStorage.setItem('simulatorLocation', JSON.stringify(newLocation));

      // Prikaži marker
      if (markerRef.current) {
        markerRef.current.remove();
      }

      const marker = L.marker([lat, lng], {
        icon: createCurrentLocationIcon(),
      }).addTo(map);
      marker.bindPopup(
        `<div><strong>Location Updated</strong><br/>Lat: ${lat.toFixed(6)}<br/>Lng: ${lng.toFixed(6)}</div>`
      ).openPopup();
      markerRef.current = marker;

      setCurrentLocation(newLocation);

      // Sacuvaj na backend
      try {
        setLoading(true);
        await positionAPI.savePosition(lat, lng);
        setMessage('Position saved successfully');
        setTimeout(() => setMessage(null), 3000);
      } catch (error) {
        console.error('Error saving position:', error);
        setMessage('Failed to save position. Using local storage only.');
        setTimeout(() => setMessage(null), 3000);
      } finally {
        setLoading(false);
      }
    };

    map.on('click', handleMapClick);
    mapRef.current = map;

    return () => {
      map.off('click', handleMapClick);
      map.remove();
      mapRef.current = null;
    };
  }, []);

  return (
    <>
      <div className="tour-hero">
        <h1>Simulator</h1>
      </div>
      <div className="tour-page">
        <div
          ref={mapContainerRef}
          style={{
            width: '100%',
            height: '70vh',
            borderRadius: '8px',
            overflow: 'hidden',
            marginTop: '20px',
          }}
        />
      </div>
    </>
  );
}
