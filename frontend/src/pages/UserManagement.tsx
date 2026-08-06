import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_URL } from '../api';
import './UserManagement.scss';

interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  firstName: string;
  lastName: string;
  blocked: boolean;
}

function UserManagement() {
  const [users, setUsers] = useState<User[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const fetchUsers = () => {
    const token = localStorage.getItem('token');
    fetch(`${API_URL}/api/user_management/users`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => {
        if (res.status === 403) throw new Error('Forbidden');
        return res.json();
      })
      .then((data) => setUsers(data ?? []))
      .catch(() => setError('You are not authorized to view this page.'));
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const toggleBlock = (id: number, currentlyBlocked: boolean) => {
    const token = localStorage.getItem('token');
    fetch(`${API_URL}/api/user_management/users/${id}/block`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ blocked: !currentlyBlocked }),
    }).then(() => fetchUsers());
  };

  const filteredUsers = users.filter((u) =>
    u.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
    u.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    u.firstName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    u.lastName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="um-page">
      <div className="um-container">
        <div className="um-header">
          <h2>User Management</h2>
          <button className="btn-back" onClick={() => navigate('/')}>Back</button>
        </div>

        {error && <p className="um-error">{error}</p>}

        <div className="um-search-box">
          <input
            type="text"
            placeholder="Search users by name, email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="um-search-input"
          />
        </div>

        <div className="um-table-wrapper">
          <table className="um-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Username</th>
                <th>Email</th>
                <th>First Name</th>
                <th>Last Name</th>
                <th>Role</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.map((u) => (
                <tr key={u.id}>
                  <td>{u.id}</td>
                  <td>{u.username}</td>
                  <td>{u.email}</td>
                  <td>{u.firstName || '—'}</td>
                  <td>{u.lastName || '—'}</td>
                  <td><span className={`badge badge-${u.role}`}>{u.role}</span></td>
                  <td><span className={`badge badge-${u.blocked ? 'blocked' : 'active'}`}>{u.blocked ? 'Blocked' : 'Active'}</span></td>
                  <td>
                    <button
                      className={`btn-block ${u.blocked ? 'btn-unblock' : ''}`}
                      onClick={() => toggleBlock(u.id, u.blocked)}
                    >
                      {u.blocked ? 'Unblock' : 'Block'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default UserManagement;