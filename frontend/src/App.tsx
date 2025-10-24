import { Route, Routes } from 'react-router-dom';
import AppHeader from './components/Layout/Header';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import SubmitStoreReview from './pages/SubmitStoreReview';
import MyReviews from './pages/MyReviews';
import MyProfile from './pages/MyProfile';
import ReviewDetail from './pages/ReviewDetail';
import StoreDetail from './pages/StoreDetail';
import AdminPending from './pages/AdminPending';
import NotFound from './pages/NotFound';
import './styles/global.css';
import './styles/home.css';
import './styles/responsive.css';
import './styles/lazy-image.css';

const App = () => {
  return (
    <div className="app-container">
      <AppHeader />
      <main className="app-main">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/reviews/:id" element={<ReviewDetail />} />
          <Route path="/stores/:id" element={<StoreDetail />} />

          <Route element={<ProtectedRoute />}>
            <Route path="/submit-review" element={<SubmitStoreReview />} />
            <Route path="/my" element={<MyReviews />} />
            <Route path="/my-profile" element={<MyProfile />} />
          </Route>

          <Route element={<ProtectedRoute requireAdmin />}>
            <Route path="/admin/reviews" element={<AdminPending />} />
           </Route>
           {/* Fallback route for 404 Not Found */}
           <Route path="*" element={<NotFound />} />
          </Routes>
         </main>
        </div>
       );
      };

export default App;
