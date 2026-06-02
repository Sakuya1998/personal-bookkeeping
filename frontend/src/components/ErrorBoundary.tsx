import React from 'react';
import { Button, Card, Typography } from 'antd';
import { withTranslation, WithTranslation } from 'react-i18next';

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps & WithTranslation, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps & WithTranslation) {
    super(props);
    this.state = { error: null, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ error, errorInfo });
    console.error(error, errorInfo);
  }

  handleRefresh = () => {
    window.location.reload();
  };

  render() {
    const { t } = this.props;
    if (this.state.error) {
      return (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '100vh',
            background: '#f0f2f5',
          }}
        >
          <Card
            style={{
              width: 420,
              textAlign: 'center',
              boxShadow: '0 2px 8px rgba(0,0,0,0.09)',
            }}
          >
            <Typography.Title level={3} style={{ marginBottom: 16 }}>
              {t('error.boundaryTitle')}
            </Typography.Title>
            <Typography.Paragraph
              type="danger"
              style={{
                marginBottom: 24,
                wordBreak: 'break-word',
                whiteSpace: 'pre-wrap',
              }}
            >
              {this.state.error.message}
            </Typography.Paragraph>
            <Button type="primary" onClick={this.handleRefresh}>
              {t('error.refresh')}
            </Button>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default withTranslation()(ErrorBoundary);
