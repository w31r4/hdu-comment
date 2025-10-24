import { useEffect, useState } from 'react';
import { Card, Typography, Spin, Alert, Descriptions } from 'antd';
import { useAuth } from '../hooks/useAuth';
import { User } from '../types';
import { fetchMe } from '../api/client';

const { Title } = Typography;

const MyProfile = () => {
    const { user: authUser, loading: authLoading } = useAuth();
    const [profile, setProfile] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const loadProfile = async () => {
            if (authUser) {
                try {
                    const userProfile = await fetchMe();
                    setProfile(userProfile);
                } catch (e) {
                    setError('无法加载个人信息。');
                    console.error(e);
                } finally {
                    setLoading(false);
                }
            } else if (!authLoading) {
                setLoading(false);
                setError('请先登录。');
            }
        };

        loadProfile();
    }, [authUser, authLoading]);

    if (loading || authLoading) {
        return <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}><Spin size="large" /></div>;
    }

    if (error) {
        return <Alert message="错误" description={error} type="error" showIcon />;
    }

    if (!profile) {
        return <Alert message="未找到用户信息" type="warning" showIcon />;
    }

    return (
        <div style={{ maxWidth: 800, margin: 'auto', padding: '24px' }}>
            <Title level={2} style={{ marginBottom: '24px' }}>个人主页</Title>
            <Card>
                <Descriptions bordered column={1}>
                    <Descriptions.Item label="昵称">{profile.display_name}</Descriptions.Item>
                    <Descriptions.Item label="邮箱">{profile.email}</Descriptions.Item>
                    <Descriptions.Item label="角色">{profile.role}</Descriptions.Item>
                    <Descriptions.Item label="注册时间">
                        {profile.created_at ? new Date(profile.created_at).toLocaleString() : 'N/A'}
                    </Descriptions.Item>
                </Descriptions>
            </Card>
        </div>
    );
};

export default MyProfile;