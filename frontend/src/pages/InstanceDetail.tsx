import ArrowLeftOutlined from '@ant-design/icons/es/icons/ArrowLeftOutlined';
import DeleteOutlined from '@ant-design/icons/es/icons/DeleteOutlined';
import Button from 'antd/es/button';
import Card from 'antd/es/card';
import Descriptions from 'antd/es/descriptions';
import message from 'antd/es/message';
import Popconfirm from 'antd/es/popconfirm';
import Spin from 'antd/es/spin';
import type { ReactElement } from 'react';
import { useEffect, useEffectEvent, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { getInstance, deleteInstance } from '../api/client';
import SecretField from '../components/SecretField';
import StatusBadge from '../components/StatusBadge';
import type { Instance } from '../types';

function InstanceDetail(): ReactElement {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [instance, setInstance] = useState<Instance | null>(null);
  const [loading, setLoading] = useState(true);
  const intervalRef = useRef<number | null>(null);
  const { t } = useTranslation();

  const fetchInstance = useEffectEvent(async (): Promise<void> => {
    if (!id) return;
    try {
      const data = await getInstance(id);
      setInstance(data);
    } catch {
      message.error(t('instance.loadFailed'));
    } finally {
      setLoading(false);
    }
  });

  function clearPolling(): void {
    if (intervalRef.current === null) {
      return;
    }

    window.clearInterval(intervalRef.current);
    intervalRef.current = null;
  }

  useEffect(() => {
    void fetchInstance();
    intervalRef.current = window.setInterval(() => {
      void fetchInstance();
    }, 5000);

    return clearPolling;
  }, [id]);

  async function handleDelete(): Promise<void> {
    if (!id) return;

    try {
      await deleteInstance(id);
      message.success(t('instance.deleted'));
      navigate('/instances');
    } catch {
      message.error(t('instance.deleteFailed'));
    }
  }

  function handleBackClick(): void {
    navigate('/instances');
  }

  if (loading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  if (!instance) {
    return <div>{t('common.instanceNotFound')}</div>;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={handleBackClick}>{t('common.back')}</Button>
        <Popconfirm title={t('instance.confirmDeleteDetail')} onConfirm={handleDelete}>
          <Button danger icon={<DeleteOutlined />}>{t('instance.deleteButton')}</Button>
        </Popconfirm>
      </div>
      <Card>
        <Descriptions title={instance.name} bordered column={1}>
          <Descriptions.Item label={t('common.id')}>{instance.id}</Descriptions.Item>
          <Descriptions.Item label={t('instance.template')}>{instance.template_id}</Descriptions.Item>
          <Descriptions.Item label={t('common.status')}><StatusBadge status={instance.status} /></Descriptions.Item>
          <Descriptions.Item label={t('instance.namespace')}>{instance.namespace}</Descriptions.Item>
          <Descriptions.Item label={t('instance.deployment')}>{instance.deployment_name}</Descriptions.Item>
          <Descriptions.Item label={t('instance.service')}>{instance.service_name}</Descriptions.Item>
          <Descriptions.Item label={t('instance.endpoint')}>{instance.endpoint || '—'}</Descriptions.Item>
          <Descriptions.Item label={t('instance.apiKey')}><SecretField configured={true} /></Descriptions.Item>
          <Descriptions.Item label={t('instance.botToken')}><SecretField configured={true} /></Descriptions.Item>
          <Descriptions.Item label={t('common.created')}>{new Date(instance.created_at).toLocaleString()}</Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
}

export default InstanceDetail;
