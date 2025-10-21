import { useState } from 'react';
import { Card, Button, Steps, message, Typography } from 'antd';
import { ShopOutlined, EditOutlined, CheckOutlined } from '@ant-design/icons';
import StoreSearch from '../components/StoreSearch';
import StoreCreateForm from '../components/StoreCreateForm';
import ReviewForm from '../components/ReviewForm';
import { Rate } from 'antd';
import { useAuth } from '../hooks/useAuth';
import type { Store } from '../types';
import { submitStoreReview, updateStoreReview, getMyStoreReview } from '../api/store_client';

const { Title, Text } = Typography;
const { Step } = Steps;

interface ReviewFormData {
  content: string;
  rating: number;
}

const SubmitStoreReview = () => {
  const [currentStep, setCurrentStep] = useState(0);
  const [selectedStore, setSelectedStore] = useState<Store | null>(null);
  const [existingReview, setExistingReview] = useState<any>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const { token } = useAuth();

  const steps = [
    {
      title: '选择店铺',
      icon: <ShopOutlined />,
      content: 'SearchStore'
    },
    {
      title: existingReview ? '更新评价' : '提交评价',
      icon: <EditOutlined />,
      content: 'ReviewForm'
    },
    {
      title: '完成',
      icon: <CheckOutlined />,
      content: 'Complete'
    }
  ];

  const handleStoreSelect = async (store: Store) => {
    setSelectedStore(store);
    
    // 检查用户是否已有该店铺的评价
    if (token) {
      try {
        const review = await getMyStoreReview(store.id, token);
        setExistingReview(review);
      } catch (error) {
        console.error('获取用户评价失败：', error);
      }
    }
    
    setCurrentStep(1);
  };

  const handleCreateNewStore = () => {
    setShowCreateForm(true);
  };

  const handleStoreCreated = (store: Store, review: any) => {
    setSelectedStore(store);
    setExistingReview(review);
    setShowCreateForm(false);
    setCurrentStep(1);
    message.success('店铺和评价创建成功！');
  };

  const handleReviewSubmit = async (formData: ReviewFormData) => {
    if (!token || !selectedStore) {
      message.error('请先登录并选择店铺');
      return;
    }

    try {
      if (existingReview) {
        // 更新现有评价
        const updatedReview = await updateStoreReview(existingReview.id, { content: formData.content, rating: formData.rating }, token);
        setExistingReview(updatedReview);
        message.success('评价更新成功，等待管理员审核');
      } else {
        // 提交新评价
        const newReview = await submitStoreReview({
          store_id: selectedStore.id,
          content: formData.content,
          rating: formData.rating
        }, token);
        setExistingReview(newReview);
        message.success('评价提交成功，等待管理员审核');
      }
      
      setCurrentStep(2);
    } catch (error: any) {
      if (error.response?.status === 409) {
        message.error('您已经对该店铺有过评价，请更新现有评价');
      } else {
        message.error('提交失败，请重试');
      }
      console.error('提交评价失败：', error);
    }
  };

  const renderStepContent = () => {
    if (showCreateForm) {
      return (
        <StoreCreateForm
          onSuccess={handleStoreCreated}
          onCancel={() => setShowCreateForm(false)}
        />
      );
    }

    switch (currentStep) {
      case 0:
        return (
          <div className="step-content">
            <StoreSearch
              onStoreSelect={handleStoreSelect}
            />
            <div className="create-store-option">
              <Text type="secondary">找不到您要的店铺？</Text>
              <Button type="link" onClick={handleCreateNewStore}>
                创建新店铺
              </Button>
            </div>
          </div>
        );
      
      case 1:
        return (
          <div className="step-content">
            {selectedStore && (
              <Card className="selected-store-info">
                <Title level={4}>{selectedStore.name}</Title>
                <Text>{selectedStore.address}</Text>
                {selectedStore.average_rating > 0 && (
                  <div className="store-rating">
                    <Rate disabled value={selectedStore.average_rating} />
                    <Text>{selectedStore.average_rating.toFixed(1)} ({selectedStore.total_reviews}条评价)</Text>
                  </div>
                )}
              </Card>
            )}
            
            <ReviewForm
              existingReview={existingReview}
              onSubmit={handleReviewSubmit}
              onCancel={() => setCurrentStep(0)}
            />
          </div>
        );
      
      case 2:
        return (
          <div className="step-content complete-step">
            <div className="success-icon">
              <CheckOutlined style={{ fontSize: '48px', color: '#52c41a' }} />
            </div>
            <Title level={3}>提交成功！</Title>
            <Text type="secondary">
              {existingReview ? '您的评价已更新' : '您的评价已提交'}，请等待管理员审核。
              审核通过后将在店铺页面显示。
            </Text>
            <div className="complete-actions">
              <Button type="primary" onClick={() => setCurrentStep(0)}>
                继续评价其他店铺
              </Button>
              <Button onClick={() => window.location.href = '/'}>
                返回首页
              </Button>
            </div>
          </div>
        );
      
      default:
        return null;
    }
  };

  return (
    <div className="submit-store-review">
      <Card className="main-card">
        <Title level={2}>
          {existingReview ? '更新店铺评价' : '提交店铺评价'}
        </Title>
        
        <Steps
          current={currentStep}
          items={steps}
          className="progress-steps"
        />
        
        <div className="step-content-wrapper">
          {renderStepContent()}
        </div>
      </Card>
    </div>
  );
};

export default SubmitStoreReview;