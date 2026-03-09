import message from 'antd/es/message';
import Table from 'antd/es/table';
import type { ReactElement } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { useEffect, useEffectEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { listTemplates } from '../api/client';
import type { ClawTemplate } from '../types';

function TemplateList(): ReactElement {
  const [templates, setTemplates] = useState<ClawTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const { t } = useTranslation();

  const loadTemplates = useEffectEvent(async (): Promise<void> => {
    try {
      setLoading(true);
      const data = await listTemplates();
      setTemplates(data);
    } catch {
      message.error(t('template.loadFailed'));
    } finally {
      setLoading(false);
    }
  });

  useEffect(() => {
    void loadTemplates();
  }, []);

  const columns: ColumnsType<ClawTemplate> = [
    { title: t('common.id'), dataIndex: 'id', key: 'id' },
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('template.description'), dataIndex: 'description', key: 'description' },
    { title: t('template.image'), dataIndex: 'image', key: 'image' },
    { title: t('template.version'), dataIndex: 'version', key: 'version' },
    { title: t('template.port'), dataIndex: 'default_port', key: 'default_port' },
    {
      title: t('common.created'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => new Date(v).toLocaleString(),
    },
  ];

  return (
    <div>
      <h2>{t('template.title')}</h2>
      <Table dataSource={templates} columns={columns} rowKey="id" loading={loading} />
    </div>
  );
}

export default TemplateList;
