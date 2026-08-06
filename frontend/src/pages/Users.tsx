import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api, { API_URL } from '../api';
import './Users.scss';

interface User {
  id: number;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  profilePicture?: string;
}

interface RecommendedUser {
  id: string;
  username?: string;
}

function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [recommendations, setRecommendations] = useState<RecommendedUser[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  const [following, setFollowing] = useState<Set<number>>(new Set());
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);
  const [loadingId, setLoadingId] = useState<number | null>(null);
  const [followingLoaded, setFollowingLoaded] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      const decoded = JSON.parse(atob(token.split('.')[1]));
      setCurrentUserId(decoded.id);
    }
  }, []);

  const fetchUsers = () => {
    const token = localStorage.getItem('token');
    fetch(`${API_URL}/api/profile/api/users`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(res => { if (res.status === 403) throw new Error('Forbidden'); return res.json(); })
      .then(data => setUsers(data ?? []))
      .catch(() => setError('Failed to load users.'));
  };

  useEffect(() => { fetchUsers(); }, []);

  const loadFollowingList = () => {
    if (!currentUserId || followingLoaded) return;
    api.get(`/following/${currentUserId}`)
      .then(res => {
        const ids = new Set<number>();
        if (Array.isArray(res.data)) {
          res.data.forEach((item: any) => {
            const id = item.id || item.ID;
            if (id) {
              const n = typeof id === 'string' ? parseInt(id) : id;
              if (!isNaN(n)) ids.add(n);
            }
          });
        }
        setFollowing(ids);
        setFollowingLoaded(true);
      })
      .catch(() => setFollowingLoaded(true));
  };

  useEffect(() => { loadFollowingList(); }, [currentUserId]);

  useEffect(() => {
    if (!currentUserId) return;
    api.get(`/recommendations/${currentUserId}`)
      .then(res => setRecommendations(res.data || []))
      .catch(() => setRecommendations([]));
  }, [currentUserId, following]);

  const filteredUsers = users.filter(u =>
    (u.username || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
    (u.email || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
    (u.firstName || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
    (u.lastName || '').toLowerCase().includes(searchQuery.toLowerCase())
  );

  const toggleFollow = (userId: number, username: string) => {
    if (!currentUserId || loadingId !== null) return;
    const isFollowing = following.has(userId);
    const newFollowing = new Set(following);
    isFollowing ? newFollowing.delete(userId) : newFollowing.add(userId);
    setFollowing(newFollowing);
    const currentUser = users.find(u => u.id === currentUserId);
    setLoadingId(userId);
    api({
      method: isFollowing ? 'DELETE' : 'POST',
      url: isFollowing ? '/unfollow' : '/follow',
      data: {
        FollowerID: String(currentUserId),
        FollowedID: String(userId),
        FollowerUsername: currentUser?.username || '',
        FollowedUsername: username,
      },
    })
      .then(() => { setError(''); setFollowingLoaded(false); })
      .catch(() => {
        const revert = new Set(following);
        isFollowing ? revert.add(userId) : revert.delete(userId);
        setFollowing(revert);
      })
      .finally(() => setLoadingId(null));
  };

  const UserCard = ({ u, id, username }: { u?: User; id: number; username?: string }) => {
    const displayUsername = username?.trim() || 'Unknown user';
    const avatarInitial = displayUsername[0]?.toUpperCase() || '?';

    return (
      <div className="user-card">
        <div className="user-avatar">
          {u?.profilePicture
            ? <img src={u.profilePicture} alt={displayUsername} />
            : <div className="avatar-placeholder">{avatarInitial}</div>
          }
        </div>
        <div className="user-info">
          <h3>{displayUsername}</h3>
          {u && <p className="user-name">{u.firstName} {u.lastName}</p>}
          {u && <p className="user-email">{u.email}</p>}
        </div>
        {currentUserId !== id && (
          <button
            className={`btn-follow ${following.has(id) ? 'btn-following' : ''}`}
            onClick={() => toggleFollow(id, displayUsername)}
            disabled={loadingId === id}
          >
            {loadingId === id ? '...' : following.has(id) ? 'Following' : 'Follow'}
          </button>
        )}
      </div>
    );
  };

  return (
    <>
      <div className="users-hero">
        <h1>Find Users</h1>
      </div>
      <div className="users-page">
        {error && <p className="users-error">{error}</p>}

        <div className="users-search-box">
          <input
            type="text"
            placeholder="Search by name or email..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="users-search-input"
          />
        </div>

        <div className="users-grid">
          {filteredUsers.length > 0
            ? filteredUsers.map(u => <UserCard key={u.id} u={u} id={u.id} username={u.username} />)
            : <div className="no-users">No users found</div>
          }
        </div>

        {recommendations.length > 0 && (
          <div className="recommendations-section">
            <h3>Recommended for you</h3>
            <div className="users-grid">
              {recommendations.map(u => (
                <UserCard key={u.id} id={Number(u.id)} username={u.username} />
              ))}
            </div>
          </div>
        )}
      </div>
    </>
  );
}

export default Users;