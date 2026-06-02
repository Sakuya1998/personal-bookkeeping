import React from 'react';
import { useTranslation } from 'react-i18next';

type Props = {
  collapsed: boolean;
};

const Brand: React.FC<Props> = ({ collapsed }) => {
  const { t } = useTranslation();
  return (
    <div
      style={{
        height: 32,
        margin: 16,
        display: 'flex',
        alignItems: 'center',
        justifyContent: collapsed ? 'center' : 'flex-start',
        gap: 10,
      }}
    >
      <img src="/favicon.svg" width={20} height={20} alt={t('app.name')} />
      {collapsed ? null : <span style={{ color: '#fff', fontWeight: 600, fontSize: 16 }}>{t('app.name')}</span>}
    </div>
  );
};

export default Brand;
