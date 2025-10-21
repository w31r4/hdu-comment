import { useState } from 'react';
import { Form, Input, Button, Select, Rate, message, Typography, Divider } from 'antd';
import type { Store } from '../types';
import { createStoreWithReview } from '../api/store_client';
import { useAuthContext } from '../contexts/AuthContext';

const { Title, Text } = Typography;
const { TextArea } = Input;

interface StoreCreateFormProps {
  onSuccess?: (store: Store, review: any) => void;
  onCancel?: () => void;
}

const StoreCreateForm: React.FC<StoreCreateFormProps> = ({ onSuccess, onCancel }) => {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const { token } = useAuthContext();

  const categories = [
    '中餐', '西餐', '快餐', '日韩料理', '火锅', '烧烤', '饮品', '甜品', '小吃', '其他'
  ];

  const handleSubmit = async (values: any) => {
    if (!token) {
      message.error('请先登录');
      return;
    }

    setSubmitting(true);
    
    try {
      const input = {
        store_name: values.store_name,
        store_address: values.store_address,
        store_phone: values.store_phone || '',
        store_category: values.store_category || '',
        store_description: values.store_description || '',
        review_title: values.review_title,
        review_content: values.review_content,
        rating: values.rating
      };

      const result = await createStoreWithReview(input, token);
      
      message.success('店铺和评价提交成功，等待管理员审核');
      
      if (onSuccess) {
        onSuccess(result.store, result.review);
      }
      
      form.resetFields();
    } catch (error: any) {
      if (error.response?.status === 409) {
        message.error('该店铺已存在，请直接评价');
      } else {
        message.error('提交失败，请重试');
      }
      console.error('创建店铺失败：', error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="store-create-form">
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        className="create-form"
      >
        <Title level={3}>创建新店铺</Title>
        <Text type="secondary">请填写店铺信息，我们将为您创建新店铺</Text>
        
        <Divider>店铺信息</Divider>
        
        <Form.Item
          label="店铺名称"
          name="store_name"
          rules={[{ required: true, message: '请输入店铺名称' }]}
        >
          <Input placeholder="如：学一食堂蛋包饭窗口" />
        </Form.Item>
        
        <Form.Item
          label="店铺地址"
          name="store_address"
          rules={[{ required: true, message: '请输入店铺地址' }]}
        >
          <Input placeholder="如：学一食堂二楼" />
        </Form.Item>
        
        <Form.Item
          label="联系电话"
          name="store_phone"
        >
          <Input placeholder="选填" />
        </Form.Item>
        
        <Form.Item
          label="店铺分类"
          name="store_category"
        >
          <Select placeholder="选择分类" allowClear>
            {categories.map(category => (
              <Select.Option key={category} value={category}>
                {category}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>
        
        <Form.Item
          label="店铺描述"
          name="store_description"
        >
          <TextArea 
            rows={3} 
            placeholder="简要描述店铺特色、环境等（选填）"
            maxLength={500}
            showCount
          />
        </Form.Item>
        
        <Divider>您的评价</Divider>
        
        <Form.Item
          label="评价标题"
          name="review_title"
          rules={[{ required: true, message: '请输入评价标题' }]}
        >
          <Input placeholder="给您的评价起个标题" />
        </Form.Item>
        
        <Form.Item
          label="评价内容"
          name="review_content"
          rules={[{ required: true, message: '请输入评价内容' }]}
        >
          <TextArea 
            rows={4} 
            placeholder="详细描述您的用餐体验、口味、服务等"
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
            <Button type="primary" htmlType="submit" loading={submitting} size="large">
              提交店铺和评价
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
            💡 提示：店铺和评价都需要管理员审核，审核通过后才会显示
          </Text>
        </div>
      </Form>
    </div>
  );
};

export default StoreCreateForm;