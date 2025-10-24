import { useEffect, useState } from 'react';
import { Card, Table, Tag, Typography, Button } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import { fetchMyReviews } from '../api/client';
import type { Review } from '../types';

const { Title } = Typography;

const statusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待审核', color: 'orange' },
  approved: { text: '已通过', color: 'green' },
  rejected: { text: '已驳回', color: 'red' }
};

const MyReviews = () => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<TablePaginationConfig>({ current: 1, pageSize: 10, total: 0 });
  const navigate = useNavigate();

  const load = async (page = 1, pageSize = 10) => {
    setLoading(true);
    try {
      const paginatedData = await fetchMyReviews({ page, page_size: pageSize, sort: 'created_at', order: 'desc' });
      setReviews(paginatedData.data);
      setPagination({
        current: paginatedData.pagination.page,
        pageSize: paginatedData.pagination.page_size,
        total: paginatedData.pagination.total
      });
    } catch (error) {
      console.error('获取点评失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const columns: ColumnsType<Review> = [
    {
      title: '菜品/店铺',
      dataIndex: ['store', 'name'],
      key: 'store_name',
      render: (text, record) => {
        if (record.store) {
          return <Link to={`/stores/${record.store.id}`}>{text}</Link>;
        }
        return '店铺信息不可用';
      }
    },
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
      render: (status: string) => {
        const statusInfo = statusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={statusInfo.color}>{statusInfo.text}</Tag>;
      }
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => rating.toFixed(1)
    },
    {
      title: '提交时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => new Date(text).toLocaleString()
    }
  ];

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    load(newPagination.current, newPagination.pageSize);
  };

  return (
    <Card>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>我的点评</Title>
        <Button type="primary" onClick={() => navigate('/submit-review')}>
          发表新评论
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={reviews}
        rowKey="id"
        loading={loading}
        pagination={pagination}
        onChange={handleTableChange}
      />
    </Card>
  );
};

export default MyReviews;
