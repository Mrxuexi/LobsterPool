import LockOutlined from '@ant-design/icons/es/icons/LockOutlined';
import UserOutlined from '@ant-design/icons/es/icons/UserOutlined';
import Button from 'antd/es/button';
import Form from 'antd/es/form';
import Input from 'antd/es/input';
import message from 'antd/es/message';
import Tabs from 'antd/es/tabs';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

const SERVER_ERROR_MESSAGE_KEYS: Record<string, string> = {
  'username already exists': 'auth.usernameAlreadyExists',
  'invalid username or password': 'auth.invalidUsernameOrPassword',
  'username and password are required': 'auth.usernameAndPasswordRequired',
  'internal error': 'auth.internalError',
  unauthorized: 'auth.unauthorized',
  'user not found': 'auth.userNotFound',
};

type AuthMode = 'login' | 'register';

interface AuthFormValues {
  username: string;
  password: string;
}

interface ApiErrorLike {
  message?: string;
  response?: {
    data?: unknown;
  };
}

function Login(): ReactElement {
  const { t } = useTranslation();
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<AuthMode>('login');
  const featureItems = [
    {
      title: t('auth.operationsCard1Title'),
      copy: t('auth.operationsCard1Copy'),
    },
    {
      title: t('auth.operationsCard2Title'),
      copy: t('auth.operationsCard2Copy'),
    },
    {
      title: t('auth.operationsCard3Title'),
      copy: t('auth.operationsCard3Copy'),
    },
  ];

  function handleTabChange(activeKey: string): void {
    if (activeKey === 'register') {
      setActiveTab('register');
      return;
    }

    setActiveTab('login');
  }

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

    if (!apiError.response) {
      return t('auth.networkError');
    }

    if (typeof apiError.message === 'string' && apiError.message.trim().length > 0) {
      return translateServerError(apiError.message);
    }

    return undefined;
  }

  async function onFinish(values: AuthFormValues): Promise<void> {
    setLoading(true);

    const isLogin = activeTab === 'login';
    const submit = isLogin ? login : register;
    const successMessage = isLogin ? t('auth.loginSuccess') : t('auth.registerSuccess');
    const fallbackErrorMessage = isLogin ? t('auth.loginFailed') : t('auth.registerFailed');

    try {
      const currentUser = await submit(values.username, values.password);
      message.success(successMessage);
      navigate(currentUser.role === 'admin' ? '/admin' : '/instances');
    } catch (err: unknown) {
      const errorMsg = extractErrorMessage(err);
      message.error(errorMsg || fallbackErrorMessage);
    } finally {
      setLoading(false);
    }
  }

  const submitButtonLabel = activeTab === 'login'
    ? t('auth.loginButton')
    : t('auth.registerButton');

  return (
    <div className="auth-stage">
      <section className="auth-editorial fade-in">
        <div>
          <span className="auth-badge">{t('auth.environmentBadge')}</span>
          <h1 className="auth-title">{t('auth.title')}</h1>
          <p className="auth-subtitle">{t('auth.subtitle')}</p>
        </div>

        <div>
          <div className="section-kicker">{t('auth.operationsTitle')}</div>
          <p className="section-copy">{t('auth.operationsCopy')}</p>
          <div className="auth-feature-grid">
            {featureItems.map((item) => (
              <article key={item.title} className="auth-feature">
                <h3>{item.title}</h3>
                <p>{item.copy}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="auth-card fade-in delay-1">
        <div className="auth-card-header">
          <h2 className="auth-card-title">{t('auth.panelTitle')}</h2>
          <p className="auth-card-copy">{t('auth.panelSubtitle')}</p>
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={handleTabChange}
          centered
          items={[
            { key: 'login', label: t('auth.login') },
            { key: 'register', label: t('auth.register') },
          ]}
        />
        <Form onFinish={onFinish} size="large" layout="vertical">
          <Form.Item
            name="username"
            label={t('auth.usernameLabel')}
            rules={[{ required: true, message: t('auth.usernameRequired') }]}
          >
            <Input prefix={<UserOutlined />} placeholder={t('auth.usernamePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="password"
            label={t('auth.passwordLabel')}
            rules={[{ required: true, message: t('auth.passwordRequired') }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.passwordPlaceholder')} />
          </Form.Item>
          <Form.Item>
            <Button className="accent-button" type="primary" htmlType="submit" loading={loading} block>
              {submitButtonLabel}
            </Button>
          </Form.Item>
        </Form>
        <div className="auth-default-admin-note">
          <strong>{t('auth.defaultAdminTitle')}</strong>
          <p>{t('auth.defaultAdminCopy')}</p>
        </div>
        <p className="auth-footnote">{t('auth.trustNote')}</p>
      </section>
    </div>
  );
}

export default Login;
