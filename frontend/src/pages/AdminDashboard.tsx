import AuditOutlined from '@ant-design/icons/es/icons/AuditOutlined';
import ClusterOutlined from '@ant-design/icons/es/icons/ClusterOutlined';
import CrownOutlined from '@ant-design/icons/es/icons/CrownOutlined';
import DeploymentUnitOutlined from '@ant-design/icons/es/icons/DeploymentUnitOutlined';
import PlusOutlined from '@ant-design/icons/es/icons/PlusOutlined';
import SafetyCertificateOutlined from '@ant-design/icons/es/icons/SafetyCertificateOutlined';
import Button from 'antd/es/button';
import Form from 'antd/es/form';
import Input from 'antd/es/input';
import InputNumber from 'antd/es/input-number';
import message from 'antd/es/message';
import Spin from 'antd/es/spin';
import Table from 'antd/es/table';
import type { ColumnsType } from 'antd/es/table';
import type { ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  createTemplate,
  getAdminOverview,
  listAdminInstances,
  listAdminUsers,
  updateUserMaxInstances,
} from '../api/client';
import StatusBadge from '../components/StatusBadge';
import type {
  AdminInstanceSummary,
  AdminOverview,
  AdminUserSummary,
  CreateTemplateRequest,
} from '../types';
import { formatDateTime } from '../utils/format';

interface CreateTemplateValues {
  id: string;
  name: string;
  description: string;
  image: string;
  version?: string;
  default_port?: number;
}

function buildQuotaDrafts(users: AdminUserSummary[]): Record<string, number> {
  return Object.fromEntries(users.map((user) => [user.id, user.max_instances]));
}

