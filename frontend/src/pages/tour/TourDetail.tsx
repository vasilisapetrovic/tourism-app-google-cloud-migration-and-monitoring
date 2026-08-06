import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api, { purchaseAPI } from '../../api';
import './Tour.scss';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import TourExecution from './TourExecution';

const createLocationIcon = (color: string = '#E74C3C') => {
  return L.divIcon({
    html: `
      <svg width="32" height="40" viewBox="0 0 32 40" xmlns="http://www.w3.org/2000/svg">
        <g>
          <path d="M16 0C8.268 0 2 6.268 2 14c0 7.732 14 26 14 26s14-18.268 14-26c0-7.732-6.268-14-14-14z" 
                fill="${color}" stroke="white" stroke-width="1"/>
          <circle cx="16" cy="14" r="5" fill="white"/>
        </g>
      </svg>
    `,
    iconSize: [32, 40],
    iconAnchor: [16, 40],
    popupAnchor: [0, -40],
    className: 'custom-marker'
  });
};

const createPendingIcon = (color: string = '#48749E') => {
  return L.divIcon({
    html: `
      <svg width="32" height="40" viewBox="0 0 32 40" xmlns="http://www.w3.org/2000/svg">
        <g opacity="0.8">
          <path d="M16 0C8.268 0 2 6.268 2 14c0 7.732 14 26 14 26s14-18.268 14-26c0-7.732-6.268-14-14-14z" 
                fill="${color}" stroke="#1E3D59" stroke-width="1" stroke-dasharray="3,3"/>
          <circle cx="16" cy="14" r="5" fill="white"/>
        </g>
      </svg>
    `,
    iconSize: [32, 40],
    iconAnchor: [16, 40],
    popupAnchor: [0, -40],
    className: 'pending-marker'
  });
};

interface TransportTime {
  transportType: string;
  minutes: number;
}

interface Tour {
  id: string;
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
  status: string;
  price: number;
  authorId: number;
  authorUsername: string;
  createdAt: string;
  publishedAt?: string;
  archivedAt?: string;
  lengthKm?: number;
  transportTimes?: TransportTime[];
}

interface Keypoint {
  id: string;
  tourId: string;
  name: string;
  description: string;
  imageUrl: string;
  lat: number;
  lng: number;
  createdAt: string;
}

interface Review {
  id: string;
  authorId: number;
  authorUsername: string;
  rating: number;
  comment: string;
  visitDate: string;
  images: string[];
  createdAt: string;
}

function getUserFromToken(): { id: number; username: string; role: string } | null {
  const token = localStorage.getItem('token');
  if (!token) return null;
  try {
    const parts = token.split('.');
    const padded = parts[1].replace(/-/g, '+').replace(/_/g, '/').padEnd(
      Math.ceil(parts[1].length / 4) * 4, '='
    );
    const payload = JSON.parse(atob(padded));
    return { id: Number(payload.id), username: String(payload.username), role: String(payload.role) };
  } catch {
    return null;
  }
}

const TRANSPORT_LABELS: Record<string, string> = {
  walking: '🚶 Walking',
  bicycle: '🚲 Bicycle',
  car: '🚗 Car',
};

const EMPTY_KP_FORM = { name: '', description: '', imageUrl: '', lat: 0, lng: 0 };

