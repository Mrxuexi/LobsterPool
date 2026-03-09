import Tag from 'antd/es/tag';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';

type StatusTone = 'live' | 'building' | 'critical' | 'warning' | 'muted';

interface StatusBadgeProps {
  status: string;
}

interface StatusConfig {
  tone: StatusTone;
  key: string;
}

const statusMap: Record<string, StatusConfig> = {
  running: { tone: 'live', key: 'status.running' },
  pending: { tone: 'building', key: 'status.pending' },
  starting: { tone: 'building', key: 'status.starting' },
  failed: { tone: 'critical', key: 'status.failed' },
  not_found: { tone: 'warning', key: 'status.notFound' },
};

function StatusBadge({ status }: StatusBadgeProps): ReactElement {
  const { t } = useTranslation();
  const config = statusMap[status];
  const tone = config ? config.tone : 'muted';
  const text = config ? t(config.key) : status;

  return <Tag className={`status-pill status-pill--${tone}`}>{text}</Tag>;
}

export default StatusBadge;
