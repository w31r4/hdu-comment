import { Card, Tabs, Typography } from 'antd';
import { ShopOutlined, MessageOutlined } from '@ant-design/icons';
import PendingReviews from '../components/admin/PendingReviews';
import PendingStores from '../components/admin/PendingStores';

const { Title } = Typography;

const AdminPending = () => {
  const items = [
    {
      label: (
        <span>
          <MessageOutlined />
          待审核点评
        </span>
      ),
      key: 'reviews',
      children: <PendingReviews />,
    },
    {
      label: (
        <span>
          <ShopOutlined />
          待审核店铺
        </span>
      ),
      key: 'stores',
      children: <PendingStores />,
    },
  ];

  return (
    <Card>
      <Title level={3}>审核中心</Title>
      <Tabs defaultActiveKey="reviews" items={items} />
    </Card>
  );
};

export default AdminPending;
