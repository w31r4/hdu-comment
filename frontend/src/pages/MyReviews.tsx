import { useEffect, useState } from 'react';
import { Card, Table, Tag, Typography, Button } from 'antd';
import { Link } from 'react-router-dom';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import { fetchMyReviews } from '../api/client';
import type { Review } from '../types';

const statusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待审核', color: 'orange' },
  approved: { text: '已通过', color: 'green' },
  rejected: { text: '已驳回', color: 'red' }
};

const MyReviews = () => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<TablePaginationConfig>({ current: 1, pageSize: 10, total: 0 });

  const load = async (page = 1, pageSize = 10) => {
    setLoading(true);
    try {
      const data = await fetchMyReviews({ page, page_size: pageSize, sort: 'created_at', order: 'desc' });
      setReviews(data.data);
      setPagination({ current: data.pagination.page, pageSize: data.pagination.page_size, total: data.pagination.total });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const columns: ColumnsType<Review> = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => <Link to={`/reviews/${record.id}`}>{text}</Link>
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={statusMap[status]?.color || 'default'}>
          {statusMap[status]?.text || status}
        </Tag>
      )
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => rating.toFixed(1)
    },
    {
      title: '驳回原因',
      dataIndex: 'rejection_reason',
      key: 'rejection_reason',
      render: (reason?: string) => reason || '-'
    },
    {
      title: '提交时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) => new Date(value).toLocaleString()
    }
  ];

  return (
    <Card title={<Typography.Title level={4}>我的点评</Typography.Title>}>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={reviews}
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          onChange: (page, pageSize) => load(page, pageSize)
        }}
      />
    </Card>
  );
};

export default MyReviews;
