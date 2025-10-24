import { useState } from 'react';
import { Form, Input, Button, Rate, message, Typography, Card } from 'antd';
import type { Review } from '../types';

const { Title, Text } = Typography;
const { TextArea } = Input;

interface ReviewFormProps {
  existingReview?: Review | null;
  onSubmit: (data: { title: string; content: string; rating: number }) => void;
  onCancel?: () => void;
}

const ReviewForm: React.FC<ReviewFormProps> = ({ existingReview, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (values: { title: string; content: string; rating: number }) => {
    setSubmitting(true);
    try {
      await onSubmit(values);
    } catch (error) {
      console.error('提交评价失败:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const initialValues = existingReview
    ? {
        title: existingReview.title,
        content: existingReview.content,
        rating: existingReview.rating
      }
    : {
        rating: 3,
        title: '',
        content: ''
      };

  return (
    <Card className="review-form-card">
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={initialValues}
        className="review-form"
      >
        {existingReview && (
          <div className="existing-review-notice">
            <Text type="secondary">
              💡 您已经对该店铺有过评价，这是更新您的评价
            </Text>
          </div>
        )}
        
        
        <Form.Item
          label="标题"
          name="title"
          rules={[{ required: true, message: '请输入一个简洁的标题' }]}
        >
          <Input placeholder="例如：惊艳的麻婆豆腐！" maxLength={50} showCount />
        </Form.Item>
        
        <Form.Item
          label="评价内容"
          name="content"
          rules={[{ required: true, message: '请输入评价内容' }]}
        >
          <TextArea 
            rows={6} 
            placeholder="详细描述您的用餐体验、口味、环境、服务等"
            maxLength={1000}
            showCount
          />
        </Form.Item>
        
        <Form.Item
          label="总体评分"
          name="rating"
          rules={[{ required: true, message: '请选择评分' }]}
        >
          <Rate allowHalf style={{ fontSize: 24 }} />
        </Form.Item>
        
        <Form.Item>
          <div className="form-actions">
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={submitting}
              size="large"
            >
              {existingReview ? '更新评价' : '提交评价'}
            </Button>
            {onCancel && (
              <Button onClick={onCancel} size="large">
                取消
              </Button>
            )}
          </div>
        </Form.Item>
        
        <div className="form-tips">
          <Text type="secondary">
            💡 提示：评价提交后需要管理员审核，审核通过后才会显示
          </Text>
        </div>
      </Form>
    </Card>
  );
};

export default ReviewForm;