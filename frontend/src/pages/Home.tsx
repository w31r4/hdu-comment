import { useEffect, useState } from 'react';
import { Button, Card, Col, Empty, Input, List, Row, Select, Space, Spin, Typography, Rate } from 'antd';
import { Link } from 'react-router-dom';
import { searchStores } from '../api/client';
import type { PaginatedResponse, Store } from '../types';

const { Title, Paragraph, Text } = Typography;

const Home = () => {
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(8);
  const [query, setQuery] = useState('');
  const [sort, setSort] = useState<string>('created_at');
  const [meta, setMeta] = useState<PaginatedResponse<Store>['pagination'] | null>(null);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await searchStores({
        page,
        page_size: pageSize,
        query: query || undefined,
        sort,
        order: 'desc'
      });
      setStores(data.data);
      setMeta(data.pagination);
    } catch (err) {
      console.error(err);
      setError('加载店铺失败，请稍后再试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, sort, query]);

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Row justify="space-between" align="middle">
        <Col>
          <Title level={3}>发现店铺</Title>
        </Col>
        <Col>
          <Space>
            <Input.Search
              allowClear
              placeholder="搜索店铺名称或地址"
              onSearch={(value) => {
                setPage(1);
                setQuery(value);
              }}
              onChange={(e) => {
                if (!e.target.value) {
                  setQuery('');
                }
              }}
              style={{ width: 260 }}
            />
            <Select
              value={sort}
              onChange={(value) => {
                setSort(value);
                setPage(1);
              }}
              options={[
                { value: 'created_at', label: '最新入驻' },
                { value: 'average_rating', label: '评分最高' },
                { value: 'total_reviews', label: '最热门' }
              ]}
              style={{ width: 120 }}
            />
            <Link to="/submit-review">
              <Button type="primary">推荐新店</Button>
            </Link>
          </Space>
        </Col>
      </Row>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '50px 0' }}><Spin size="large" /></div>
      ) : error ? (
        <Card><Text type="danger">{error}</Text></Card>
      ) : stores.length === 0 ? (
        <Empty description="没有找到符合条件的店铺，快来推荐一家吧！" />
      ) : (
        <List
          grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 4 }}
          dataSource={stores}
          pagination={meta ? {
            current: meta.page,
            pageSize: meta.page_size,
            total: meta.total,
            onChange: (p, size) => {
              setPage(p);
              setPageSize(size);
            },
            showSizeChanger: true
          } : false}
          renderItem={(store) => (
            <List.Item key={store.id}>
              <Link to={`/stores/${store.id}`}>
                <Card
                  title={store.name}
                  hoverable
                  cover={
                    <img
                      alt={store.name}
                      src={`https://via.placeholder.com/400x200.png?text=${encodeURIComponent(store.name)}`}
                      style={{ height: 150, objectFit: 'cover' }}
                    />
                  }
                >
                  <Paragraph ellipsis={{ rows: 2 }}>{store.address}</Paragraph>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Rate allowHalf disabled value={store.average_rating} style={{ fontSize: 16 }} />
                    <Text type="secondary">{store.total_reviews} 条评价</Text>
                  </div>
                </Card>
              </Link>
            </List.Item>
          )}
        />
      )}
    </Space>
  );
};

export default Home;
