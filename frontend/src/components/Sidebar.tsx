import { Link, useLocation, useNavigate } from 'react-router-dom';
import './Sidebar.scss';

export default function Sidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const role = JSON.parse(atob(localStorage.getItem('token')!.split('.')[1])).role;

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  const links = [
    { to: '/', label: 'Home', num: '01' },
    { to: '/blogs', label: 'Blogs', num: '02' },
    { to: '/profile', label: 'Profile', num: '03' },
    { to: '/users', label: 'Find Users', num: '04' },
    ...(role === 'admin' ? [{ to: '/user-management', label: 'Users', num: '05' }] : []),
    ...(role === 'guide' ? [{ to: '/tours', label: 'My Tours', num: '06' }] : []),
    ...(role === 'tourist' ? [{ to: '/tours/explore', label: 'Explore Tours', num: '06' }] : []),
    ...(role === 'tourist' ? [{ to: '/my-purchases', label: 'My Purchases', num: '07' }] : []),
    ...(role === 'tourist' ? [{ to: '/simulator', label: 'Simulator', num: '08' }] : []),
  ];

  return (
    <aside className="sidebar">
      <div className="sidebar-logo">Tourism</div>
      <nav className="sidebar-nav">
        {links.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className={`sidebar-link ${location.pathname === link.to ? 'active' : ''}`}
          >
            {link.label}<span className="sidebar-num">{link.num}</span>
          </Link>
        ))}
      </nav>
      <div className="sidebar-footer">
        <button onClick={handleLogout}>Logout</button>
      </div>
    </aside>
  );
}