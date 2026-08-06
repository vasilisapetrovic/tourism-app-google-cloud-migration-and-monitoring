import { useState, useEffect } from 'react';
import { cartAPI } from '../../api';
import './Tour.scss';
import './ShoppingCart.scss';

interface OrderItem {
  tourId: string;
  tourName: string;
  price: number;
}

interface ShoppingCart {
  id: string;
  touristId: number;
  items: OrderItem[];
  totalPrice: number;
}

export default function ShoppingCart() {
  const [cart, setCart] = useState<ShoppingCart | null>(null);
  const [loading, setLoading] = useState(true);
  const [removing, setRemoving] = useState<string | null>(null);
  const [checkingOut, setCheckingOut] = useState(false);
  const [checkoutDone, setCheckoutDone] = useState(false);
  const [checkoutError, setCheckoutError] = useState('');

  useEffect(() => {
    cartAPI.getCart()
      .then((res) => setCart(res.data))
      .catch(() => setCart(null))
      .finally(() => setLoading(false));
  }, []);

  const handleRemove = async (tourId: string) => {
    setRemoving(tourId);
    try {
      const res = await cartAPI.removeItem(tourId);
      setCart(res.data);
    } catch {
      // ignore
    } finally {
      setRemoving(null);
    }
  };

  const handleCheckout = async () => {
    setCheckingOut(true);
    setCheckoutError('');
    try {
      await cartAPI.checkout();
      setCheckoutDone(true);
      setCart(null);
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.response?.data?.Error;
      setCheckoutError(msg || 'Checkout failed. Please try again.');
    } finally {
      setCheckingOut(false);
    }
  };

  return (
    <>
      <div className="tour-hero">
        <h1>Cart</h1>
      </div>
      <div className="tour-page">
        {loading ? (
          <p>Loading...</p>
        ) : checkoutDone ? (
          <div className="checkout-success">
            <p>Purchase complete! Your tours are now unlocked.</p>
          </div>
        ) : !cart || cart.items.length === 0 ? (
          <p>Your cart is empty.</p>
        ) : (
          <>
            <div className="cart-items">
              {cart.items.map((item) => (
                <div key={item.tourId} className="cart-item">
                  <div className="cart-item-info">
                    <h3>{item.tourName}</h3>
                    <span className="tour-price">
                      {item.price === 0 ? 'Free' : `$${item.price.toFixed(2)}`}
                    </span>
                  </div>
                  <button
                    className="cart-remove-btn"
                    onClick={() => handleRemove(item.tourId)}
                    disabled={removing === item.tourId}
                  >
                    {removing === item.tourId ? 'Removing...' : 'Remove'}
                  </button>
                </div>
              ))}
            </div>
            <div className="cart-summary">
              <div>
                <span className="cart-total-label">Total</span>
                <span className="cart-total-price" style={{ marginLeft: '16px' }}>
                  {cart.totalPrice === 0 ? 'Free' : `$${cart.totalPrice.toFixed(2)}`}
                </span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '8px' }}>
                <button
                  className="cart-checkout-btn"
                  onClick={handleCheckout}
                  disabled={checkingOut}
                >
                  {checkingOut ? 'Processing...' : 'Checkout'}
                </button>
                <p style={{
                  margin: 0,
                  fontSize: '0.85rem',
                  color: '#e74c3c',
                  minHeight: '1.2em',
                  textAlign: 'right',
                  maxWidth: '260px',
                }}>
                  {checkoutError}
                </p>
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
}
