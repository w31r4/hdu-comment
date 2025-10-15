import { useMemo } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button, Menu, Space, Typography } from 'antd';
import type { MenuProps } from 'antd';
import { useAuth } from '../hooks/useAuth';

const NavBar = () => {
  const { user, logout } = useAuth();
  const location = useLocation();

  const menuItems = useMemo<MenuProps['items']>(() => {
    const items: MenuProps['items'] = [
      { key: '/', label: <Link to="/">首页</Link> }
    ];
    if (user) {
      items.push({ key: '/submit', label: <Link to="/submit">提交点评</Link> });
      items.push({ key: '/my', label: <Link to="/my">我的点评</Link> });
    }
    if (user?.role === 'admin') {
      items.push({ key: '/admin/reviews', label: <Link to="/admin/reviews">审核中心</Link> });
    }
    return items;
  }, [user]);

  const selectedKey = useMemo(() => {
    if (location.pathname.startsWith('/admin')) return '/admin/reviews';
    if (location.pathname.startsWith('/submit')) return '/submit';
    if (location.pathname.startsWith('/my')) return '/my';
    return '/';
  }, [location.pathname]);

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 24 }}>
      <Typography.Title level={4} style={{ margin: 0 }}>
        <Link to="/" style={{ color: 'inherit' }}>
          杭电食物点评
        </Link>
      </Typography.Title>
      <Menu
        mode="horizontal"
        selectedKeys={[selectedKey]}
        items={menuItems}
        style={{ flex: 1, marginLeft: 24 }}
      />
      <Space size="middle">
        {user ? (
          <>
            <Typography.Text>你好，{user.display_name}</Typography.Text>
            <Button onClick={() => { void logout(); }}>退出</Button>
          </>
        ) : (
          <>
            <Button type="link">
              <Link to="/login">登录</Link>
            </Button>
            <Button type="primary">
              <Link to="/register">注册</Link>
            </Button>
          </>
        )}
      </Space>
    </div>
  );
};

export default NavBar;
