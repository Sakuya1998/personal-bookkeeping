import React from 'react';

type Props = {
  header?: React.ReactNode;
  toolbar?: React.ReactNode;
  children: React.ReactNode;
};

const PageLayout: React.FC<Props> = ({ header, toolbar, children }) => {
  return (
    <div className="ui-page">
      {header ? <div className="ui-pageHeader">{header}</div> : null}
      {toolbar ? <div style={{ marginBottom: 16 }}>{toolbar}</div> : null}
      {children}
    </div>
  );
};

export default PageLayout;
