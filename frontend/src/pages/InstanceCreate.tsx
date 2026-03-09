import Button from 'antd/es/button';
import Form from 'antd/es/form';
import Input from 'antd/es/input';
import message from 'antd/es/message';
import Select from 'antd/es/select';
import type { ReactElement } from 'react';
import { useEffect, useEffectEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { createInstance, listInstances, listTemplates } from '../api/client';
import { useAuth } from '../contexts/AuthContext';
import type { ClawTemplate, CreateInstanceRequest } from '../types';

const SERVER_ERROR_MESSAGE_KEYS: Record<string, string> = {
  'instance limit reached': 'instance.limitReached',
  'template not found': 'instance.templateNotFound',
};

interface ApiErrorLike {
  message?: string;
  response?: {
    data?: unknown;
  };
}

function InstanceCreate(): ReactElement {
  const [templates, setTemplates] = useState<ClawTemplate[]>([]);
  const [instanceCount, setInstanceCount] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { user } = useAuth();
  const selectedTemplateId = Form.useWatch('template_id', form) as string | undefined;
  const selectedTemplate = templates.find((template) => template.id === selectedTemplateId);
  const maxInstances = user?.max_instances ?? 0;
  const limitReached = maxInstances > 0 && instanceCount >= maxInstances;

  const templateOptions = templates.map((tpl) => ({
    value: tpl.id,
    label: `${tpl.name} (${tpl.image}:${tpl.version})`,
  }));

  function translateServerError(msg?: string): string | undefined {
    if (!msg) {
      return undefined;
    }

    const normalized = msg.trim().toLowerCase();
    const key = SERVER_ERROR_MESSAGE_KEYS[normalized];
    return key ? t(key) : msg;
  }

  function extractErrorMessage(err: unknown): string | undefined {
    if (!err || typeof err !== 'object') {
      return undefined;
    }

    const apiError = err as ApiErrorLike;
    const data = apiError.response?.data;
    if (typeof data === 'string' && data.trim().length > 0) {
      return translateServerError(data);
    }

    if (data && typeof data === 'object') {
      const dataObj = data as { error?: unknown; message?: unknown };
      if (typeof dataObj.error === 'string' && dataObj.error.trim().length > 0) {
        return translateServerError(dataObj.error);
      }
      if (typeof dataObj.message === 'string' && dataObj.message.trim().length > 0) {
        return translateServerError(dataObj.message);
      }
    }

    if (typeof apiError.message === 'string' && apiError.message.trim().length > 0) {
      return translateServerError(apiError.message);
    }

    return undefined;
  }

  const loadData = useEffectEvent(async (): Promise<void> => {
    setTemplatesLoading(true);
    const [templateResult, instanceResult] = await Promise.allSettled([
      listTemplates(),
      listInstances(),
    ]);

    if (templateResult.status === 'fulfilled') {
      setTemplates(templateResult.value);
    } else {
      message.error(t('instance.templateLoadFailed'));
    }

    if (instanceResult.status === 'fulfilled') {
      setInstanceCount(instanceResult.value.length);
    } else {
      message.error(t('instance.loadFailed'));
    }

    setTemplatesLoading(false);
  });

  useEffect(() => {
    void loadData();
  }, []);

  async function onFinish(values: CreateInstanceRequest): Promise<void> {
    if (limitReached) {
      message.error(t('instance.limitReached'));
      return;
    }

    try {
      setSubmitting(true);
      const inst = await createInstance(values);
      message.success(t('instance.created'));
      navigate(`/instances/${inst.id}`);
    } catch (err: unknown) {
      message.error(extractErrorMessage(err) || t('instance.createFailed'));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page-stack">
      <section className="page-hero fade-in">
        <div>
          <span className="page-eyebrow">{t('instance.createEyebrow')}</span>
          <h2 className="page-title page-title--compact">{t('instance.createTitle')}</h2>
          <p className="page-subtitle">{t('instance.createSubtitle')}</p>
        </div>
        <div className="page-actions">
          <Button className="ghost-button" onClick={() => navigate('/instances')}>
            {t('common.back')}
          </Button>
        </div>
      </section>

      <div className="create-grid fade-in delay-1">
        <section className="content-panel">
          <div className="section-header">
            <div>
              <div className="section-kicker">{t('instance.checklistTitle')}</div>
              <h3 className="section-title">{t('instance.specTitle')}</h3>
              <p className="section-copy">{t('instance.specCopy')}</p>
            </div>
          </div>

          <div className="checklist">
            <div className="check-item">
              <span className="check-index">01</span>
              <div>
                <strong>{t('instance.checklistName')}</strong>
                <span className="muted-copy">{t('instance.checklistNameCopy')}</span>
              </div>
            </div>
            <div className="check-item">
              <span className="check-index">02</span>
              <div>
                <strong>{t('instance.checklistTemplate')}</strong>
                <span className="muted-copy">{t('instance.checklistTemplateCopy')}</span>
              </div>
            </div>
            <div className="check-item">
              <span className="check-index">03</span>
              <div>
                <strong>{t('instance.checklistSecrets')}</strong>
                <span className="muted-copy">{t('instance.checklistSecretsCopy')}</span>
              </div>
            </div>
          </div>

          <div className="detail-list">
            <div className="meta-pill">
              <span className="meta-pill-label">{t('instance.selectedTemplate')}</span>
              <span className="meta-pill-value">
                {selectedTemplate ? selectedTemplate.name : t('instance.selectedTemplateEmpty')}
              </span>
            </div>
            <div className="meta-pill">
              <span className="meta-pill-label">{t('instance.specRuntime')}</span>
              <span className="meta-pill-value">
                {selectedTemplate ? `${selectedTemplate.image}:${selectedTemplate.version}` : '—'}
              </span>
            </div>
            <div className="meta-pill">
              <span className="meta-pill-label">{t('instance.specPort')}</span>
              <span className="meta-pill-value">
                {selectedTemplate ? selectedTemplate.default_port : '—'}
              </span>
            </div>
            <div className="meta-pill">
              <span className="meta-pill-label">{t('instance.instanceQuota')}</span>
              <span className="meta-pill-value">
                {maxInstances > 0
                  ? t('instance.instanceQuotaValue', { count: maxInstances })
                  : t('instance.instanceQuotaUnlimited')}
              </span>
            </div>
            <div className="meta-pill">
              <span className="meta-pill-label">{t('instance.instanceQuotaUsed')}</span>
              <span className="meta-pill-value">{instanceCount}</span>
            </div>
          </div>

          <div className="security-list">
            <div className="security-item">
              <div>
                <strong>{t('instance.securityTitle')}</strong>
                <span className="muted-copy">{t('instance.securityCopy')}</span>
              </div>
            </div>
          </div>

          {limitReached && (
            <div className="quota-note">
              <strong>{t('instance.limitReachedTitle')}</strong>
              <p>{t('instance.limitReachedCopy', { count: maxInstances })}</p>
            </div>
          )}
        </section>

        <section className="content-panel">
          <div className="section-header">
            <div>
              <div className="section-kicker">{t('instance.createEyebrow')}</div>
              <h3 className="section-title">{t('instance.launchButton')}</h3>
              <p className="section-copy">{t('instance.formCopy')}</p>
            </div>
          </div>

          <Form form={form} layout="vertical" onFinish={onFinish}>
            <Form.Item name="name" label={t('instance.nameLabel')} rules={[{ required: true, message: t('instance.nameRequired') }]}>
              <Input autoComplete="off" placeholder={t('instance.namePlaceholder')} />
            </Form.Item>

            <Form.Item name="template_id" label={t('instance.templateLabel')} rules={[{ required: true, message: t('instance.templateRequired') }]}>
              <Select
                placeholder={t('instance.templatePlaceholder')}
                options={templateOptions}
                loading={templatesLoading}
              />
            </Form.Item>

            <Form.Item name="api_key" label={t('instance.apiKeyLabel')} rules={[{ required: true, message: t('instance.apiKeyRequired') }]}>
              <Input.Password placeholder={t('instance.apiKeyPlaceholder')} />
            </Form.Item>

            <Form.Item name="mm_bot_token" label={t('instance.botTokenLabel')} rules={[{ required: true, message: t('instance.botTokenRequired') }]}>
              <Input.Password placeholder={t('instance.botTokenPlaceholder')} />
            </Form.Item>

            <Form.Item>
              <Button
                className="accent-button"
                type="primary"
                htmlType="submit"
                loading={submitting}
                disabled={limitReached}
                block
              >
                {t('instance.launchButton')}
              </Button>
            </Form.Item>
          </Form>

          <p className="form-note">
            {limitReached ? t('instance.limitReachedFormNote') : t('instance.formNote')}
          </p>
        </section>
      </div>
    </div>
  );
}

export default InstanceCreate;
