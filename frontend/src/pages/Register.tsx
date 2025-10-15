import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Form, Input, Typography } from 'antd';
import { useAuth } from '../hooks/useAuth';

const Register = () => {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (values: { email: string; password: string; displayName: string }) => {
    setError('');
    setLoading(true);
    try {
      await register(values.email, values.password, values.displayName);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError('注册失败，请稍后再试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card style={{ maxWidth: 420, margin: '48px auto' }}>
      <Typography.Title level={3}>注册</Typography.Title>
      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item label="昵称" name="displayName" rules={[{ required: true, message: '请输入昵称' }]}> 
          <Input placeholder="展示名称" />
        </Form.Item>
        <Form.Item label="邮箱" name="email" rules={[{ required: true, message: '请输入邮箱' }]}> 
          <Input type="email" placeholder="name@example.com" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}> 
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
        <Button type="primary" htmlType="submit" block loading={loading}>
          注册
        </Button>
      </Form>
      <Typography.Paragraph style={{ marginTop: 16 }}>
        已有账号？<Link to="/login">直接登录</Link>
      </Typography.Paragraph>
    </Card>
  );
};

export default Register;
