import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api, { purchaseAPI } from '../../api';
import './Tour.scss';

interface PurchaseToken {
  tourId: string;
  tourName: string;
  price: number;
  purchasedAt: string;
}

interface Tour {
  id: string;
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
  status: string;
  price: number;
  lengthKm?: number;
}

export default function MyPurchasedTours() {
  const navigate = useNavigate();
  const [tokens, setTokens] = useState<PurchaseToken[]>([]);
  const [tours, setTours] = useState<Record<string, Tour>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    purchaseAPI.getMyPurchases()
      .then(async (res) => {
        const list: PurchaseToken[] = Array.isArray(res.data) ? res.data : [];
        setTokens(list);

        const results = await Promise.allSettled(
          list.map((t) => api.get(`/tours/${t.tourId}`))
        );
        const map: Record<string, Tour> = {};
        results.forEach((r, i) => {
          if (r.status === 'fulfilled') {
            const t = r.value.data as Tour;
            map[list[i].tourId] = t;
          }
        });
        setTours(map);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p>Loading...</p>;

  return (
    <>
      <div className="tour-hero">
        <h1>My Purchased Tours</h1>
      </div>
      <div className="tour-page">
        {tokens.length === 0 ? (
          <p>You haven't purchased any tours yet. <span
            style={{ color: '#FA9819', cursor: 'pointer', textDecoration: 'underline' }}
            onClick={() => navigate('/tours/explore')}
          >Explore tours</span></p>
        ) : (
          tokens.map((token) => {
            const tour = tours[token.tourId];
            return (
              <div
                key={token.tourId}
                className="tour-card tour-card--purchased"
                onClick={() => navigate(`/tours/${token.tourId}`)}
              >
                <div className="tour-card-header">
                  <h3>{token.tourName}</h3>
                  <span className="tour-purchased-badge">Purchased</span>
                </div>
                {tour && (
                  <>
                    <p className="tour-meta">Difficulty: {tour.difficulty}</p>
                    <p>{tour.description}</p>
                    {tour.lengthKm !== undefined && tour.lengthKm > 0 && (
                      <p className="tour-meta">Length: {tour.lengthKm} km</p>
                    )}
                    <div className="tour-tags">
                      {(tour.tags ?? []).map((tag, i) => (
                        <span key={i}>#{tag}</span>
                      ))}
                    </div>
                  </>
                )}
                <p className="tour-meta" style={{ marginTop: 8, color: '#888', fontSize: '0.85rem' }}>
                  Purchased on {new Date(token.purchasedAt).toLocaleDateString()}
                </p>
                <div className="tour-card-actions">
                  <button
                    className="btn-primary"
                    style={{ padding: '6px 16px', fontSize: '0.9rem' }}
                    onClick={(e) => { e.stopPropagation(); navigate(`/tours/${token.tourId}`); }}
                  >
                    View Full Tour
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>
    </>
  );
}
