import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Form, Input, Typography } from 'antd';
import { useAuth } from '../hooks/useAuth';

const Login = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (values: { email: string; password: string }) => {
    setLoading(true);
    setError('');
    try {
      await login(values.email, values.password);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError('登录失败，请检查账号或密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card style={{ maxWidth: 420, margin: '48px auto' }}>
      <Typography.Title level={3}>登录</Typography.Title>
      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item label="邮箱" name="email" rules={[{ required: true, message: '请输入邮箱' }]}> 
          <Input type="email" placeholder="name@example.com" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}> 
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
        <Button type="primary" htmlType="submit" block loading={loading}>
          登录
        </Button>
      </Form>
      <Typography.Paragraph style={{ marginTop: 16 }}>
        还没有账号？<Link to="/register">立即注册</Link>
      </Typography.Paragraph>
    </Card>
  );
};

export default Login;
