import Badge from 'antd/es/badge';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';

type BadgeStatus = 'success' | 'processing' | 'error' | 'warning' | 'default';

interface StatusBadgeProps {
  status: string;
}

interface StatusConfig {
  status: BadgeStatus;
  key: string;
}

const statusMap: Record<string, StatusConfig> = {
  running: { status: 'success', key: 'status.running' },
  pending: { status: 'processing', key: 'status.pending' },
  starting: { status: 'processing', key: 'status.starting' },
  failed: { status: 'error', key: 'status.failed' },
  not_found: { status: 'warning', key: 'status.notFound' },
};

function StatusBadge({ status }: StatusBadgeProps): ReactElement {
  const { t } = useTranslation();
  const config = statusMap[status];
  const badgeStatus = config ? config.status : 'default';
  const text = config ? t(config.key) : status;

  return <Badge status={badgeStatus} text={text} />;
}

export default StatusBadge;
