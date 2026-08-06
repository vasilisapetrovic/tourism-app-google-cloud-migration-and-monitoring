import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../api';
import './Tour.scss';

export default function CreateTour() {
  const navigate = useNavigate();
  const [form, setForm] = useState({
    name: '',
    description: '',
    difficulty: 'easy',
    tags: '',
  });
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  setError('');

  if (!form.tags.trim()) {
    setError('Please enter at least one tag');
    return;
  }

  try {
    await api.post('/tours', {
      ...form,
      tags: form.tags.split(',').map((t) => t.trim()).filter((t) => t !== ''),
    });
    navigate('/tours');
  } catch {
    setError('Failed to create tour');
  }
};

  return (
    <div className="blog-page" style={{ paddingTop: '40px' }}>
      <h2>Create Tour</h2>
      {error && <p className="auth-error">{error}</p>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Name</label>
          <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        </div>
        <div className="form-group">
          <label>Description</label>
          <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={5} required />
        </div>
        <div className="form-group">
          <label>Difficulty</label>
          <select value={form.difficulty} onChange={(e) => setForm({ ...form, difficulty: e.target.value })}>
            <option value="easy">Easy</option>
            <option value="medium">Medium</option>
            <option value="hard">Hard</option>
          </select>
        </div>
        <div className="form-group">
          <label>Tags (comma separated)</label>
          <input value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} placeholder="priroda, planina, jezero" />
        </div>
        <button className="btn" type="submit">Create Tour</button>
      </form>
    </div>
  );
}