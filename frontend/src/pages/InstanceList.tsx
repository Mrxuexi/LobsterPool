import DeleteOutlined from '@ant-design/icons/es/icons/DeleteOutlined';
import EyeOutlined from '@ant-design/icons/es/icons/EyeOutlined';
import PlusOutlined from '@ant-design/icons/es/icons/PlusOutlined';
import Button from 'antd/es/button';
import message from 'antd/es/message';
import Popconfirm from 'antd/es/popconfirm';
import Spin from 'antd/es/spin';
import Space from 'antd/es/space';
import Table from 'antd/es/table';
import type { ReactElement } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { useEffect, useEffectEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { deleteInstance, listInstances } from '../api/client';
import StatusBadge from '../components/StatusBadge';
import type { Instance } from '../types';
import { formatDateTime } from '../utils/format';

function useCompactLayout(): boolean {
  const [isCompact, setIsCompact] = useState(() => (
    typeof window !== 'undefined' ? window.innerWidth < 768 : false
  ));

  useEffect(() => {
    const mediaQuery = window.matchMedia('(max-width: 767px)');
    const handleChange = (event: MediaQueryListEvent): void => {
      setIsCompact(event.matches);
    };

    mediaQuery.addEventListener('change', handleChange);

    return (): void => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  }, []);

  return isCompact;
}

function InstanceList(): ReactElement {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const useCardList = useCompactLayout();

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

  const runningCount = instances.filter((instance) => instance.status === 'running').length;
  const endpointCount = instances.filter((instance) => instance.endpoint).length;
  const templateCoverage = new Set(instances.map((instance) => instance.template_id)).size;

  const summaryCards = [
    {
      label: t('instance.totalCount'),
      value: instances.length.toString(),
      copy: t('instance.totalCountCopy'),
    },
    {
      label: t('instance.activeCount'),
      value: runningCount.toString(),
      copy: t('instance.activeCountCopy'),
    },
    {
      label: t('instance.publishedCount'),
      value: endpointCount.toString(),
      copy: t('instance.publishedCountCopy'),
    },
    {
      label: t('instance.templateCoverage'),
      value: templateCoverage.toString(),
      copy: t('instance.templateCoverageCopy'),
    },
  ];

  const columns: ColumnsType<Instance> = [
    {
      title: t('common.name'),
      dataIndex: 'name',
      key: 'name',
      render: (_: string, record: Instance) => (
        <div className="table-primary">
          <strong>{record.name}</strong>
          <span>{record.id}</span>
        </div>
      ),
    },
    {
      title: t('instance.template'),
      dataIndex: 'template_id',
      key: 'template_id',
      render: (templateId: string) => (
        <div className="table-primary">
          <strong>{templateId}</strong>
          <span>{t('instance.templateLabel')}</span>
        </div>
      ),
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusBadge status={status} />,
    },
    {
      title: t('instance.endpoint'),
      dataIndex: 'endpoint',
      key: 'endpoint',
      render: (endpoint: string) => endpoint ? (
        <a className="endpoint-link" href={endpoint} target="_blank" rel="noreferrer">
          {endpoint}
        </a>
      ) : (
        <span className="muted-copy">{t('common.endpointUnavailable')}</span>
      ),
    },
    {
      title: t('common.created'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) => formatDateTime(value, i18n.language),
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
    <div className="page-stack">
      <section className="page-hero fade-in">
        <div>
          <span className="page-eyebrow">{t('instance.registryEyebrow')}</span>
          <h2 className="page-title">{t('instance.title')}</h2>
          <p className="page-subtitle">{t('instance.subtitle')}</p>
        </div>
        <div className="page-actions">
          <Button className="accent-button" type="primary" icon={<PlusOutlined />} onClick={navigateToCreate}>
            {t('instance.createButton')}
          </Button>
        </div>
      </section>

      <section className="summary-grid fade-in delay-1">
        {summaryCards.map((card) => (
          <article key={card.label} className="summary-card">
            <div className="summary-label">{card.label}</div>
            <div className="summary-value">{card.value}</div>
            <div className="summary-copy">{card.copy}</div>
          </article>
        ))}
      </section>

      <section className="content-panel fade-in delay-2">
        <div className="section-header">
          <div>
            <div className="section-kicker">{t('instance.registryEyebrow')}</div>
            <h3 className="section-title">{t('instance.registryTitle')}</h3>
            <p className="section-copy">{t('instance.registryHint')}</p>
          </div>
          <div className="section-note">{t('instance.registryPolling')}</div>
        </div>

        {instances.length === 0 && !loading ? (
          <div className="empty-state">
            <h3 className="empty-title">{t('instance.emptyTitle')}</h3>
            <p className="empty-copy">{t('instance.emptyCopy')}</p>
            <Button className="accent-button" type="primary" icon={<PlusOutlined />} onClick={navigateToCreate}>
              {t('instance.createButton')}
            </Button>
          </div>
        ) : loading ? (
          <div className="full-page-spinner">
            <Spin size="large" />
          </div>
        ) : useCardList ? (
          <div className="mobile-instance-list">
            {instances.map((instance) => (
              <article key={instance.id} className="mobile-instance-card">
                <div className="mobile-instance-card-header">
                  <div>
                    <h4 className="mobile-instance-card-title">{instance.name}</h4>
                    <div className="mobile-instance-card-id">{instance.id}</div>
                  </div>
                  <StatusBadge status={instance.status} />
                </div>

                <div className="detail-list">
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('instance.template')}</span>
                    <span className="meta-pill-value">{instance.template_id}</span>
                  </div>
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('instance.endpoint')}</span>
                    <span className="meta-pill-value">
                      {instance.endpoint ? (
                        <a className="endpoint-link" href={instance.endpoint} target="_blank" rel="noreferrer">
                          {t('common.open')}
                        </a>
                      ) : (
                        t('common.endpointUnavailable')
                      )}
                    </span>
                  </div>
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('common.created')}</span>
                    <span className="meta-pill-value">{formatDateTime(instance.created_at, i18n.language)}</span>
                  </div>
                </div>

                <div className="mobile-instance-card-actions">
                  <Button type="default" icon={<EyeOutlined />} onClick={() => navigateToDetail(instance.id)}>
                    {t('common.view')}
                  </Button>
                  <Popconfirm title={t('instance.confirmDelete')} onConfirm={() => handleDelete(instance.id)}>
                    <Button danger icon={<DeleteOutlined />}>
                      {t('common.delete')}
                    </Button>
                  </Popconfirm>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <Table
            className="editorial-table"
            dataSource={instances}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 6, hideOnSinglePage: true }}
          />
        )}
      </section>
    </div>
  );
}

export default InstanceList;
