import { Layout } from 'antd';
import { Route, Routes } from 'react-router-dom';
import NavBar from './components/NavBar';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import SubmitReview from './pages/SubmitReview';
import MyReviews from './pages/MyReviews';
import ReviewDetail from './pages/ReviewDetail';
import AdminPending from './pages/AdminPending';

const { Header, Content } = Layout;

const App = () => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ background: '#fff' }}>
        <NavBar />
      </Header>
      <Content style={{ padding: '24px 48px' }}>
        <main>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/reviews/:id" element={<ReviewDetail />} />

            <Route element={<ProtectedRoute />}>
              <Route path="/submit" element={<SubmitReview />} />
              <Route path="/my" element={<MyReviews />} />
            </Route>

            <Route element={<ProtectedRoute requireAdmin />}>
              <Route path="/admin/reviews" element={<AdminPending />} />
            </Route>
          </Routes>
        </main>
      </Content>
    </Layout>
  );
};

export default App;
