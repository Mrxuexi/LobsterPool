import CheckCircleOutlined from '@ant-design/icons/es/icons/CheckCircleOutlined';
import CloseCircleOutlined from '@ant-design/icons/es/icons/CloseCircleOutlined';
import Tag from 'antd/es/tag';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';

interface SecretFieldProps {
  configured: boolean;
}

function SecretField({ configured }: SecretFieldProps): ReactElement {
  const { t } = useTranslation();

  if (configured) {
    return <Tag className="secret-pill secret-pill--configured" icon={<CheckCircleOutlined />}>{t('secret.configured')}</Tag>;
  }

  return <Tag className="secret-pill secret-pill--missing" icon={<CloseCircleOutlined />}>{t('secret.notConfigured')}</Tag>;
}

export default SecretField;
