import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import   api, { API_URL }  from '../api';
import './Auth.scss';
import './Profile.scss';

interface ProfileData {
  id: number;
  username: string;
  email: string;
  role: string;
  firstName: string;
  lastName: string;
  profilePicture: string;
  biography: string;
  motto: string;
}

interface User {
  id: string;
  username: string;
}

export default function Profile() {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [following, setFollowing] = useState<User[]>([]);
  const [editing, setEditing] = useState(false);
  const [form, setForm] = useState({ firstName: '', lastName: '', biography: '', motto: '' });
  const [pictureFile, setPictureFile] = useState<File | null>(null);
  const [picturePreview, setPicturePreview] = useState<string>('');
  const [error, setError] = useState('');

  useEffect(() => {
    api.get('/profile')
      .then(res => {
        setProfile(res.data);
        setForm({
          firstName: res.data.firstName || '',
          lastName: res.data.lastName || '',
          biography: res.data.biography || '',
          motto: res.data.motto || '',
        });
        api.get(`/following/${res.data.id}`)
          .then(res => setFollowing(res.data || []))
          .catch(() => setFollowing([]));
      })
      .catch(() => navigate('/login'));
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      if (pictureFile) {
        const formData = new FormData();
        formData.append('picture', pictureFile);
        const token = localStorage.getItem('token');
        await fetch(`${API_URL}/api/profile/picture`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
          body: formData,
        });
      }
      const res = await api.put('/profile', form);
      setProfile(res.data);
      setEditing(false);
      setPictureFile(null);
      const updated = await api.get('/profile');
      setProfile(updated.data);
    } catch {
      setError('Failed to update profile.');
    }
  };

  if (!profile) return <p>Loading...</p>;

  const initials = `${profile.firstName?.[0] || profile.username?.[0] || '?'}${profile.lastName?.[0] || ''}`.toUpperCase();

  return (
    <>
      <div className="profile-hero">
        <h1>My Profile</h1>
      </div>

      <div className="profile-page">
        <div className="profile-card">
          {error && <p className="auth-error">{error}</p>}

          <div className="profile-view">
            <div className="profile-left">
              {(picturePreview || profile.profilePicture) ? (
                <img src={picturePreview || profile.profilePicture} alt="profile" className="profile-avatar-img" />
              ) : (
                <div className="profile-avatar-initials">{initials}</div>
              )}
              <span className={`profile-role-badge role-${profile.role}`}>{profile.role}</span>

              <div className="profile-stats">
                <div className="profile-stat">
                  <span className="profile-stat-number">{following.length}</span>
                  <span className="profile-stat-label">Following</span>
                </div>
              </div>

              {!editing && (
                <button className="btn-edit-profile" onClick={() => setEditing(true)}>Edit Profile</button>
              )}
            </div>

            {!editing ? (
              <div className="profile-fields">
                <div className="profile-field">
                  <span className="profile-field-label">Username</span>
                  <span className="profile-field-value">{profile.username}</span>
                </div>
                <div className="profile-field">
                  <span className="profile-field-label">Email</span>
                  <span className="profile-field-value">{profile.email}</span>
                </div>
                <div className="profile-field">
                  <span className="profile-field-label">First Name</span>
                  <span className="profile-field-value">{profile.firstName || '—'}</span>
                </div>
                <div className="profile-field">
                  <span className="profile-field-label">Last Name</span>
                  <span className="profile-field-value">{profile.lastName || '—'}</span>
                </div>
                <div className="profile-field">
                  <span className="profile-field-label">Biography</span>
                  <span className="profile-field-value">{profile.biography || '—'}</span>
                </div>
                <div className="profile-field profile-field--motto">
                  <span className="profile-field-label">Motto</span>
                  <span className="profile-field-value">{profile.motto ? `"${profile.motto}"` : '—'}</span>
                </div>

               
      
              </div>
            ) : (
              <form className="profile-fields" onSubmit={handleSave}>
                <div className="form-group">
                  <label>First Name</label>
                  <input value={form.firstName} onChange={e => setForm({ ...form, firstName: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>Last Name</label>
                  <input value={form.lastName} onChange={e => setForm({ ...form, lastName: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>Profile Picture</label>
                  <input type="file" accept="image/*" onChange={e => {
                    const file = e.target.files?.[0] || null;
                    setPictureFile(file);
                    setPicturePreview(file ? URL.createObjectURL(file) : '');
                  }} />
                </div>
                <div className="form-group">
                  <label>Biography</label>
                  <textarea value={form.biography} onChange={e => setForm({ ...form, biography: e.target.value })} rows={3} style={{ width: '100%' }} />
                </div>
                <div className="form-group">
                  <label>Motto</label>
                  <input value={form.motto} onChange={e => setForm({ ...form, motto: e.target.value })} />
                </div>
                <div className="profile-edit-actions">
                  <button className="btn" type="submit">Save</button>
                  <button className="btn-cancel-profile" type="button" onClick={() => { setEditing(false); setPictureFile(null); setPicturePreview(''); }}>Cancel</button>
                </div>
              </form>
            )}
          </div>
        </div>
      </div>
    </>
  );
}