import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api, { positionAPI } from '../../api';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import './ActiveExecution.scss';

interface CompletedKeypoint {
  keypointId: string;
  name: string;
  completedAt: string;
}

interface Execution {
  id: string;
  tourId: string;
  status: string;
  startedAt: string;
  endedAt?: string;
  currentLat: number;
  currentLng: number;
  completedKeypoints: CompletedKeypoint[];
}

interface Keypoint {
  id: string;
  name: string;
  description: string;
  imageUrl: string;
  lat: number;
  lng: number;
}

interface Tour {
  id: string;
  name: string;
  difficulty: string;
}

const POLL_MS = 10_000;
const RADIUS_M = 50;

function getUserFromToken() {
  const token = localStorage.getItem('token');
  if (!token) return null;
  try {
    const parts = token.split('.');
    const padded = parts[1].replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(parts[1].length / 4) * 4, '=');
    return JSON.parse(atob(padded));
  } catch {
    return null;
  }
}

const completedIcon = L.divIcon({
  html: `<svg width="36" height="44" viewBox="0 0 36 44" xmlns="http://www.w3.org/2000/svg">
    <path d="M18 0C9.163 0 2 7.163 2 16c0 8.837 16 28 16 28S34 24.837 34 16C34 7.163 26.837 0 18 0z"
          fill="#22c55e" stroke="white" stroke-width="2"/>
    <circle cx="18" cy="16" r="7" fill="white" opacity="0.25"/>
    <text x="18" y="21" text-anchor="middle" font-size="12" fill="white" font-weight="bold">✓</text>
  </svg>`,
  iconSize: [36, 44],
  iconAnchor: [18, 44],
  popupAnchor: [0, -46],
  className: '',
});

const pendingIcon = L.divIcon({
  html: `<svg width="36" height="44" viewBox="0 0 36 44" xmlns="http://www.w3.org/2000/svg">
    <path d="M18 0C9.163 0 2 7.163 2 16c0 8.837 16 28 16 28S34 24.837 34 16C34 7.163 26.837 0 18 0z"
          fill="#F45F00" stroke="white" stroke-width="2"/>
    <circle cx="18" cy="16" r="6" fill="white" opacity="0.3"/>
    <circle cx="18" cy="16" r="3" fill="white"/>
  </svg>`,
  iconSize: [36, 44],
  iconAnchor: [18, 44],
  popupAnchor: [0, -46],
  className: '',
});

const playerIcon = L.divIcon({
  html: `<div style="
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: linear-gradient(135deg, #F45F00 0%, #FA9819 60%, #C8602A 100%);
    border: 3px solid white;
    box-shadow: 0 2px 8px rgba(244, 95, 0, 0.8);
  "></div>`,
  iconSize: [20, 20],
  iconAnchor: [10, 10],
  className: 'player-marker',
});

