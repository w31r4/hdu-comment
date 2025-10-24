import { useEffect, useState } from 'react';
import { Button, Form, Input, Modal, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import { updateStoreStatus, adminDeleteStore, fetchPendingStores } from '../../api/client';
import type { Store } from '../../types';

const PendingStores = () => {
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<TablePaginationConfig>({ current: 1, pageSize: 10, total: 0 });
  const [query, setQuery] = useState('');
  const [rejectModalOpen, setRejectModalOpen] = useState(false);
  const [selectedStore, setSelectedStore] = useState<Store | null>(null);
  const [form] = Form.useForm();

  const load = async (page = 1, pageSize = 10, keyword = query) => {
    setLoading(true);
    try {
      const data = await fetchPendingStores({
        page,
        page_size: pageSize,
        query: keyword || undefined,
        sort: 'created_at',
        order: 'asc'
      });
      setStores(data.data);
      setPagination({ current: data.pagination.page, pageSize: data.pagination.page_size, total: data.pagination.total });
    } catch (err) {
      console.error(err);
      message.error('加载待审核店铺失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleApprove = async (store: Store) => {
    try {
      await updateStoreStatus(store.id, 'approved');
      message.success('已通过审核');
      load(pagination.current, pagination.pageSize);
    } catch (err) {
      console.error(err);
      message.error('通过失败，请稍后再试');
    }
  };

  const openRejectModal = (store: Store) => {
    setSelectedStore(store);
    form.resetFields();
    setRejectModalOpen(true);
  };

  const handleDelete = (store: Store) => {
    Modal.confirm({
      title: `删除店铺：${store.name}`,
      content: '删除后不可恢复，确认继续吗？',
      okText: '确认删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        try {
          await adminDeleteStore(store.id);
          message.success('已删除该店铺');
          load(pagination.current, pagination.pageSize);
        } catch (err) {
          console.error(err);
          message.error('删除失败，请稍后再试');
        }
      }
    });
  };

  const handleReject = async () => {
    try {
      const values = await form.validateFields();
      if (!selectedStore) return;
      await updateStoreStatus(selectedStore.id, 'rejected', values.reason);
      message.success('已驳回该店铺');
      setRejectModalOpen(false);
      load(pagination.current, pagination.pageSize);
    } catch (err) {
      const validationError = err as { errorFields?: unknown } | undefined;
      if (validationError?.errorFields) return;
      console.error(err);
      message.error('驳回失败，请稍后再试');
    }
  };

  const columns: ColumnsType<Store> = [
    {
      title: '店铺名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
    },
    {
      title: '提交时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) => new Date(value).toLocaleString()
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button type="primary" onClick={() => handleApprove(record)}>
            通过
          </Button>
          <Button danger onClick={() => openRejectModal(record)}>
            驳回
          </Button>
          <Button danger ghost onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      )
    }
  ];

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="搜索"
          allowClear
          onSearch={(value) => {
            setQuery(value);
            load(1, pagination.pageSize ?? 10, value);
          }}
          style={{ width: 240 }}
        />
        <Tag color="orange">共 {pagination.total ?? 0} 条待处理</Tag>
      </Space>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={stores}
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          onChange: (page, pageSize) => load(page, pageSize)
        }}
      />

      <Modal
        open={rejectModalOpen}
        title={`驳回店铺：${selectedStore?.name ?? ''}`}
        onCancel={() => setRejectModalOpen(false)}
        onOk={handleReject}
        okText="确认驳回"
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="驳回原因"
            name="reason"
            rules={[{ required: true, message: '请输入驳回原因' }]}
          >
            <Input.TextArea rows={4} placeholder="请填写驳回原因" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default PendingStores;