function AdminDashboard(): ReactElement {
  const { t, i18n } = useTranslation();
  const [form] = Form.useForm<CreateTemplateValues>();
  const [overview, setOverview] = useState<AdminOverview | null>(null);
  const [users, setUsers] = useState<AdminUserSummary[]>([]);
  const [instances, setInstances] = useState<AdminInstanceSummary[]>([]);
  const [quotaDrafts, setQuotaDrafts] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [updatingUserID, setUpdatingUserID] = useState<string | null>(null);

  async function reloadAdminData(): Promise<void> {
    try {
      setLoading(true);
      const [overviewData, usersData, instancesData] = await Promise.all([
        getAdminOverview(),
        listAdminUsers(),
        listAdminInstances(),
      ]);
      setOverview(overviewData);
      setUsers(usersData);
      setQuotaDrafts(buildQuotaDrafts(usersData));
      setInstances(instancesData);
    } catch {
      message.error(t('admin.loadFailed'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void (async () => {
      try {
        setLoading(true);
        const [overviewData, usersData, instancesData] = await Promise.all([
          getAdminOverview(),
          listAdminUsers(),
          listAdminInstances(),
        ]);
        setOverview(overviewData);
        setUsers(usersData);
        setQuotaDrafts(buildQuotaDrafts(usersData));
        setInstances(instancesData);
      } catch {
        message.error(t('admin.loadFailed'));
      } finally {
        setLoading(false);
      }
    })();
  }, [t]);

  async function handleCreateTemplate(values: CreateTemplateValues): Promise<void> {
    const payload: CreateTemplateRequest = {
      ...values,
      version: values.version?.trim() || undefined,
      default_port: values.default_port || undefined,
    };

    try {
      setSubmitting(true);
      await createTemplate(payload);
      message.success(t('admin.templateCreated'));
      form.resetFields();
      await reloadAdminData();
    } catch {
      message.error(t('admin.templateCreateFailed'));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleUpdateUserMaxInstances(userID: string): Promise<void> {
    const maxInstances = quotaDrafts[userID];
    if (typeof maxInstances !== 'number' || Number.isNaN(maxInstances) || maxInstances < 0) {
      message.error(t('admin.instanceLimitInvalid'));
      return;
    }

    try {
      setUpdatingUserID(userID);
      await updateUserMaxInstances(userID, { max_instances: maxInstances });
      message.success(t('admin.instanceLimitUpdated'));
      await reloadAdminData();
    } catch {
      message.error(t('admin.instanceLimitUpdateFailed'));
    } finally {
      setUpdatingUserID(null);
    }
  }

  const summaryCards = overview ? [
    {
      label: t('admin.totalUsers'),
      value: overview.total_users.toString(),
      copy: t('admin.totalUsersCopy'),
      icon: <AuditOutlined />,
    },
    {
      label: t('admin.adminUsers'),
      value: overview.admin_users.toString(),
      copy: t('admin.adminUsersCopy'),
      icon: <CrownOutlined />,
    },
    {
      label: t('admin.totalInstances'),
      value: overview.total_instances.toString(),
      copy: t('admin.totalInstancesCopy'),
      icon: <ClusterOutlined />,
    },
    {
      label: t('admin.runningInstances'),
      value: overview.running_instances.toString(),
      copy: t('admin.runningInstancesCopy'),
      icon: <DeploymentUnitOutlined />,
    },
    {
      label: t('admin.totalTemplates'),
      value: overview.total_templates.toString(),
      copy: t('admin.totalTemplatesCopy'),
      icon: <SafetyCertificateOutlined />,
    },
  ] : [];

  const userColumns: ColumnsType<AdminUserSummary> = [
    {
      title: t('common.name'),
      dataIndex: 'username',
      key: 'username',
      render: (username: string, record: AdminUserSummary) => (
        <div className="table-primary">
          <strong>{username}</strong>
          <span>{record.id}</span>
        </div>
      ),
    },
    {
      title: t('admin.role'),
      dataIndex: 'role',
      key: 'role',
      render: (role: AdminUserSummary['role']) => (
        <span className={`role-pill role-pill--${role}`}>
          {role === 'admin' ? t('auth.adminRole') : t('auth.memberRole')}
        </span>
      ),
    },
    {
      title: t('admin.ownedInstances'),
      dataIndex: 'instance_count',
      key: 'instance_count',
    },
    {
      title: t('admin.instanceLimit'),
      dataIndex: 'max_instances',
      key: 'max_instances',
      render: (maxInstances: number, record: AdminUserSummary) => {
        const draftValue = quotaDrafts[record.id] ?? maxInstances;

        return (
          <div className="admin-limit-cell">
            <div className="admin-limit-row">
              <InputNumber
                min={0}
                value={draftValue}
                onChange={(value) => {
                  setQuotaDrafts((current) => ({
                    ...current,
                    [record.id]: typeof value === 'number' ? value : 0,
                  }));
                }}
              />
              <Button
                type="link"
                onClick={() => { void handleUpdateUserMaxInstances(record.id); }}
                loading={updatingUserID === record.id}
              >
                {t('common.save')}
              </Button>
            </div>
            <span className="muted-copy">
              {draftValue === 0
                ? t('admin.instanceLimitUnlimited')
                : t('admin.instanceLimitValue', { count: draftValue })}
            </span>
          </div>
        );
      },
    },
    {
      title: t('common.created'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) => formatDateTime(value, i18n.language),
    },
  ];

  const instanceColumns: ColumnsType<AdminInstanceSummary> = [
    {
      title: t('common.name'),
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: AdminInstanceSummary) => (
        <div className="table-primary">
          <strong>{name}</strong>
          <span>{record.id}</span>
        </div>
      ),
    },
    {
      title: t('admin.owner'),
      dataIndex: 'username',
      key: 'username',
      render: (username: string, record: AdminInstanceSummary) => (
        <div className="table-primary">
          <strong>{username || t('admin.unknownOwner')}</strong>
          <span>{record.user_id}</span>
        </div>
      ),
    },
    {
      title: t('instance.template'),
      dataIndex: 'template_id',
      key: 'template_id',
    },
    {
      title: t('instance.cluster'),
      dataIndex: 'cluster',
      key: 'cluster',
      render: (cluster: string, record: AdminInstanceSummary) => (
        <div className="table-primary">
          <strong>{cluster}</strong>
          <span>{record.namespace}</span>
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
  ];

  return (
    <div className="page-stack">
      <section className="page-hero admin-hero fade-in">
        <div>
          <span className="page-eyebrow">{t('admin.eyebrow')}</span>
          <h2 className="page-title">{t('admin.title')}</h2>
          <p className="page-subtitle">{t('admin.subtitle')}</p>
        </div>
        <div className="admin-hero-aside">
          <div className="admin-hero-card">
            <span>{t('admin.securityLabel')}</span>
            <strong>{t('admin.securityValue')}</strong>
            <p>{t('admin.securityCopy')}</p>
          </div>
        </div>
      </section>

      {loading && !overview ? (
        <div className="full-page-spinner">
          <Spin size="large" />
        </div>
      ) : (
        <>
          <section className="summary-grid admin-summary-grid fade-in delay-1">
            {summaryCards.map((card) => (
              <article key={card.label} className="summary-card admin-summary-card">
                <div className="admin-summary-icon">{card.icon}</div>
                <div className="summary-label">{card.label}</div>
                <div className="summary-value">{card.value}</div>
                <div className="summary-copy">{card.copy}</div>
              </article>
            ))}
          </section>

          <section className="admin-grid fade-in delay-2">
            <div className="content-panel admin-briefing">
              <div className="section-header">
                <div>
                  <div className="section-kicker">{t('admin.briefingEyebrow')}</div>
                  <h3 className="section-title">{t('admin.briefingTitle')}</h3>
                  <p className="section-copy">{t('admin.briefingCopy')}</p>
                </div>
              </div>
              <div className="admin-briefing-list">
                <article className="admin-briefing-item">
                  <span className="admin-briefing-index">01</span>
                  <div>
                    <strong>{t('admin.briefingUsersTitle')}</strong>
                    <p>{t('admin.briefingUsersCopy')}</p>
                  </div>
                </article>
                <article className="admin-briefing-item">
                  <span className="admin-briefing-index">02</span>
                  <div>
                    <strong>{t('admin.briefingInstancesTitle')}</strong>
                    <p>{t('admin.briefingInstancesCopy')}</p>
                  </div>
                </article>
                <article className="admin-briefing-item">
                  <span className="admin-briefing-index">03</span>
                  <div>
                    <strong>{t('admin.briefingTemplatesTitle')}</strong>
                    <p>{t('admin.briefingTemplatesCopy')}</p>
                  </div>
                </article>
              </div>
            </div>

            <div className="content-panel admin-form-panel">
              <div className="section-header">
                <div>
                  <div className="section-kicker">{t('admin.templateFormEyebrow')}</div>
                  <h3 className="section-title">{t('admin.templateFormTitle')}</h3>
                  <p className="section-copy">{t('admin.templateFormCopy')}</p>
                </div>
              </div>

              <Form form={form} layout="vertical" onFinish={handleCreateTemplate}>
                <Form.Item
                  label={t('common.id')}
                  name="id"
                  rules={[{ required: true, message: t('admin.templateIdRequired') }]}
                >
                  <Input data-testid="template-id" placeholder="openclaw-mm-v2" />
                </Form.Item>
                <Form.Item
                  label={t('common.name')}
                  name="name"
                  rules={[{ required: true, message: t('admin.templateNameRequired') }]}
                >
                  <Input data-testid="template-name" placeholder={t('admin.templateNamePlaceholder')} />
                </Form.Item>
                <Form.Item label={t('template.description')} name="description">
                  <Input.TextArea data-testid="template-description" rows={3} placeholder={t('admin.templateDescriptionPlaceholder')} />
                </Form.Item>
                <Form.Item
                  label={t('template.image')}
                  name="image"
                  rules={[{ required: true, message: t('admin.templateImageRequired') }]}
                >
                  <Input data-testid="template-image" placeholder="registry.company.com/openclaw/mm-bot" />
                </Form.Item>
                <div className="admin-form-split">
                  <Form.Item label={t('template.version')} name="version">
                    <Input data-testid="template-version" placeholder="latest" />
                  </Form.Item>
                  <Form.Item label={t('template.port')} name="default_port">
                    <InputNumber data-testid="template-port" min={1} max={65535} style={{ width: '100%' }} placeholder="8080" />
                  </Form.Item>
                </div>
                <Form.Item>
                  <Button
                    data-testid="template-submit"
                    className="accent-button"
                    type="primary"
                    htmlType="submit"
                    icon={<PlusOutlined />}
                    loading={submitting}
                  >
                    {t('admin.createTemplate')}
                  </Button>
                </Form.Item>
              </Form>
            </div>
          </section>

          <section className="content-panel fade-in delay-3">
            <div className="section-header">
              <div>
                <div className="section-kicker">{t('admin.userRegistryEyebrow')}</div>
                <h3 className="section-title">{t('admin.userRegistryTitle')}</h3>
                <p className="section-copy">{t('admin.userRegistryCopy')}</p>
              </div>
            </div>
            <Table
              className="editorial-table"
              dataSource={users}
              columns={userColumns}
              rowKey="id"
              pagination={{ pageSize: 6, hideOnSinglePage: true }}
              scroll={{ x: 900 }}
            />
          </section>

          <section className="content-panel fade-in delay-3">
            <div className="section-header">
              <div>
                <div className="section-kicker">{t('admin.instanceRegistryEyebrow')}</div>
                <h3 className="section-title">{t('admin.instanceRegistryTitle')}</h3>
                <p className="section-copy">{t('admin.instanceRegistryCopy')}</p>
              </div>
            </div>
            <Table
              className="editorial-table"
              dataSource={instances}
              columns={instanceColumns}
              rowKey="id"
              pagination={{ pageSize: 8, hideOnSinglePage: true }}
              scroll={{ x: 980 }}
            />
          </section>
        </>
      )}
    </div>
  );
}

export default AdminDashboard;
