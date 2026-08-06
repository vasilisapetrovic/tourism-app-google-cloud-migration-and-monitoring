import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api, { cartAPI, purchaseAPI } from '../../api';
import './Tour.scss';

interface Tour {
  id: string;
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
  status: string;
  price: number;
  createdAt: string;
}

export default function ExploreTours() {
  const [tours, setTours] = useState<Tour[]>([]);
  const [cartIds, setCartIds] = useState<Set<string>>(new Set());
  const [purchasedIds, setPurchasedIds] = useState<Set<string>>(new Set());
  const [adding, setAdding] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    cartAPI.getCart()
      .then((res) => {
        const ids = (res.data.items as { tourId: string }[]).map((i) => i.tourId);
        setCartIds(new Set(ids));
      })
      .catch(() => {});

    purchaseAPI.getMyPurchases()
      .then((res) => {
        const ids = (res.data as { tourId: string }[]).map((t) => t.tourId);
        setPurchasedIds(new Set(ids));
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    api.get('/tours/published')
      .then((res) => setTours(res.data || []))
      .catch(() => setTours([]));
  }, []);

  const handleAddToCart = async (e: React.MouseEvent, tour: Tour) => {
    e.stopPropagation();
    setAdding(tour.id);
    try {
      await cartAPI.addItem(tour.id, tour.name, tour.price);
      setCartIds((prev) => new Set(prev).add(tour.id));
    } catch (err: any) {
      const msg = err?.response?.data?.error;
      if (msg) alert(msg);
    } finally {
      setAdding(null);
    }
  };

  const getCardAction = (tour: Tour) => {
    if (purchasedIds.has(tour.id)) {
      return (
        <span className="tour-purchased-badge">Purchased</span>
      );
    }
    return (
      <button
        className={`cart-add-btn${cartIds.has(tour.id) ? ' in-cart' : ''}`}
        onClick={(e) => handleAddToCart(e, tour)}
        disabled={cartIds.has(tour.id) || adding === tour.id}
      >
        {cartIds.has(tour.id) ? 'In Cart' : adding === tour.id ? 'Adding...' : 'Add to Cart'}
      </button>
    );
  };

  return (
    <>
      <div className="tour-hero">
        <h1>Explore Tours</h1>
      </div>
      <div className="tour-page">
        <div className="tour-header">
          <button className="cart-icon-btn" onClick={() => navigate('/cart')} title="My Cart">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
            {cartIds.size > 0 && <span className="cart-badge">{cartIds.size}</span>}
          </button>
        </div>
        {tours.length === 0 ? (
          <p>No published tours yet.</p>
        ) : (
          tours.map((tour) => (
            <div
              key={tour.id}
              className={`tour-card${purchasedIds.has(tour.id) ? ' tour-card--purchased' : ''}`}
              onClick={() => navigate(`/tours/${tour.id}`)}
            >
              <div className="tour-card-header">
                <h3>{tour.name}</h3>
                <span className="tour-price">{tour.price === 0 ? 'Free' : `$${tour.price}`}</span>
              </div>
              <p className="tour-meta">Difficulty: {tour.difficulty}</p>
              <p>{tour.description}</p>
              <div className="tour-tags">
                {tour.tags.map((tag, i) => (
                  <span key={i}>#{tag}</span>
                ))}
              </div>
              <div className="tour-card-actions">
                {getCardAction(tour)}
              </div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