export default function TourDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [tour, setTour] = useState<Tour | null>(null);
  const [keypoints, setKeypoints] = useState<Keypoint[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [avgRating, setAvgRating] = useState(0);
  const [currentUser] = useState(getUserFromToken);
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);
  const [isPurchased, setIsPurchased] = useState(false);

  const [editingPrice, setEditingPrice] = useState(false);
  const [priceInput, setPriceInput] = useState('');
  const [priceError, setPriceError] = useState('');

  const [transportForm, setTransportForm] = useState({ transportType: 'walking', minutes: 60 });
  const [transportError, setTransportError] = useState('');
  const [showTransportForm, setShowTransportForm] = useState(false);

  const [kpForm, setKpForm] = useState(EMPTY_KP_FORM);
  const [editingKpId, setEditingKpId] = useState<string | null>(null);
  const [editingKpPosition, setEditingKpPosition] = useState(false);
  const [pendingLatLng, setPendingLatLng] = useState<{ lat: number; lng: number } | null>(null);
  const [kpError, setKpError] = useState('');
  const [kpImagePreview, setKpImagePreview] = useState('');
  const [showRoute, setShowRoute] = useState(true);

  const [reviewForm, setReviewForm] = useState({ rating: 5, comment: '', visitDate: new Date().toISOString().split('T')[0], images: [] as string[] });
  const [reviewImagePreviews, setReviewImagePreviews] = useState<string[]>([]);
  const [reviewError, setReviewError] = useState('');

  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markersRef = useRef<L.Layer[]>([]);
  const pendingMarkerRef = useRef<L.Marker | null>(null);
  const polylineRef = useRef<L.Polyline | null>(null);
  const canEditRef = useRef(false);
  const editingKpIdRef = useRef<string | null>(null);
  const editingKpPositionRef = useRef(false);

  const isAuthor = !!(tour && currentUser && tour.authorId === currentUser.id);
  const canEditKeypoints = isAuthor && tour?.status === 'draft';

  canEditRef.current = canEditKeypoints;
  editingKpIdRef.current = editingKpId;
  editingKpPositionRef.current = editingKpPosition;

  const fetchTour = () => {
    api.get(`/tours/${id}`)
      .then(res => {
        const t = res.data;
        setTour({
          ...t,
          tags: t.tags ?? [],
          transportTimes: t.transportTimes ?? [],
          price: t.price ?? 0,
          lengthKm: t.lengthKm ?? 0,
        });
      })
      .catch(() => {
        const role = currentUser?.role;
        navigate(role === 'tourist' ? '/tours/explore' : '/tours');
      });
  };

  const fetchKeypoints = () => {
    api.get(`/tours/${id}/keypoints`)
      .then(res => setKeypoints(res.data || []))
      .catch(() => {});
  };

  const fetchReviews = () => {
    api.get(`/tours/${id}/reviews`)
      .then(res => {
        setReviews(res.data.reviews || []);
        setAvgRating(res.data.averageRating || 0);
      })
      .catch(() => {});
  };

  const handleKeypointClick = (kp: Keypoint) => {
    if (mapRef.current) {
      mapRef.current.setView([kp.lat, kp.lng], 15, { animate: true, duration: 0.5 });
    }
  };

  useEffect(() => {
    fetchTour();
    fetchKeypoints();
    fetchReviews();
    if (currentUser?.role === 'tourist') {
      purchaseAPI.getMyPurchases()
        .then((res) => {
          const purchased = (res.data as { tourId: string }[]).some((t) => t.tourId === id);
          setIsPurchased(purchased);
        })
        .catch(() => {});
    }
  }, [id]);

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return;

    const map = L.map(mapContainerRef.current).setView([44.8176, 20.4569], 6);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
    }).addTo(map);

    map.on('click', (e: L.LeafletMouseEvent) => {
      if (!canEditRef.current) return;

      if (editingKpIdRef.current && editingKpPositionRef.current) {
        if (pendingMarkerRef.current) pendingMarkerRef.current.remove();
        const m = L.marker(e.latlng, { icon: createLocationIcon('#4CAF50') })
          .addTo(map).bindPopup('Updated key point position').openPopup();
        pendingMarkerRef.current = m;
        setKpForm(prev => ({ ...prev, lat: e.latlng.lat, lng: e.latlng.lng }));
        return;
      }

      if (editingKpIdRef.current || pendingLatLng) return;

      if (pendingMarkerRef.current) pendingMarkerRef.current.remove();
      const m = L.marker(e.latlng, { icon: createPendingIcon() })
        .addTo(map).bindPopup('New key point — fill the form below').openPopup();
      pendingMarkerRef.current = m;
      setPendingLatLng({ lat: e.latlng.lat, lng: e.latlng.lng });
      setKpForm(prev => ({ ...prev, lat: e.latlng.lat, lng: e.latlng.lng }));
      setKpError('');
    });

    mapRef.current = map;
    return () => { map.remove(); mapRef.current = null; };
  }, [tour]);

  useEffect(() => {
    if (!mapRef.current) return;
    markersRef.current.forEach(m => m.remove());
    markersRef.current = [];
    if (polylineRef.current) { polylineRef.current.remove(); polylineRef.current = null; }

    const visibleKps = (currentUser?.role === 'tourist' && !isPurchased)
      ? keypoints.slice(0, 1)
      : keypoints;

    if (visibleKps.length > 1 && showRoute) {
      const coords = visibleKps.map(kp => [kp.lat, kp.lng] as [number, number]);
      polylineRef.current = L.polyline(coords, {
        color: '#3498DB', weight: 4, opacity: 0.8, dashArray: '8, 4',
        className: 'tour-polyline'
      }).addTo(mapRef.current);
    }

    visibleKps.forEach(kp => {
      const marker = L.marker([kp.lat, kp.lng], { icon: createLocationIcon('#E74C3C') })
        .addTo(mapRef.current!).bindPopup(`<b>${kp.name}</b><br/>${kp.description}`);
      markersRef.current.push(marker);
    });

    if (visibleKps.length > 0) {
      const bounds = L.latLngBounds(visibleKps.map(kp => [kp.lat, kp.lng] as [number, number]));
      mapRef.current.fitBounds(bounds, { padding: [40, 40] });
    }
  }, [keypoints, tour, showRoute, isPurchased]);

  const handleKpImageFile = (file: File | null) => {
    if (!file) { setKpImagePreview(''); setKpForm(p => ({ ...p, imageUrl: '' })); return; }
    const reader = new FileReader();
    reader.onload = e => {
      const base64 = e.target?.result as string;
      setKpImagePreview(base64);
      setKpForm(p => ({ ...p, imageUrl: base64 }));
    };
    reader.readAsDataURL(file);
  };

  const clearPending = () => {
    if (pendingMarkerRef.current) { pendingMarkerRef.current.remove(); pendingMarkerRef.current = null; }
    setPendingLatLng(null);
    setKpForm(EMPTY_KP_FORM);
    setKpImagePreview('');
    setKpError('');
    setEditingKpPosition(false);
  };

  const handleAddKeypoint = async () => {
    setKpError('');
    if (!kpForm.name || !kpForm.description) { setKpError('Name and description are required'); return; }
    try {
      await api.post(`/tours/${id}/keypoints`, kpForm);
      clearPending();
      fetchKeypoints();
      fetchTour();
    } catch (err: any) {
      setKpError(err.response?.data?.error || 'Failed to add key point');
    }
  };

  const startEdit = (kp: Keypoint) => {
    clearPending();
    setEditingKpId(kp.id);
    setEditingKpPosition(false);
    setKpForm({ name: kp.name, description: kp.description, imageUrl: kp.imageUrl, lat: kp.lat, lng: kp.lng });
    setKpImagePreview(kp.imageUrl || '');
  };

  const startEditPosition = () => {
    setEditingKpPosition(true);
    if (pendingMarkerRef.current) pendingMarkerRef.current.remove();
    pendingMarkerRef.current = null;
  };

  const handleUpdateKeypoint = async (kpId: string) => {
    setKpError('');
    try {
      await api.put(`/tours/${id}/keypoints/${kpId}`, kpForm);
      setEditingKpId(null);
      setEditingKpPosition(false);
      setKpForm(EMPTY_KP_FORM);
      setKpImagePreview('');
      fetchKeypoints();
      fetchTour();
    } catch (err: any) {
      setKpError(err.response?.data?.error || 'Failed to update key point');
    }
  };

  const handleDeleteKeypoint = async (kpId: string) => {
    if (!window.confirm('Are you sure you want to delete this key point?')) return;
    try {
      await api.delete(`/tours/${id}/keypoints/${kpId}`);
      fetchKeypoints();
      fetchTour();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete key point');
    }
  };

  const handleReviewImages = (files: FileList | null) => {
    if (!files) return;
    Array.from(files).forEach(file => {
      const reader = new FileReader();
      reader.onload = e => {
        const base64 = e.target?.result as string;
        setReviewImagePreviews(prev => [...prev, base64]);
        setReviewForm(prev => ({ ...prev, images: [...prev.images, base64] }));
      };
      reader.readAsDataURL(file);
    });
  };

  const handlePublish = async () => {
    try {
      const res = await api.put(`/tours/${id}/publish`, {});
      setTour(res.data);
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to publish tour');
    }
  };

  const handleArchive = async () => {
    if (!window.confirm('Archive this tour?')) return;
    try {
      const res = await api.put(`/tours/${id}/archive`, {});
      setTour(res.data);
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to archive tour');
    }
  };

  const handleReactivate = async () => {
    try {
      const res = await api.put(`/tours/${id}/reactivate`, {});
      setTour(res.data);
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to reactivate tour');
    }
  };

  const handleAddTransportTime = async () => {
    setTransportError('');
    if (transportForm.minutes <= 0) { setTransportError('Minutes must be greater than 0'); return; }
    try {
      const res = await api.post(`/tours/${id}/transport-times`, transportForm);
      setTour(res.data);
      setShowTransportForm(false);
      setTransportForm({ transportType: 'walking', minutes: 60 });
    } catch (err: any) {
      setTransportError(err.response?.data?.error || 'Failed to add transport time');
    }
  };

  const handleUpdatePrice = async () => {
    setPriceError('');
    const price = priceInput === '' ? 0 : parseFloat(priceInput);
    if (isNaN(price) || price < 0) { setPriceError('Enter a valid positive number'); return; }
    try {
      const res = await api.patch(`/tours/${id}/price`, { price });
      setTour(res.data);
      setEditingPrice(false);
      setPriceInput('');
    } catch (err: any) {
      setPriceError(err.response?.data?.error || 'Failed to update price');
    }
  };

  const handleAddReview = async () => {
    setReviewError('');
    if (!reviewForm.comment.trim()) { setReviewError('Comment cannot be empty'); return; }
    try {
      await api.post(`/tours/${id}/reviews`, reviewForm);
      setReviewForm({ rating: 5, comment: '', visitDate: new Date().toISOString().split('T')[0], images: [] });
      setReviewImagePreviews([]);
      fetchReviews();
    } catch (err: any) {
      setReviewError(err.response?.data?.error || 'Failed to submit review');
    }
  };

  if (!tour) return <p>Loading...</p>;

  return (
    <>
      <div className="tour-hero">
        <h1>{tour.name}</h1>
      </div>
      <div className="tour-page">
        <button className="back-btn" onClick={() => navigate(currentUser?.role === 'guide' ? '/tours' : '/tours/explore')}>← Back</button>

        <div className="tour-detail-card">
          <div className="tour-detail-top">
            <div>
              <h2 style={{ margin: '0 0 12px 0' }}>{tour.name}</h2>
              <div className="tour-detail-badges">
                <span className={`tour-difficulty tour-difficulty--${tour.difficulty}`}>{tour.difficulty}</span>
                <span className={`tour-status-badge tour-status-badge--${tour.status}`}>{tour.status}</span>
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '10px' }}>
              {isAuthor && tour.status === 'draft' && editingPrice ? (
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '6px' }}>
                  <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                    <span style={{ fontWeight: 600, color: '#555' }}>$</span>
                    <input
                      type="number"
                      min="0"
                      step="0.01"
                      placeholder="0.00"
                      value={priceInput}
                      onChange={e => setPriceInput(e.target.value)}
                      style={{ width: '90px', padding: '4px 8px', borderRadius: '6px', border: '1px solid #ccc', fontSize: '0.95rem' }}
                      autoFocus
                    />
                    <button className="btn-primary" style={{ padding: '4px 12px' }} onClick={handleUpdatePrice}>Save</button>
                    <button className="btn-secondary" style={{ padding: '4px 10px' }} onClick={() => { setEditingPrice(false); setPriceError(''); }}>✕</button>
                  </div>
                  {priceError && <p className="form-error" style={{ margin: 0 }}>{priceError}</p>}
                </div>
              ) : (
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span className="tour-price">{tour.price === 0 ? '$0' : `$${tour.price}`}</span>
                  {isAuthor && tour.status === 'draft' && (
                    <button className="btn-small" onClick={() => { setEditingPrice(true); setPriceInput(String(tour.price)); }}>
                      Edit price
                    </button>
                  )}
                </div>
              )}

              {isAuthor && tour.status === 'draft' && (
                <button className="btn-primary" onClick={handlePublish}>Publish Tour</button>
              )}
              {isAuthor && tour.status === 'published' && (
                <button className="btn-archive" onClick={handleArchive}>Archive Tour</button>
              )}
              {isAuthor && tour.status === 'archived' && (
                <button className="btn-primary" onClick={handleReactivate}>Reactivate Tour</button>
              )}
            </div>
          </div>

          <p className="tour-detail-description">{tour.description}</p>

          <div className="tour-detail-section">
            <span className="tour-detail-label">Tags</span>
            <div className="tour-tags">
              {(tour.tags ?? []).map((tag, i) => <span key={i}>#{tag}</span>)}
            </div>
          </div>

          {(tour.lengthKm !== undefined && tour.lengthKm > 0) && (
            <div className="tour-detail-section">
              <span className="tour-detail-label">Length</span>
              <span className="tour-detail-value">{tour.lengthKm} km</span>
            </div>
          )}

          <div className="tour-detail-section">
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '8px' }}>
              <span className="tour-detail-label">Transport Times</span>
              {isAuthor && tour.status === 'draft' && (
                <button className="btn-small" onClick={() => setShowTransportForm(!showTransportForm)}>
                  {showTransportForm ? 'Cancel' : '+ Add'}
                </button>
              )}
            </div>

            {(tour.transportTimes ?? []).length > 0 ? (
              <div className="transport-times-list">
                {(tour.transportTimes ?? []).map((tt, i) => (
                  <div key={i} className="transport-time-item">
                    <span>{TRANSPORT_LABELS[tt.transportType] || tt.transportType}</span>
                    <span className="transport-time-minutes">{tt.minutes} min</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="tour-empty" style={{ margin: 0, fontSize: '0.9rem' }}>
                No transport times added yet.
                {isAuthor && tour.status === 'draft' && ' Add at least one before publishing.'}
              </p>
            )}

            {showTransportForm && (
              <div className="transport-form">
                <select
                  value={transportForm.transportType}
                  onChange={e => setTransportForm(p => ({ ...p, transportType: e.target.value }))}
                  className="kp-input"
                >
                  <option value="walking">🚶 Walking</option>
                  <option value="bicycle">🚲 Bicycle</option>
                  <option value="car">🚗 Car</option>
                </select>
                <input
                  type="number"
                  min={1}
                  placeholder="Minutes"
                  value={transportForm.minutes}
                  onChange={e => setTransportForm(p => ({ ...p, minutes: Number(e.target.value) }))}
                  className="kp-input"
                  style={{ width: '120px' }}
                />
                {transportError && <p className="form-error">{transportError}</p>}
                <button className="btn-primary" onClick={handleAddTransportTime}>Save</button>
              </div>
            )}
          </div>

          <div className="tour-detail-footer">
            <span className="tour-meta">Created: {tour.createdAt ? new Date(tour.createdAt).toLocaleDateString() : '—'}</span>
            {tour.publishedAt && (
              <span className="tour-meta">Published: {new Date(tour.publishedAt).toLocaleString()}</span>
            )}
            {tour.archivedAt && (
              <span className="tour-meta">Archived: {new Date(tour.archivedAt).toLocaleString()}</span>
            )}
          </div>
        </div>

        {currentUser?.role === 'tourist' && (
          <div className="tour-section">
            <TourExecution
              tourId={id!}
              tourStatus={tour.status}
              keypointsCount={keypoints.length}
              isPurchased={isPurchased}
            />
          </div>
        )}

        <div className="tour-section">
          <div className="tour-section-header">
            <h2 className="tour-section-title">Key Points</h2>
            {keypoints.length > 1 && (
              <button
                className={`tour-route-toggle ${showRoute ? 'route-visible' : 'route-hidden'}`}
                onClick={() => setShowRoute(!showRoute)}
                title={showRoute ? 'Hide route' : 'Show route'}
              >
                {showRoute ? (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                  </svg>
                ) : (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                )}
              </button>
            )}
          </div>



          {canEditKeypoints && !pendingLatLng && !editingKpId && (
            <p className="tour-section-hint">Click on the map to place a new key point.</p>
          )}
          {canEditKeypoints && editingKpPosition && (
            <p className="tour-section-hint" style={{ color: '#4CAF50' }}>Click on the map to set the new position.</p>
          )}

          <div className="tour-map-with-sidebar">
            <div className="keypoints-sidebar">
              <h3 className="sidebar-title">Key Points</h3>
              {keypoints.length === 0 ? (
                <p className="sidebar-empty">No key points</p>
              ) : (
                <ul className="keypoints-list">
                  {(currentUser?.role === 'tourist' && !isPurchased ? keypoints.slice(0, 1) : keypoints).map((kp) => (
                    <li key={kp.id} className="keypoint-item" onClick={() => handleKeypointClick(kp)}>
                      <span className="keypoint-name">{kp.name}</span>
                      {canEditKeypoints && (
                        <button className="keypoint-close-btn" onClick={(e) => { e.stopPropagation(); handleDeleteKeypoint(kp.id); }} title="Delete">✕</button>
                      )}
                    </li>
                  ))}
                  {currentUser?.role === 'tourist' && !isPurchased && keypoints.length > 1 && (
                    <li className="keypoint-item keypoint-locked">
                      <span>🔒︎ {keypoints.length - 1} more keypoint{keypoints.length - 1 !== 1 ? 's' : ''} — purchase to unlock</span>
                    </li>
                  )}
                </ul>
              )}
            </div>
            <div ref={mapContainerRef} className="tour-map" />
          </div>

          {canEditKeypoints && pendingLatLng && !editingKpId && (
            <div className="kp-form">
              <h3 className="kp-form-title">New Key Point</h3>
              <p className="kp-coords">Position: {kpForm.lat.toFixed(5)}, {kpForm.lng.toFixed(5)}</p>
              <input className="kp-input" placeholder="Name *" value={kpForm.name} onChange={e => setKpForm(p => ({ ...p, name: e.target.value }))} />
              <input className="kp-input" placeholder="Description *" value={kpForm.description} onChange={e => setKpForm(p => ({ ...p, description: e.target.value }))} />
              <label className="kp-file-label">Image<input type="file" accept="image/*" className="kp-file-input" onChange={e => handleKpImageFile(e.target.files?.[0] || null)} /></label>
              {kpImagePreview && <img src={kpImagePreview} alt="preview" className="kp-img-preview" />}
              {kpError && <p className="form-error">{kpError}</p>}
              <div className="kp-form-actions">
                <button className="btn-primary" onClick={handleAddKeypoint}>Add Key Point</button>
                <button className="btn-secondary" onClick={clearPending}>Cancel</button>
              </div>
            </div>
          )}

          {keypoints.length === 0 ? (
            <p className="tour-empty">No key points yet.</p>
          ) : (
            <div className="kp-list">
              {(currentUser?.role === 'tourist' && !isPurchased ? keypoints.slice(0, 1) : keypoints).map(kp => (
                <div key={kp.id} className="kp-card">
                  {editingKpId === kp.id ? (
                    <>
                      <h3 className="kp-form-title">Edit Key Point</h3>
                      <input className="kp-input" placeholder="Name *" value={kpForm.name} onChange={e => setKpForm(p => ({ ...p, name: e.target.value }))} />
                      <input className="kp-input" placeholder="Description *" value={kpForm.description} onChange={e => setKpForm(p => ({ ...p, description: e.target.value }))} />
                      <label className="kp-file-label">Image<input type="file" accept="image/*" className="kp-file-input" onChange={e => handleKpImageFile(e.target.files?.[0] || null)} /></label>
                      {kpImagePreview && <img src={kpImagePreview} alt="preview" className="kp-img-preview" />}
                      <p className="kp-coords" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span>Position: {kpForm.lat.toFixed(5)}, {kpForm.lng.toFixed(5)}</span>
                        {canEditKeypoints && (
                          <button className="kp-position-edit-btn" onClick={startEditPosition}>
                            {editingKpPosition ? 'Setting...' : 'Change on map'}
                          </button>
                        )}
                      </p>
                      {kpError && <p className="form-error">{kpError}</p>}
                      <div className="kp-form-actions">
                        <button className="btn-primary" onClick={() => handleUpdateKeypoint(kp.id)}>Save</button>
                        <button className="btn-secondary" onClick={() => { setEditingKpId(null); setEditingKpPosition(false); setKpForm(EMPTY_KP_FORM); setKpImagePreview(''); setKpError(''); }}>Cancel</button>
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="kp-card-header">
                        <h4 className="kp-name">{kp.name}</h4>
                        {canEditKeypoints && (
                          <div className="kp-actions">
                            <button className="kp-btn" onClick={() => startEdit(kp)}>Edit</button>
                            <button className="kp-btn kp-btn--danger" onClick={() => handleDeleteKeypoint(kp.id)}>Delete</button>
                          </div>
                        )}
                      </div>
                      <p className="kp-desc">{kp.description}</p>
                      {kp.imageUrl && <img className="kp-img" src={kp.imageUrl} alt={kp.name} onClick={() => setLightboxSrc(kp.imageUrl)} />}
                      <span className="kp-coords">Lat: {kp.lat.toFixed(5)}, Lng: {kp.lng.toFixed(5)}</span>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
          {currentUser?.role === 'tourist' && !isPurchased && keypoints.length > 1 && (
            <div className="kp-locked-banner">
              <span>🔒︎</span>
              <p>{keypoints.length - 1} more keypoint{keypoints.length - 1 !== 1 ? 's' : ''} hidden. Purchase this tour to see the full route.</p>
            </div>
          )}
        </div>

        <div className="tour-section">
          <div className="review-section-header">
            <h2 className="tour-section-title">Reviews</h2>
            {reviews.length > 0 && (
              <span className="review-avg">★ {avgRating.toFixed(1)} · {reviews.length} review{reviews.length !== 1 ? 's' : ''}</span>
            )}
          </div>

          {currentUser?.role === 'tourist' && (
            <div className="review-form">
              <h3 className="kp-form-title">Leave a Review</h3>
              <div className="review-rating-row">
                <label className="review-rating-label">Rating</label>
                <div className="star-picker">
                  {[1, 2, 3, 4, 5].map(n => (
                    <button key={n} className={`star-btn ${reviewForm.rating >= n ? 'star-btn--active' : ''}`} onClick={() => setReviewForm(p => ({ ...p, rating: n }))}>★</button>
                  ))}
                </div>
              </div>
              <textarea className="review-textarea" placeholder="Your comment..." value={reviewForm.comment} onChange={e => setReviewForm(p => ({ ...p, comment: e.target.value }))} />
              <input className="kp-input" type="date" value={reviewForm.visitDate} onChange={e => setReviewForm(p => ({ ...p, visitDate: e.target.value }))} />
              <label className="kp-file-label">Images (optional)<input type="file" accept="image/*" multiple className="kp-file-input" onChange={e => handleReviewImages(e.target.files)} /></label>
              {reviewImagePreviews.length > 0 && (
                <div className="review-img-previews">
                  {reviewImagePreviews.map((src, i) => <img key={i} src={src} alt={`preview-${i}`} className="kp-img-preview" />)}
                </div>
              )}
              {reviewError && <p className="form-error">{reviewError}</p>}
              <button className="btn-primary" onClick={handleAddReview}>Submit Review</button>
            </div>
          )}

          {reviews.length === 0 ? (
            <p className="tour-empty">No reviews yet. Be the first!</p>
          ) : (
            <div className="review-list">
              {reviews.map(review => (
                <div key={review.id} className="review-card">
                  <div className="review-header">
                    <span className="review-author">{review.authorUsername}</span>
                    <span className="review-stars">{'★'.repeat(review.rating)}{'☆'.repeat(5 - review.rating)}</span>
                    <span className="review-date">{new Date(review.createdAt).toLocaleDateString()}</span>
                  </div>
                  <p className="review-comment">{review.comment}</p>
                  {review.visitDate && <span className="review-visit">Visited: {new Date(review.visitDate).toLocaleDateString()}</span>}
                  {(review.images ?? []).length > 0 && (
                    <div className="review-imgs">
                      {(review.images ?? []).map((src, i) => <img key={i} src={src} alt={`review-img-${i}`} onClick={() => setLightboxSrc(src)} />)}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {lightboxSrc && (
        <div className="lightbox-overlay" onClick={() => setLightboxSrc(null)}>
          <img className="lightbox-img" src={lightboxSrc} alt="preview" onClick={e => e.stopPropagation()} />
        </div>
      )}
    </>
  );
}