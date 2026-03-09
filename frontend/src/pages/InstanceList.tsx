import DeleteOutlined from '@ant-design/icons/es/icons/DeleteOutlined';
import EyeOutlined from '@ant-design/icons/es/icons/EyeOutlined';
import PlusOutlined from '@ant-design/icons/es/icons/PlusOutlined';
import Button from 'antd/es/button';
import message from 'antd/es/message';
import Popconfirm from 'antd/es/popconfirm';
import Space from 'antd/es/space';
import Table from 'antd/es/table';
import type { ReactElement } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { useEffect, useEffectEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { listInstances, deleteInstance } from '../api/client';
import StatusBadge from '../components/StatusBadge';
import type { Instance } from '../types';

function InstanceList(): ReactElement {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const loadInstances = useEffectEvent(async (): Promise<void> => {
    try {
      setLoading(true);
      const data = await listInstances();
      setInstances(data);
    } catch {
      message.error(t('instance.loadFailed'));
    } finally {
      setLoading(false);
    }
  });

  useEffect(() => {
    void loadInstances();
  }, []);

  async function handleDelete(id: string): Promise<void> {
    try {
      await deleteInstance(id);
      message.success(t('instance.deleted'));
      const data = await listInstances();
      setInstances(data);
    } catch {
      message.error(t('instance.deleteFailed'));
    }
  }

  function navigateToCreate(): void {
    navigate('/instances/create');
  }

  function navigateToDetail(id: string): void {
    navigate(`/instances/${id}`);
  }

  const columns: ColumnsType<Instance> = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.id'), dataIndex: 'id', key: 'id' },
    { title: t('instance.template'), dataIndex: 'template_id', key: 'template_id' },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusBadge status={status} />,
    },
    { title: t('instance.endpoint'), dataIndex: 'endpoint', key: 'endpoint', ellipsis: true },
    {
      title: t('common.created'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => new Date(v).toLocaleString(),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_: unknown, record: Instance) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => navigateToDetail(record.id)}
          >
            {t('common.view')}
          </Button>
          <Popconfirm title={t('instance.confirmDelete')} onConfirm={() => handleDelete(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2>{t('instance.title')}</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={navigateToCreate}>
          {t('instance.createButton')}
        </Button>
      </div>
      <Table
        dataSource={instances}
        columns={columns}
        rowKey="id"
        loading={loading}
      />
    </div>
  );
}

export default InstanceList;
