import React from 'react';
import { Card } from 'antd';
import type { CardProps } from 'antd';

type Props = Omit<CardProps, 'size'> & {
  size?: 'default' | 'small';
};

const ContentCard: React.FC<Props> = ({ children, size = 'default', ...rest }) => {
  return (
    <Card size={size} {...rest}>
      {children}
    </Card>
  );
};

export default ContentCard;
