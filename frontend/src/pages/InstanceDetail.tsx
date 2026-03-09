import ArrowLeftOutlined from '@ant-design/icons/es/icons/ArrowLeftOutlined';
import DeleteOutlined from '@ant-design/icons/es/icons/DeleteOutlined';
import Button from 'antd/es/button';
import message from 'antd/es/message';
import Popconfirm from 'antd/es/popconfirm';
import Spin from 'antd/es/spin';
import type { ReactElement, ReactNode } from 'react';
import { useEffect, useEffectEvent, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { deleteInstance, getInstance } from '../api/client';
import SecretField from '../components/SecretField';
import StatusBadge from '../components/StatusBadge';
import type { Instance } from '../types';
import { formatDateTime } from '../utils/format';

interface DetailRow {
  label: string;
  value: ReactNode;
}

function InstanceDetail(): ReactElement {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [instance, setInstance] = useState<Instance | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<string>('');
  const intervalRef = useRef<number | null>(null);
  const { t, i18n } = useTranslation();

  const fetchInstance = useEffectEvent(async (): Promise<void> => {
    if (!id) {
      setLoading(false);
      return;
    }

    try {
      const data = await getInstance(id);
      setInstance(data);
      setLastUpdated(new Date().toISOString());
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

  function renderDetailRows(rows: DetailRow[]): ReactElement {
    return (
      <div className="detail-list">
        {rows.map((row) => (
          <div className="detail-row" key={row.label}>
            <span className="detail-label">{row.label}</span>
            <span className="detail-value">{row.value}</span>
          </div>
        ))}
      </div>
    );
  }

  if (loading) {
    return (
      <div className="full-page-spinner">
        <Spin size="large" />
      </div>
    );
  }

  if (!instance) {
    return (
      <div className="empty-state">
        <h3 className="empty-title">{t('common.instanceNotFound')}</h3>
        <p className="empty-copy">{t('instance.detailEmptyCopy')}</p>
        <Button className="ghost-button" onClick={handleBackClick}>
          {t('common.back')}
        </Button>
      </div>
    );
  }

  const runtimeRows: DetailRow[] = [
    { label: t('common.id'), value: instance.id },
    { label: t('instance.template'), value: instance.template_id },
    { label: t('common.created'), value: formatDateTime(instance.created_at, i18n.language) },
    { label: t('instance.lastSync'), value: lastUpdated ? formatDateTime(lastUpdated, i18n.language) : '—' },
  ];

  const infrastructureRows: DetailRow[] = [
    { label: t('instance.namespace'), value: instance.namespace },
    { label: t('instance.deployment'), value: instance.deployment_name },
    { label: t('instance.service'), value: instance.service_name },
    {
      label: t('instance.endpoint'),
      value: instance.endpoint ? (
        <a className="endpoint-link" href={instance.endpoint} target="_blank" rel="noreferrer">
          {instance.endpoint}
        </a>
      ) : (
        t('common.endpointUnavailable')
      ),
    },
  ];

  const secretRows: DetailRow[] = [
    { label: t('instance.apiKey'), value: <SecretField configured /> },
    { label: t('instance.botToken'), value: <SecretField configured /> },
  ];
  const statusHeadlineKey = instance.status === 'running'
    ? 'status.running'
    : instance.status === 'failed'
      ? 'status.failed'
      : instance.status === 'not_found'
        ? 'status.notFound'
        : instance.status === 'starting'
          ? 'status.starting'
          : 'status.pending';

  return (
    <div className="page-stack">
      <section className="page-hero fade-in">
        <div>
          <span className="page-eyebrow">{t('layout.sectionDetail')}</span>
          <h2 className="page-title page-title--compact">{instance.name}</h2>
          <p className="page-subtitle">{t('instance.detailSubtitle')}</p>
        </div>
        <div className="page-actions">
          <Button className="ghost-button" icon={<ArrowLeftOutlined />} onClick={handleBackClick}>
            {t('common.back')}
          </Button>
          <Popconfirm title={t('instance.confirmDeleteDetail')} onConfirm={handleDelete}>
            <Button className="danger-button" danger icon={<DeleteOutlined />}>
              {t('instance.deleteButton')}
            </Button>
          </Popconfirm>
        </div>
      </section>

      <section className="detail-grid fade-in delay-1">
        <article className="status-panel">
          <div className="summary-label">{t('common.status')}</div>
          <div className="summary-value">{t(statusHeadlineKey)}</div>
          <div className="summary-copy"><StatusBadge status={instance.status} /></div>
        </article>
        <article className="status-panel">
          <div className="summary-label">{t('instance.endpoint')}</div>
          <div className="summary-value">{instance.endpoint ? t('instance.endpointReady') : t('instance.endpointPending')}</div>
          <div className="summary-copy">
            {instance.endpoint ? (
              <a className="endpoint-link" href={instance.endpoint} target="_blank" rel="noreferrer">
                {t('common.open')}
              </a>
            ) : (
              t('common.endpointUnavailable')
            )}
          </div>
        </article>
        <article className="status-panel">
          <div className="summary-label">{t('instance.template')}</div>
          <div className="summary-value">{instance.template_id}</div>
          <div className="summary-copy">{t('instance.registryPolling')}</div>
        </article>
      </section>

      <section className="detail-grid fade-in delay-2">
        <article className="content-panel">
          <div className="section-kicker">{t('instance.runtimeEyebrow')}</div>
          <h3 className="section-title">{t('instance.runtimeTitle')}</h3>
          <p className="section-copy">{t('instance.runtimeCopy')}</p>
          {renderDetailRows(runtimeRows)}
        </article>

        <article className="content-panel">
          <div className="section-kicker">{t('instance.infraEyebrow')}</div>
          <h3 className="section-title">{t('instance.infraTitle')}</h3>
          <p className="section-copy">{t('instance.infraCopy')}</p>
          {renderDetailRows(infrastructureRows)}
        </article>

        <article className="content-panel">
          <div className="section-kicker">{t('instance.secretsEyebrow')}</div>
          <h3 className="section-title">{t('instance.secretsTitle')}</h3>
          <p className="section-copy">{t('instance.secretsCopy')}</p>
          {renderDetailRows(secretRows)}
        </article>
      </section>
    </div>
  );
}

export default InstanceDetail;
