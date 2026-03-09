import message from 'antd/es/message';
import Spin from 'antd/es/spin';
import type { ReactElement } from 'react';
import { useEffect, useEffectEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { listTemplates } from '../api/client';
import StatusBadge from '../components/StatusBadge';
import type { ClawTemplate } from '../types';
import { formatDateTime } from '../utils/format';

function TemplateList(): ReactElement {
  const [templates, setTemplates] = useState<ClawTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const { t, i18n } = useTranslation();

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

  const imageFamilies = new Set(
    templates
      .map((template) => template.image.split('/')[0])
      .filter((value) => value.length > 0),
  ).size;

  const summaryCards = [
    {
      label: t('template.totalCount'),
      value: templates.length.toString(),
      copy: t('template.totalCountCopy'),
    },
    {
      label: t('template.imageFamilies'),
      value: imageFamilies.toString(),
      copy: t('template.imageFamiliesCopy'),
    },
    {
      label: t('template.deploymentReady'),
      value: templates.filter((template) => template.default_port > 0).length.toString(),
      copy: t('template.deploymentReadyCopy'),
    },
  ];

  return (
    <div className="page-stack">
      <section className="page-hero fade-in">
        <div>
          <span className="page-eyebrow">{t('template.catalogEyebrow')}</span>
          <h2 className="page-title">{t('template.title')}</h2>
          <p className="page-subtitle">{t('template.subtitle')}</p>
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
            <div className="section-kicker">{t('template.catalogEyebrow')}</div>
            <h3 className="section-title">{t('template.catalogTitle')}</h3>
            <p className="section-copy">{t('template.catalogCopy')}</p>
          </div>
        </div>

        {loading ? (
          <div className="full-page-spinner">
            <Spin size="large" />
          </div>
        ) : templates.length === 0 ? (
          <div className="empty-state">
            <h3 className="empty-title">{t('template.emptyTitle')}</h3>
            <p className="empty-copy">{t('template.emptyCopy')}</p>
          </div>
        ) : (
          <div className="template-grid">
            {templates.map((template) => (
              <article key={template.id} className="template-card">
                <div className="template-card-header">
                  <div>
                    <h3 className="template-card-title">{template.name}</h3>
                    <p className="template-card-copy">{template.description}</p>
                  </div>
                  <StatusBadge status="running" />
                </div>

                <div className="meta-pills">
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('template.image')}</span>
                    <span className="meta-pill-value">{template.image}</span>
                  </div>
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('template.version')}</span>
                    <span className="meta-pill-value">{template.version}</span>
                  </div>
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('template.port')}</span>
                    <span className="meta-pill-value">{template.default_port}</span>
                  </div>
                  <div className="meta-pill">
                    <span className="meta-pill-label">{t('template.createdLabel')}</span>
                    <span className="meta-pill-value">{formatDateTime(template.created_at, i18n.language)}</span>
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export default TemplateList;
