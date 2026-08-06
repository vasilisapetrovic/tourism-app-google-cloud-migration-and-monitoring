import axios from 'axios';

// Osnovni URL gatewaya, menja se preko VITE_API_URL u .env fajlu
export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_URL}/api`,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Position endpoints
export const positionAPI = {
  // Cuva poziciju trenutnog korisnika
  savePosition: (latitude: number, longitude: number) =>
    api.post('/position', { latitude, longitude }),

  // Vraca poziciju trenutnog korisnika
  getMyPosition: () => api.get('/position/me'),

  // Vraca poziciju određenog korisnika
  getUserPosition: (userId: number) => api.get(`/position/${userId}`),
};

// Cart endpoints
export const cartAPI = {
  getCart: () => api.get('/cart'),
  addItem: (tourId: string, tourName: string, price: number) =>
    api.post('/cart/items', { tourId, tourName, price }),
  removeItem: (tourId: string) => api.delete(`/cart/items/${tourId}`),
  checkout: () => api.post('/cart/checkout'),
};

// Purchase endpoints
export const purchaseAPI = {
  getMyPurchases: () => api.get('/purchases'),
};

export default api;