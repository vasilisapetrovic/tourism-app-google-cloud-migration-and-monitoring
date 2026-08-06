import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../../api';
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

export default function Tours() {
  const [tours, setTours] = useState<Tour[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
  api.get('/tours/my')
    .then((res) => setTours(res.data || []))
    .catch(() => setTours([]));
}, []);

  return (
    <>
      <div className="tour-hero">
        <h1>My Tours</h1>
      </div>
      <div className="tour-page">
        <div className="tour-header">
          <Link to="/tours/create" className="btn">New Tour</Link>
        </div>
        {tours.length === 0 ? (
          <p>No tours yet.</p>
        ) : (
          tours.map((tour) => (
            <div key={tour.id} className="tour-card" onClick={() => navigate(`/tours/${tour.id}`)}>
                <div className="tour-card-header">
                    <h3>{tour.name}</h3>
                    <span className="tour-price">{tour.price === 0 ? '$0' : `$${tour.price}`}</span>
                </div>
                <p className="tour-meta">
                    <span className={`tour-difficulty tour-difficulty--${tour.difficulty} `}>{tour.difficulty}</span>
                    <span> · Status: {tour.status} </span>
                </p>
                <p className="tour-description">{tour.description}</p>
                <div className="tour-tags">
                    {tour.tags.map((tag, i) => (
                    <span key={i}>#{tag}</span>
                    ))}
                </div>
            </div>
          ))
        )}
      </div>
    </>
  );
}