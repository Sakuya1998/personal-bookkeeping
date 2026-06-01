import React from 'react';

type Props = {
  collapsed: boolean;
};

const Brand: React.FC<Props> = ({ collapsed }) => {
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
      <img src="/favicon.svg" width={20} height={20} alt="个人记账" />
      {collapsed ? null : <span style={{ color: '#fff', fontWeight: 600, fontSize: 16 }}>个人记账</span>}
    </div>
  );
};

export default Brand;
