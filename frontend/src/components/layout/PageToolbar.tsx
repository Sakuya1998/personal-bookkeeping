import React from 'react';

type Props = {
  left?: React.ReactNode;
  right?: React.ReactNode;
};

const PageToolbar: React.FC<Props> = ({ left, right }) => {
  if (!left && !right) return null;
  return (
    <div className="ui-toolbar">
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>{left}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>{right}</div>
    </div>
  );
};

export default PageToolbar;