export default function ActiveExecution() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [tour, setTour] = useState<Tour | null>(null);
  const [keypoints, setKeypoints] = useState<Keypoint[]>([]);
  const [execution, setExecution] = useState<Execution | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedKp, setSelectedKp] = useState<Keypoint | null>(null);
  const [flash, setFlash] = useState('');
  const [elapsedSec, setElapsedSec] = useState(0);

  const mapRef = useRef<L.Map | null>(null);
  const mapDivRef = useRef<HTMLDivElement>(null);
  const playerMarkerRef = useRef<L.Marker | null>(null);
  const kpMarkersRef = useRef<Map<string, L.Marker>>(new Map());
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const executionRef = useRef<Execution | null>(null);

  executionRef.current = execution;

  const completedIds = new Set(execution?.completedKeypoints.map(c => c.keypointId) ?? []);
  const progress = keypoints.length > 0 ? Math.round((completedIds.size / keypoints.length) * 100) : 0;

  const formatTime = (sec: number) => {
    const m = Math.floor(sec / 60).toString().padStart(2, '0');
    const s = (sec % 60).toString().padStart(2, '0');
    return `${m}:${s}`;
  };

  const updateMarkerIcon = useCallback((kpId: string, done: boolean) => {
    const marker = kpMarkersRef.current.get(kpId);
    if (marker) marker.setIcon(done ? completedIcon : pendingIcon);
  }, []);

  const sendPosition = useCallback(async (execId: string) => {
    try {
      // Preuzmi poziciju sa servera
      const posRes = await positionAPI.getMyPosition();
      console.log('Position response:', posRes.data);
      
      let lat = 44.8176;
      let lng = 20.4569;
      
      if (posRes.data) {
        const pos = posRes.data;
        // Koristi server poziciju samo ako je validna (nije 0,0)
        if ((pos.latitude !== 0 && pos.latitude !== undefined) || (pos.longitude !== 0 && pos.longitude !== undefined)) {
          lat = pos.latitude || 44.8176;
          lng = pos.longitude || 20.4569;
        }
      }

      console.log('Using position:', { lat, lng });

      // Ažuriraj marker samo ako je mapa inicijalizirana
      if (mapRef.current) {
        if (playerMarkerRef.current) {
          console.log('Updating existing marker to:', lat, lng);
          try {
            playerMarkerRef.current.setLatLng([lat, lng]);
          } catch (e) {
            console.error('Error updating marker:', e);
            playerMarkerRef.current = null;
          }
        }
        
        if (!playerMarkerRef.current) {
          console.log('Creating new marker at:', lat, lng);
          try {
            playerMarkerRef.current = L.marker([lat, lng], { icon: playerIcon, zIndexOffset: 1000 })
              .addTo(mapRef.current);
            console.log('Marker created successfully');
            // Pani ka marker lokaciji
            mapRef.current.panTo([lat, lng]);
          } catch (e) {
            console.error('Error creating marker:', e);
          }
        }
      } else {
        console.log('Map not ready yet');
      }

      // Pošalji poziciju na backend
      const res = await api.put(`/executions/${execId}/position`, { lat, lng });
      const { execution: updated, newlyCompleted: nc, allCompleted } = res.data;
      setExecution(updated);

      if (nc?.length > 0) {
        nc.forEach((k: CompletedKeypoint) => updateMarkerIcon(k.keypointId, true));
        setFlash(`Reached: ${nc.map((k: CompletedKeypoint) => k.name).join(', ')}`);
        setTimeout(() => setFlash(''), 4000);
      }

      if (allCompleted) {
        clearInterval(intervalRef.current!);
        clearInterval(timerRef.current!);
      }
    } catch (err) {
      console.error('Error sending position:', err);
    }
  }, [updateMarkerIcon]);

  const startPolling = useCallback((execId: string) => {
    if (intervalRef.current) clearInterval(intervalRef.current);
    
    console.log('Starting polling for execution:', execId);
    
    // Čekaj da se mapa inicijalizira prije prvog slanja pozicije
    const checkAndSend = () => {
      if (mapRef.current) {
        console.log('Map ready, sending first position');
        sendPosition(execId);
      } else {
        console.log('Waiting for map...');
        setTimeout(checkAndSend, 100);
      }
    };
    
    checkAndSend();
    intervalRef.current = setInterval(() => {
      console.log('Polling position update');
      sendPosition(execId);
    }, POLL_MS);
  }, [sendPosition]);

  useEffect(() => {
    const init = async () => {
      try {
        const [tourRes, kpRes, execRes, posRes] = await Promise.all([
          api.get(`/tours/${id}`),
          api.get(`/tours/${id}/keypoints`),
          api.get('/executions/my'),
          positionAPI.getMyPosition().catch(() => ({ data: { latitude: 44.8176, longitude: 20.4569 } })),
        ]);

        setTour(tourRes.data);
        const kps: Keypoint[] = kpRes.data || [];
        setKeypoints(kps);

        // Detektuj validnu lokaciju
        let startLat = 44.8176;
        let startLng = 20.4569;
        
        if (posRes.data) {
          const pos = posRes.data;
          // Koristi server poziciju samo ako je validna (nije 0,0)
          if ((pos.latitude !== 0 && pos.latitude !== undefined) || (pos.longitude !== 0 && pos.longitude !== undefined)) {
            startLat = pos.latitude || 44.8176;
            startLng = pos.longitude || 20.4569;
          }
        }

        const active: Execution = (execRes.data as Execution[]).find(
          e => e.tourId === id && e.status === 'active'
        ) ?? null as unknown as Execution;

        if (!active) {
          const startRes = await api.post(`/tours/${id}/executions`, {
            startLat,
            startLng,
          });
          setExecution(startRes.data);
          startPolling(startRes.data.id);

          timerRef.current = setInterval(() => setElapsedSec(s => s + 1), 1000);
        } else {
          setExecution(active);
          startPolling(active.id);
          const elapsed = Math.floor((Date.now() - new Date(active.startedAt).getTime()) / 1000);
          setElapsedSec(elapsed);
          timerRef.current = setInterval(() => setElapsedSec(s => s + 1), 1000);
        }
      } catch (e: any) {
        setError(e.response?.data?.error || 'Failed to start execution');
      } finally {
        setLoading(false);
      }
    };

    init();

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [id]);

  useEffect(() => {
    if (!mapDivRef.current || mapRef.current || keypoints.length === 0) return;

    const center: [number, number] = keypoints.length > 0
      ? [keypoints[0].lat, keypoints[0].lng]
      : [44.8176, 20.4569];

    const map = L.map(mapDivRef.current, { zoomControl: false }).setView(center, 14);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
      maxZoom: 19,
    }).addTo(map);

    L.control.zoom({ position: 'bottomright' }).addTo(map);

    if (keypoints.length > 1) {
      L.polyline(keypoints.map(kp => [kp.lat, kp.lng] as [number, number]), {
        color: '#F45F00',
        weight: 3,
        opacity: 0.5,
        dashArray: '8, 6',
      }).addTo(map);
    }

    const completedSet = new Set(executionRef.current?.completedKeypoints.map(c => c.keypointId) ?? []);

    keypoints.forEach(kp => {
      const icon = completedSet.has(kp.id) ? completedIcon : pendingIcon;
      const marker = L.marker([kp.lat, kp.lng], { icon })
        .addTo(map)
        .on('click', () => setSelectedKp(kp));
      kpMarkersRef.current.set(kp.id, marker);
    });

    const bounds = L.latLngBounds(keypoints.map(kp => [kp.lat, kp.lng] as [number, number]));
    map.fitBounds(bounds, { padding: [60, 60] });

    // NE kreiraj marker ovde - neka se kreira u sendPosition
    // playerMarkerRef će biti postavljen kada se preuzme pozicija sa servera

    mapRef.current = map;

    return () => {
      map.remove();
      mapRef.current = null;
    };
  }, [keypoints]);

  useEffect(() => {
    if (!execution) return;
    execution.completedKeypoints.forEach(ck => updateMarkerIcon(ck.keypointId, true));
  }, [execution?.completedKeypoints.length]);

  const handleAbandon = async () => {
    if (!execution || !window.confirm('Abandon this tour?')) return;
    try {
      await api.put(`/executions/${execution.id}/abandon`, {});
      clearInterval(intervalRef.current!);
      clearInterval(timerRef.current!);
      navigate(`/tours/${id}`);
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to abandon');
    }
  };

  const focusKeypoint = (kp: Keypoint) => {
    setSelectedKp(kp);
    mapRef.current?.flyTo([kp.lat, kp.lng], 16, { duration: 0.6 });
  };

  if (loading) return (
    <div className="exec-loading">
      <div className="exec-loading-spinner" />
      <p>Preparing your tour...</p>
    </div>
  );

  if (error) return (
    <div className="exec-loading">
      <p className="exec-error">{error}</p>
      <button onClick={() => navigate(`/tours/${id}`)}>← Back</button>
    </div>
  );

  if (execution?.status === 'completed') return (
    <div className="exec-loading">
      <div className="exec-complete-icon">🎉</div>
      <h2>Tour Complete!</h2>
      <p>{keypoints.length}/{keypoints.length} key points reached</p>
      <button className="exec-back-btn" onClick={() => navigate(`/tours/${id}`)}>← Back to Tour</button>
    </div>
  );

  return (
    <div className="exec-page">
      <div ref={mapDivRef} className="exec-map" />

      <div className="exec-topbar">
        <button className="exec-topbar-back" onClick={() => navigate(`/tours/${id}`)}>←</button>
        <div className="exec-topbar-info">
          <span className="exec-topbar-name">{tour?.name}</span>
          <span className="exec-topbar-time">{formatTime(elapsedSec)}</span>
        </div>
        <button className="exec-topbar-abandon" onClick={handleAbandon}>Abandon</button>
      </div>

      {flash && (
        <div className="exec-flash">
          <span>📍</span> {flash}
        </div>
      )}

      <div className="exec-panel">
        <div className="exec-progress-header">
          <span className="exec-progress-label">{completedIds.size}/{keypoints.length} key points</span>
          <span className="exec-progress-pct">{progress}%</span>
        </div>
        <div className="exec-progress-track">
          <div className="exec-progress-fill" style={{ width: `${progress}%` }} />
        </div>

        <ul className="exec-kp-list">
          {keypoints.map(kp => {
            const done = completedIds.has(kp.id);
            const entry = execution?.completedKeypoints.find(c => c.keypointId === kp.id);
            return (
              <li
                key={kp.id}
                className={`exec-kp-row ${done ? 'done' : ''} ${selectedKp?.id === kp.id ? 'selected' : ''}`}
                onClick={() => focusKeypoint(kp)}
              >
                <span className="exec-kp-icon">{done ? '✓' : '○'}</span>
                <span className="exec-kp-name">{kp.name}</span>
                {done && entry && (
                  <span className="exec-kp-time">{new Date(entry.completedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                )}
              </li>
            );
          })}
        </ul>

        <p className="exec-hint">Updates every 10s · {RADIUS_M}m proximity radius</p>
      </div>

      {selectedKp && (
        <div className="exec-kp-detail">
          <button className="exec-kp-detail-close" onClick={() => setSelectedKp(null)}>✕</button>
          {selectedKp.imageUrl && (
            <img className="exec-kp-detail-img" src={selectedKp.imageUrl} alt={selectedKp.name} />
          )}
          <div className="exec-kp-detail-body">
            <div className="exec-kp-detail-status">
              {completedIds.has(selectedKp.id)
                ? <span className="exec-badge-done">✓ Reached</span>
                : <span className="exec-badge-pending">Not yet reached</span>
              }
            </div>
            <h3 className="exec-kp-detail-name">{selectedKp.name}</h3>
            <p className="exec-kp-detail-desc">{selectedKp.description}</p>
            <p className="exec-kp-detail-coords">{selectedKp.lat.toFixed(5)}, {selectedKp.lng.toFixed(5)}</p>
          </div>
        </div>
      )}
    </div>
  );
}