import type { ReactElement, ReactNode } from 'react';
import { lazy, Suspense } from 'react';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import ConfigProvider from 'antd/es/config-provider';
import Spin from 'antd/es/spin';
import { useTranslation } from 'react-i18next';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';

const AppLayout = lazy(() => import('./components/Layout'));
const InstanceCreate = lazy(() => import('./pages/InstanceCreate'));
const InstanceDetail = lazy(() => import('./pages/InstanceDetail'));
const InstanceList = lazy(() => import('./pages/InstanceList'));
const Login = lazy(() => import('./pages/Login'));
const TemplateList = lazy(() => import('./pages/TemplateList'));
const AdminDashboard = lazy(() => import('./pages/AdminDashboard'));

const antdLocales: Record<string, typeof enUS> = {
  en: enUS,
  zh: zhCN,
};

type RouteMode = 'protected' | 'public' | 'admin';

interface AuthRouteProps {
  children: ReactNode;
  mode: RouteMode;
  redirectTo: string;
}

const appTheme = {
  token: {
    colorPrimary: '#c24b31',
    colorInfo: '#c24b31',
    colorSuccess: '#176246',
    colorWarning: '#b56a14',
    colorError: '#a62e28',
    colorText: '#181411',
    colorTextSecondary: '#4f453d',
    colorBgBase: '#fff8f0',
    colorBgContainer: 'rgba(255, 250, 242, 0.88)',
    colorBorder: 'rgba(94, 67, 47, 0.16)',
    colorSplit: 'rgba(94, 67, 47, 0.1)',
    borderRadius: 18,
    fontFamily: '"IBM Plex Sans", "Noto Sans SC", sans-serif',
  },
  components: {
    Button: {
      controlHeight: 44,
      fontWeight: 600,
      borderRadius: 16,
    },
    Card: {
      borderRadiusLG: 24,
    },
    Input: {
      borderRadius: 16,
      controlHeight: 46,
      activeBorderColor: '#c24b31',
      hoverBorderColor: '#c24b31',
    },
    Select: {
      borderRadius: 16,
      controlHeight: 46,
      activeBorderColor: '#c24b31',
      hoverBorderColor: '#c24b31',
    },
    Table: {
      headerBg: 'transparent',
      rowHoverBg: 'rgba(194, 75, 49, 0.05)',
      borderColor: 'rgba(94, 67, 47, 0.12)',
    },
    Tabs: {
      itemActiveColor: '#c24b31',
      itemColor: '#4f453d',
      itemSelectedColor: '#c24b31',
      inkBarColor: '#c24b31',
    },
  },
};

function FullPageSpinner(): ReactElement {
  return (
    <div className="full-page-spinner">
      <Spin size="large" />
    </div>
  );
}

function RouteContent({ children }: { children: ReactNode }): ReactElement {
  return (
    <Suspense fallback={<FullPageSpinner />}>
      {children}
    </Suspense>
  );
}

function AuthRoute({ children, mode, redirectTo }: AuthRouteProps): ReactElement {
  const { user, isAdmin, loading } = useAuth();

  if (loading) {
    return <FullPageSpinner />;
  }

  if (mode === 'protected' && !user) {
    return <Navigate to={redirectTo} replace />;
  }

  if (mode === 'admin') {
    if (!user) {
      return <Navigate to="/login" replace />;
    }

    if (!isAdmin) {
      return <Navigate to={redirectTo} replace />;
    }
  }

  if (mode === 'public' && user) {
    return <Navigate to={isAdmin ? '/admin' : redirectTo} replace />;
  }

  return <>{children}</>;
}

function App(): ReactElement {
  const { i18n } = useTranslation();
  const localeKey = i18n.language?.startsWith('zh') ? 'zh' : 'en';
  const antdLocale = antdLocales[localeKey];

  return (
    <ConfigProvider locale={antdLocale} theme={appTheme}>
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route
              path="/login"
              element={(
                <AuthRoute mode="public" redirectTo="/instances">
                  <RouteContent>
                    <Login />
                  </RouteContent>
                </AuthRoute>
              )}
            />
            <Route
              element={(
                <AuthRoute mode="protected" redirectTo="/login">
                  <RouteContent>
                    <AppLayout />
                  </RouteContent>
                </AuthRoute>
              )}
            >
              <Route path="/" element={<Navigate to="/instances" replace />} />
              <Route path="/instances" element={<RouteContent><InstanceList /></RouteContent>} />
              <Route path="/instances/create" element={<RouteContent><InstanceCreate /></RouteContent>} />
              <Route path="/instances/:id" element={<RouteContent><InstanceDetail /></RouteContent>} />
              <Route path="/templates" element={<RouteContent><TemplateList /></RouteContent>} />
              <Route
                path="/admin"
                element={(
                  <AuthRoute mode="admin" redirectTo="/instances">
                    <RouteContent>
                      <AdminDashboard />
                    </RouteContent>
                  </AuthRoute>
                )}
              />
            </Route>
          </Routes>
        </AuthProvider>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;
