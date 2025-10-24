import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Card, Spin, Typography, List, Rate, Button, message, Empty, Space } from 'antd';
import { getStore, getStoreReviews } from '../api/client';
import type { Store, Review, PaginatedResponse } from '../types';
import ReviewCard from '../components/ReviewCard';

const { Title, Paragraph, Text } = Typography;

const StoreDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [store, setStore] = useState<Store | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [reviewMeta, setReviewMeta] = useState<PaginatedResponse<Review>['pagination'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (!id) {
      setError('店铺 ID 无效');
      setLoading(false);
      return;
    }

    const loadData = async () => {
      setLoading(true);
      setError('');
      try {
        const storeData = await getStore(id);
        setStore(storeData);

        const reviewData = await getStoreReviews(id, { page: 1, page_size: 10 });
        setReviews(reviewData.data);
        setReviewMeta(reviewData.pagination);

      } catch (err) {
        console.error(err);
        setError('加载店铺信息失败，请检查 ID 是否正确或稍后再试。');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [id]);

  const loadMoreReviews = async (page: number, pageSize: number) => {
    if (!id) return;
    try {
        const reviewData = await getStoreReviews(id, { page, page_size: pageSize });
        setReviews(reviewData.data);
        setReviewMeta(reviewData.pagination);
    } catch (err) {
        message.error('加载更多评价失败');
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '50px 0' }}><Spin size="large" /></div>;
  }

  if (error) {
    return <Card><Text type="danger">{error}</Text></Card>;
  }

  if (!store) {
    return <Card><Text>未找到该店铺信息。</Text></Card>;
  }

  return (
    <div className="store-detail-page">
      <Card>
        <Title level={2}>{store.name}</Title>
        <Paragraph>{store.address}</Paragraph>
        <Space align="center" style={{ marginBottom: 16 }}>
            <Rate allowHalf disabled value={store.average_rating} />
            <Text strong>{store.average_rating.toFixed(1)}</Text>
            <Text type="secondary">({store.total_reviews} 条评价)</Text>
        </Space>
        <Paragraph type="secondary">{store.description || '暂无店铺简介'}</Paragraph>
        <Button type="primary">
            <Link to={`/submit-review?storeId=${store.id}`}>为这家店写点评</Link>
        </Button>
      </Card>

      <Title level={3} style={{ marginTop: 32 }}>全部点评</Title>
      
      {reviews.length > 0 ? (
        <List
            dataSource={reviews}
            renderItem={review => (
                <List.Item>
                    <ReviewCard review={review} />
                </List.Item>
            )}
            pagination={reviewMeta ? {
                current: reviewMeta.page,
                pageSize: reviewMeta.page_size,
                total: reviewMeta.total,
                onChange: loadMoreReviews,
            } : false}
        />
      ) : (
        <Empty description="该店铺还没有任何评价，快来抢沙发吧！" />
      )}
    </div>
  );
};

export default StoreDetail;