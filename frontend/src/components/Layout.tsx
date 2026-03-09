import AppstoreOutlined from '@ant-design/icons/es/icons/AppstoreOutlined';
import CloudServerOutlined from '@ant-design/icons/es/icons/CloudServerOutlined';
import CrownOutlined from '@ant-design/icons/es/icons/CrownOutlined';
import GlobalOutlined from '@ant-design/icons/es/icons/GlobalOutlined';
import LogoutOutlined from '@ant-design/icons/es/icons/LogoutOutlined';
import PlusOutlined from '@ant-design/icons/es/icons/PlusOutlined';
import UserOutlined from '@ant-design/icons/es/icons/UserOutlined';
import Button from 'antd/es/button';
import Dropdown from 'antd/es/dropdown';
import Form from 'antd/es/form';
import Input from 'antd/es/input';
import Layout from 'antd/es/layout';
import Menu from 'antd/es/menu';
import message from 'antd/es/message';
import Modal from 'antd/es/modal';
import type { MenuProps } from 'antd/es/menu';
import type { ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { formatDisplayDate } from '../utils/format';

const { Header, Content, Footer } = Layout;

function getSelectedMenuKey(pathname: string): string {
  if (pathname.startsWith('/admin')) {
    return '/admin';
  }

  if (pathname.startsWith('/templates')) {
    return '/templates';
  }

  return '/instances';
}

function getSectionContent(pathname: string): { titleKey: string; noteKey: string } {
  if (pathname.startsWith('/admin')) {
    return {
      titleKey: 'layout.sectionAdmin',
      noteKey: 'layout.sectionAdminNote',
    };
  }

  if (pathname.startsWith('/templates')) {
    return {
      titleKey: 'layout.sectionTemplates',
      noteKey: 'layout.sectionTemplatesNote',
    };
  }

  if (pathname.startsWith('/instances/create')) {
    return {
      titleKey: 'layout.sectionCreate',
      noteKey: 'layout.sectionCreateNote',
    };
  }

  if (pathname.startsWith('/instances/')) {
    return {
      titleKey: 'layout.sectionDetail',
      noteKey: 'layout.sectionDetailNote',
    };
  }

  return {
    titleKey: 'layout.sectionInstances',
    noteKey: 'layout.sectionInstancesNote',
  };
}

function AppLayout(): ReactElement {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { user, isAdmin, changePassword, logout } = useAuth();
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const [passwordSubmitting, setPasswordSubmitting] = useState(false);
  const [passwordForm] = Form.useForm<{ newPassword: string; confirmPassword: string }>();
  const selectedKey = getSelectedMenuKey(location.pathname);
  const languageLabel = i18n.language?.startsWith('zh') ? '中文' : 'EN';
  const sectionContent = getSectionContent(location.pathname);
  const todayLabel = formatDisplayDate(new Date(), i18n.language);
  const isAdminRoute = location.pathname.startsWith('/admin');

  useEffect(() => {
    setPasswordModalOpen(Boolean(user?.must_change_password));
  }, [user?.must_change_password]);

  const menuItems: MenuProps['items'] = [
    { key: '/instances', icon: <CloudServerOutlined />, label: t('menu.instances') },
    { key: '/templates', icon: <AppstoreOutlined />, label: t('menu.templates') },
  ];
  if (isAdmin) {
    menuItems.push({ key: '/admin', icon: <CrownOutlined />, label: t('menu.admin') });
  }

  const handleHomeClick = (): void => {
    navigate('/');
  };

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key);
  };

  const handleLanguageClick: MenuProps['onClick'] = ({ key }) => {
    void i18n.changeLanguage(key);
  };

  const handleUserMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key === 'change-password') {
      setPasswordModalOpen(true);
      return;
    }

    if (key !== 'logout') {
      return;
    }

    logout();
    navigate('/login');
  };

  const languageMenu: MenuProps = {
    items: [
      { key: 'en', label: 'English' },
      { key: 'zh', label: '中文' },
    ],
    onClick: handleLanguageClick,
  };

  const userMenu: MenuProps = {
    items: [
      {
        key: 'change-password',
        label: t('auth.changePassword'),
      },
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: t('auth.logout'),
      },
    ],
    onClick: handleUserMenuClick,
  };

  async function handlePasswordFinish(values: { newPassword: string; confirmPassword: string }): Promise<void> {
    if (values.newPassword !== values.confirmPassword) {
      message.error(t('auth.passwordMismatch'));
      return;
    }

    try {
      setPasswordSubmitting(true);
      await changePassword(values.newPassword);
      message.success(t('auth.passwordChanged'));
      passwordForm.resetFields();
      setPasswordModalOpen(false);
    } catch {
      message.error(t('auth.passwordChangeFailed'));
    } finally {
      setPasswordSubmitting(false);
    }
  }

  return (
    <Layout className="app-shell">
      <aside className="app-sidebar fade-in">
        <button type="button" className="brand-button" onClick={handleHomeClick}>
          <span className="brand-mark">LP</span>
          <span className="brand-text">
            <span className="brand-title">{t('common.appName')}</span>
            <span className="brand-copy">{t('layout.launchpadDescription')}</span>
          </span>
        </button>

        <div>
          <div className="app-side-label">{t('layout.navigation')}</div>
          <Menu
            mode="inline"
            selectedKeys={[selectedKey]}
            items={menuItems}
            onClick={handleMenuClick}
            className="app-nav-menu"
          />
        </div>

        <Button className="accent-button" type="primary" icon={<PlusOutlined />} onClick={() => navigate('/instances/create')}>
          {t('instance.createButton')}
        </Button>

        <div className="app-sidebar-card">
          <span className="sidebar-kicker">{isAdminRoute ? t('layout.adminStation') : t('layout.launchpad')}</span>
          <strong className="sidebar-headline">
            {isAdminRoute ? t('layout.sectionAdmin') : t('layout.sectionInstances')}
          </strong>
          <p className="sidebar-note">
            {isAdminRoute ? t('layout.adminStationNote') : t('layout.workspace')}
          </p>
        </div>
      </aside>

      <Layout className="app-main">
        <Header className="app-topbar fade-in delay-1">
          <div className="topbar-copy">
            <span className="topbar-kicker">{t('layout.workspace')}</span>
            <h1 className="topbar-title">{t(sectionContent.titleKey)}</h1>
            <p className="topbar-note">{t(sectionContent.noteKey)}</p>
          </div>
          <div className="topbar-actions">
            <div className="topbar-date">
              <span>{t('layout.today')}</span>
              <strong>{todayLabel}</strong>
            </div>
            <Dropdown menu={languageMenu} trigger={['click']}>
              <Button className="ghost-button" icon={<GlobalOutlined />}>
                {languageLabel}
              </Button>
            </Dropdown>
            {user && (
              <Dropdown menu={userMenu} trigger={['click']}>
                <Button className="ghost-button" icon={<UserOutlined />}>
                  {user.username}{user.role === 'admin' ? ` · ${t('auth.adminRole')}` : ''}
                </Button>
              </Dropdown>
            )}
          </div>
        </Header>
        <Content className="app-content">
          <div className="page-shell fade-in delay-2">
            <Outlet />
          </div>
        </Content>
        <Footer className="app-footer">
          {t('common.footer')}
        </Footer>
      </Layout>
      <Modal
        title={user?.must_change_password ? t('auth.forcePasswordChangeTitle') : t('auth.changePassword')}
        open={passwordModalOpen}
        maskClosable={false}
        closable={!user?.must_change_password}
        keyboard={!user?.must_change_password}
        footer={null}
        onCancel={() => {
          if (user?.must_change_password) {
            return;
          }
          passwordForm.resetFields();
          setPasswordModalOpen(false);
        }}
      >
        <p className="password-modal-copy">
          {user?.must_change_password ? t('auth.forcePasswordChangeCopy') : t('auth.changePasswordCopy')}
        </p>
        <Form form={passwordForm} layout="vertical" onFinish={handlePasswordFinish}>
          <Form.Item
            name="newPassword"
            label={t('auth.newPasswordLabel')}
            rules={[{ required: true, message: t('auth.newPasswordRequired') }]}
          >
            <Input.Password placeholder={t('auth.newPasswordPlaceholder')} />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label={t('auth.confirmPasswordLabel')}
            rules={[{ required: true, message: t('auth.confirmPasswordRequired') }]}
          >
            <Input.Password placeholder={t('auth.confirmPasswordPlaceholder')} />
          </Form.Item>
          <Form.Item>
            <Button className="accent-button" type="primary" htmlType="submit" loading={passwordSubmitting} block>
              {t('auth.changePasswordSubmit')}
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
}

export default AppLayout;
