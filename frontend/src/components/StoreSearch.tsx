import { useState, useEffect } from 'react';
import { Input, List, Card, Button, Spin, Empty, Rate, Tag } from 'antd';
import { SearchOutlined, ShopOutlined, PhoneOutlined, EnvironmentOutlined } from '@ant-design/icons';
import type { Store } from '../types';
import { searchStores } from '../api/client';

const { Search } = Input;

interface StoreSearchProps {
  onStoreSelect?: (store: Store) => void;
  selectedStoreId?: string;
}

const StoreSearch: React.FC<StoreSearchProps> = ({ onStoreSelect, selectedStoreId }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = async (query: string) => {
    if (!query.trim()) {
      setStores([]);
      setHasSearched(false);
      return;
    }

    setLoading(true);
    setHasSearched(true);
    
    try {
      const result = await searchStores({ query, page: 1, page_size: 10 });
      setStores(result.data);
    } catch (error) {
      console.error('搜索店铺失败：', error);
      setStores([]);
    } finally {
      setLoading(false);
    }
  };

  const handleStoreSelect = (store: Store) => {
    if (onStoreSelect) {
      onStoreSelect(store);
    }
  };

  const renderStoreCard = (store: Store) => (
    <Card
      key={store.id}
      hoverable
      className={`store-card ${selectedStoreId === store.id ? 'selected' : ''}`}
      onClick={() => handleStoreSelect(store)}
      actions={[
        <Button type="primary" size="small">
          选择此店铺
        </Button>
      ]}
    >
      <div className="store-info">
        <div className="store-header">
          <h3 className="store-name">{store.name}</h3>
          <div className="store-rating">
            <Rate disabled value={store.average_rating} />
            <span className="rating-text">{store.average_rating.toFixed(1)}</span>
            <span className="review-count">({store.total_reviews}条评价)</span>
          </div>
        </div>
        
        <div className="store-details">
          <div className="detail-item">
            <EnvironmentOutlined />
            <span>{store.address}</span>
          </div>
          
          {store.phone && (
            <div className="detail-item">
              <PhoneOutlined />
              <span>{store.phone}</span>
            </div>
          )}
          
          {store.category && (
            <div className="detail-item">
              <Tag color="blue">{store.category}</Tag>
            </div>
          )}
        </div>
        
        {store.description && (
          <div className="store-description">
            {store.description}
          </div>
        )}
      </div>
    </Card>
  );

  return (
    <div className="store-search">
      <div className="search-header">
        <h2>
          <ShopOutlined />
          选择店铺
        </h2>
        <p className="search-subtitle">搜索您要评价的店铺，或创建新店铺</p>
      </div>
      
      <Search
        placeholder="输入店铺名称或地址进行搜索"
        allowClear
        enterButton={<SearchOutlined />}
        size="large"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        onSearch={handleSearch}
        loading={loading}
        className="search-input"
      />
      
      <div className="search-results">
        {loading && (
          <div className="loading-container">
            <Spin size="large" />
          </div>
        )}
        
        {!loading && hasSearched && stores.length === 0 && (
          <Empty
            description="未找到相关店铺"
            className="empty-result"
          >
            <Button type="primary" onClick={() => onStoreSelect?.({} as Store)}>
              创建新店铺
            </Button>
          </Empty>
        )}
        
        {!loading && stores.length > 0 && (
          <List
            grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3 }}
            dataSource={stores}
            renderItem={renderStoreCard}
            className="store-list"
          />
        )}
      </div>
    </div>
  );
};

export default StoreSearch;