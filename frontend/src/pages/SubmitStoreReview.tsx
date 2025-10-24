import { useState } from 'react';
import { Card, Button, message, Typography, Result } from 'antd';
import { useNavigate } from 'react-router-dom';
import StoreSearch from '../components/StoreSearch';
import StoreCreateForm from '../components/StoreCreateForm';
import ReviewForm from '../components/ReviewForm';
import { useAuth } from '../hooks/useAuth';
import type { Store, Review } from '../types';
import { createReviewForStore, updateReview, fetchMyReviews } from '../api/client';

const { Title, Paragraph } = Typography;

type Step = 'search' | 'create_store' | 'review' | 'complete';

const SubmitStoreReview = () => {
  const [step, setStep] = useState<Step>('search');
  const [selectedStore, setSelectedStore] = useState<Store | null>(null);
  const [existingReview, setExistingReview] = useState<Review | null>(null);
  const { user } = useAuth();
  const navigate = useNavigate();

  const handleStoreSelect = async (store: Store) => {
    setSelectedStore(store);
    if (user) {
      try {
        const response = await fetchMyReviews({ query: `store_id:${store.id}`, page_size: 1 });
        setExistingReview(response.data.length > 0 ? response.data[0] : null);
      } catch (error) {
        console.error('获取用户评价失败：', error);
        setExistingReview(null);
      }
    }
    setStep('review');
  };

  const handleStartCreate = () => {
    setStep('create_store');
  };

  const handleReviewSubmit = async (formData: { title: string; content: string; rating: number }) => {
    if (!user || !selectedStore) {
      message.error('发生错误，请刷新重试');
      return;
    }

    try {
      if (existingReview) {
        await updateReview(selectedStore.id, existingReview.id, formData);
        message.success('评价更新成功，等待管理员审核');
      } else {
        await createReviewForStore(selectedStore.id, formData);
        message.success('评价提交成功，等待管理员审核');
      }
      setStep('complete');
    } catch (error: any) {
      message.error(error.response?.data?.detail || '提交失败，请重试');
    }
  };

  const handleCreateSuccess = (store: Store, review: Review) => {
    setSelectedStore(store);
    setExistingReview(review);
    setStep('complete');
  };

  const resetFlow = () => {
    setStep('search');
    setSelectedStore(null);
    setExistingReview(null);
  };

  const renderContent = () => {
    switch (step) {
      case 'search':
        return (
          <>
            <Title level={3}>第一步：选择店铺</Title>
            <Paragraph type="secondary">请先搜索您想要评价的店铺。</Paragraph>
            <StoreSearch onStoreSelect={handleStoreSelect} />
            <div style={{ marginTop: 24, textAlign: 'center' }}>
              <Paragraph>找不到店铺？</Paragraph>
              <Button type="primary" onClick={handleStartCreate}>
                推荐一个新店铺
              </Button>
            </div>
          </>
        );
      case 'create_store':
        return (
          <>
            <Title level={3}>推荐新店铺</Title>
            <Paragraph type="secondary">请填写店铺信息和您的第一条评价。</Paragraph>
            <StoreCreateForm onSuccess={handleCreateSuccess} onCancel={resetFlow} />
          </>
        );
      case 'review':
        if (!selectedStore) return null;
        return (
          <>
            <Title level={3}>{existingReview ? '更新您的评价' : '撰写新评价'}</Title>
            <Card style={{ marginBottom: 24 }}>
              <Title level={5}>{selectedStore.name}</Title>
              <Paragraph>{selectedStore.address}</Paragraph>
            </Card>
            <ReviewForm
              existingReview={existingReview}
              onSubmit={handleReviewSubmit}
              onCancel={resetFlow}
            />
          </>
        );
      case 'complete':
        return (
          <Result
            status="success"
            title="提交成功！"
            subTitle="您的评价已提交，等待管理员审核后即可显示。"
            extra={[
              <Button type="primary" key="continue" onClick={resetFlow}>
                评价其他店铺
              </Button>,
              <Button key="home" onClick={() => navigate('/')}>
                返回首页
              </Button>,
            ]}
          />
        );
    }
  };

  return (
    <Card>
      {renderContent()}
    </Card>
  );
};

export default SubmitStoreReview;