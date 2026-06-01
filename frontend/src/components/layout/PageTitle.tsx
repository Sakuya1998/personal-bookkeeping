import React from 'react';

type Props = {
  title: string;
  description?: string;
};

const PageTitle: React.FC<Props> = ({ title, description }) => {
  return (
    <div>
      <h1 className="ui-pageTitle">{title}</h1>
      {description ? <div className="ui-pageDesc">{description}</div> : null}
    </div>
  );
};

export default PageTitle;
