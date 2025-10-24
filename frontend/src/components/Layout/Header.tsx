import { useMemo, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button, Drawer, Menu, Typography, Avatar, Dropdown } from 'antd';
import {
    MenuOutlined,
    CloseOutlined,
    UserOutlined,
    LogoutOutlined,
    PlusOutlined,
    FileTextOutlined,
    AuditOutlined,
    CalendarOutlined,
    StarOutlined,
    HomeOutlined,
    ShopOutlined
} from '@ant-design/icons';
import { useAuth } from '../../hooks/useAuth';
import type { MenuProps } from 'antd';

const { Title, Text } = Typography;

const AppHeader = () => {
    const { user, logout } = useAuth();
    const location = useLocation();
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    const menuItems = useMemo(() => {
        const items: MenuProps['items'] = [
            {
                key: '/',
                label: <Link to="/">首页</Link>,
                icon: <HomeOutlined />
            },
            {
                key: '/popular',
                label: <Link to="/?sort=rating">热门点评</Link>,
                icon: <StarOutlined />
            },
            {
                key: '/latest',
                label: <Link to="/?sort=created_at">最新发布</Link>,
                icon: <CalendarOutlined />
            }
        ];

        if (user) {
            items.push(
                {
                    key: '/submit-store',
                    label: <Link to="/submit-store">店铺评价</Link>,
                    icon: <ShopOutlined />
                },
                {
                    key: '/my',
                    label: <Link to="/my">我的点评</Link>,
                    icon: <FileTextOutlined />
                }
            );
        }

        if (user?.role === 'admin') {
            items.push({
                key: '/admin/reviews',
                label: <Link to="/admin/reviews">审核中心</Link>,
                icon: <AuditOutlined />
            });
        }
        return items;
    }, [user]);

    const selectedKey = location.pathname.startsWith('/admin')
        ? '/admin/reviews'
        : location.pathname.startsWith('/submit-store')
            ? '/submit-store'
            : location.pathname.startsWith('/submit')
                ? '/submit'
                : location.pathname.startsWith('/my')
                    ? '/my'
                    : '/';

    const MobileMenu = () => (
        <Drawer
            title="菜单"
            placement="left"
            closable={false}
            onClose={() => setMobileMenuOpen(false)}
            open={mobileMenuOpen}
            width={280}
            styles={{ body: { padding: 0 } }}
        >
            <div style={{ padding: 16 }}>
                <Menu
                    mode="vertical"
                    selectedKeys={[selectedKey]}
                    items={menuItems}
                    style={{ border: 'none' }}
                />
            </div>
        </Drawer>
    );

    return (
        <>
            <MobileMenu />

            <header style={{
                background: '#fff',
                borderBottom: '1px solid #e5e7eb',
                position: 'sticky',
                top: 0,
                zIndex: 1000,
                backdropFilter: 'blur(8px)',
                backgroundColor: 'rgba(255, 255, 255, 0.95)',
            }}>
                <div style={{
                    maxWidth: 1200,
                    margin: '0 auto',
                    padding: '0 16px',
                    height: 64,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                }}>
                    {/* 移动端菜单按钮 */}
                    <Button
                        type="text"
                        icon={<MenuOutlined />}
                        onClick={() => setMobileMenuOpen(true)}
                        className="mobile-menu-btn"
                        style={{ display: 'none' }}
                    />

                    {/* Logo */}
                    <Link to="/" style={{ textDecoration: 'none' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <Title level={3} style={{ margin: 0, color: '#2563eb', fontSize: 24 }}>
                                杭电点评
                            </Title>
                        </div>
                    </Link>

                    {/* 桌面端导航 */}
                    <nav className="desktop-nav">
                        <Menu
                            mode="horizontal"
                            selectedKeys={[selectedKey]}
                            items={menuItems}
                            style={{
                                border: 'none',
                                background: 'transparent',
                                lineHeight: '64px',
                                flex: 1
                            }}
                        />
                    </nav>

                    {/* 用户区域 */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        {user ? (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                <Avatar size={32} icon={<UserOutlined />} style={{ backgroundColor: '#2563eb' }} />
                                <Text>{user.display_name}</Text>
                                <Link to="/my-profile">
                                    <Button type="link">个人主页</Button>
                                </Link>
                                <Button type="link" onClick={logout}>退出登录</Button>
                            </div>
                        ) : (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                <Link to="/login">
                                    <Button type="text">登录</Button>
                                </Link>
                                <Link to="/register">
                                    <Button type="primary">注册</Button>
                                </Link>
                            </div>
                        )}
                    </div>
                </div>
            </header>

            <style>{`
        .mobile-menu-btn {
          display: none !important;
        }
        
        @media (max-width: 768px) {
          .mobile-menu-btn {
            display: block !important;
          }
        }
      `}</style>
        </>
    );
};

export default AppHeader;