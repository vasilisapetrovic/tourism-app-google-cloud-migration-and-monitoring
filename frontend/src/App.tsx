import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Profile from './pages/Profile';
import UserManagement from './pages/UserManagement';
import Users from './pages/Users';
import Blogs from './pages/blog/Blog';
import CreateBlog from './pages/blog/CreateBlog';
import Sidebar from './components/Sidebar';
import BlogDetail from './pages/blog/BlogDetail';
import Tours from './pages/tour/Tours';
import CreateTour from './pages/tour/CreateTour';
import TourDetail from './pages/tour/TourDetail';
import ExploreTours from './pages/tour/ExploreTours';
import Simulator from './pages/tour/Simulator';
import ActiveExecution from './pages/tour/ActiveExecution';
import ShoppingCart from './pages/tour/ShoppingCart';
import MyPurchasedTours from './pages/tour/MyPurchasedTours';

function ProtectedRoute() {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/login" replace />;
  return (
    <div className="layout">
      <Sidebar />
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<Home />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/user-management" element={<UserManagement />} />
          <Route path="/users" element={<Users />} />
          <Route path="/blogs" element={<Blogs />} />
          <Route path="/blogs/create" element={<CreateBlog />} />
          <Route path="/blogs/:id" element={<BlogDetail />} />
          <Route path="/tours" element={<Tours />} />
          <Route path="/tours/create" element={<CreateTour />} />
          <Route path="/tours/:id" element={<TourDetail />} />
          <Route path="/tours/explore" element={<ExploreTours />} />
          <Route path="/simulator" element={<Simulator />} />
          <Route path="/tours/:id/execute" element={<ActiveExecution />} />
          <Route path="/cart" element={<ShoppingCart />} />
          <Route path="/my-purchases" element={<MyPurchasedTours />} />
        </Route>
        <Route path="*" element={<Navigate to="/login" />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;