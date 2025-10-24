import { Card, Tag, Typography, Avatar, Space, Rate, Button } from 'antd';
import { Link } from 'react-router-dom';
import {
    EnvironmentOutlined,
    CalendarOutlined,
    UserOutlined,
    EyeOutlined
} from '@ant-design/icons';
import type { Review } from '../types';
import { useAuth } from '../hooks/useAuth';

const { Title, Paragraph, Text } = Typography;

interface ReviewCardProps {
    review: Review;
    store?: { name: string; address: string };
    status?: 'approved' | 'pending' | 'rejected';
    onDelete?: (review: Review) => void;
    showStatus?: boolean;
}

const ReviewCard = ({ review, store, status, onDelete, showStatus = false }: ReviewCardProps) => {
    const { user } = useAuth();

    const getStatusColor = (s: string) => {
        switch (s) {
            case 'approved': return 'success';
            case 'pending': return 'warning';
            case 'rejected': return 'error';
            default: return 'default';
        }
    };

    const getStatusText = (s: string) => {
        switch (s) {
            case 'approved': return '已发布';
            case 'pending': return '待审核';
            case 'rejected': return '已拒绝';
            default: return s;
        }
    };

    return (
        <Card
            hoverable
            className="review-card"
            cover={
                review.images && review.images.length > 0 && (
                    <div className="review-card-image-container">
                        <img
                            alt={store?.name}
                            src={review.images[0].url}
                            className="review-card-image"
                        />
                        {showStatus && status && (
                            <Tag
                                color={getStatusColor(status)}
                                className="review-status-tag"
                            >
                                {getStatusText(status)}
                            </Tag>
                        )}
                    </div>
                )
            }
            actions={[
                <Link to={`/reviews/${review.id}`} key="view">
                    <Button type="text" icon={<EyeOutlined />}>
                        查看详情
                    </Button>
                </Link>,
                ...(user?.role === 'admin' && onDelete ? [
                    <Button
                        key="delete"
                        type="text"
                        danger
                        onClick={() => onDelete(review)}
                    >
                        删除
                    </Button>
                ] : [])
            ]}
        >
            <div className="review-card-content">
                <Title level={4} className="review-title" ellipsis={{ rows: 1 }}>
                    {review.title}
                </Title>

                <Space className="review-meta" size="small">
                    <Rate
                        disabled
                        defaultValue={review.rating}
                        className="review-rating"
                    />
                    <Text type="secondary" className="rating-text">
                        {review.rating.toFixed(1)}
                    </Text>
                </Space>

                <Paragraph
                    className="review-description"
                    ellipsis={{ rows: 2 }}
                >
                    {review.content || '暂无详细点评'}
                </Paragraph>

                <Space direction="vertical" size="small" className="review-info">
                    {store && (
                        <Space className="info-item">
                            <EnvironmentOutlined className="info-icon" />
                            <Text type="secondary" className="info-text">
                                {store.name} - {store.address}
                            </Text>
                        </Space>
                    )}

                    <Space className="info-item">
                        <UserOutlined className="info-icon" />
                        <Text type="secondary" className="info-text">
                            {review.author?.display_name || '匿名用户'}
                        </Text>
                    </Space>

                    <Space className="info-item">
                        <CalendarOutlined className="info-icon" />
                        <Text type="secondary" className="info-text">
                            {new Date(review.created_at).toLocaleDateString('zh-CN')}
                        </Text>
                    </Space>
                </Space>
            </div>
        </Card>
    );
};

export default ReviewCard;