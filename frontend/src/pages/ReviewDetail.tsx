import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Alert, Card, Descriptions, Image, Space, Spin, Tag, Typography } from 'antd';
import { fetchReviewDetail } from '../api/client';
import type { Review } from '../types';

const statusMap: Record<Review['status'], { text: string; color: string }> = {
  pending: { text: '待审核', color: 'orange' },
  approved: { text: '已通过', color: 'green' },
  rejected: { text: '已驳回', color: 'red' }
};

const ReviewDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [review, setReview] = useState<Review | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;

    const load = async () => {
      setLoading(true);
      try {
        const data = await fetchReviewDetail(id);
        setReview(data);
        setError('');
      } catch (err) {
        console.error(err);
        setError('获取点评失败或没有权限查看');
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [id]);

  if (loading) {
    return <Spin />;
  }

  if (error || !review) {
    return <Alert type="error" message={error || '点评不存在'} />;
  }

  return (
    <Card>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Typography.Title level={3} style={{ margin: 0 }}>
            {review.store?.name}
          </Typography.Title>
          <Tag color={statusMap[review.status].color}>{statusMap[review.status].text}</Tag>
        </div>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="地址">{review.store?.address}</Descriptions.Item>
          <Descriptions.Item label="评分">{review.rating.toFixed(1)} 分</Descriptions.Item>
          <Descriptions.Item label="点评">
            {review.content || '暂无详细描述'}
          </Descriptions.Item>
          {review.status === 'rejected' && review.rejection_reason && (
            <Descriptions.Item label="驳回原因">
              <Typography.Text type="danger">{review.rejection_reason}</Typography.Text>
            </Descriptions.Item>
          )}
        </Descriptions>
        {review.images && review.images.length > 0 && (
          <Space wrap>
            {review.images.map((image) => (
              <Image
                key={image.id}
                src={image.url}
                alt={review.store?.name}
                width={240}
                height={180}
                style={{ objectFit: 'cover' }}
              />
            ))}
          </Space>
        )}
      </Space>
    </Card>
  );
};

export default ReviewDetail;
