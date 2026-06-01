import React from 'react';
import { ConfigProvider } from 'antd';
import { useTranslation } from 'react-i18next';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

const antdLocales: Record<string, typeof zhCN> = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

const dayjsLocales: Record<string, string> = {
  'zh-CN': 'zh-cn',
  'en-US': 'en',
};

const LocaleProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { i18n } = useTranslation();
  const lang = i18n.language in antdLocales ? i18n.language : 'zh-CN';

  React.useEffect(() => {
    dayjs.locale(dayjsLocales[lang] || 'en');
  }, [lang]);

  return (
    <ConfigProvider locale={antdLocales[lang]}>{children}</ConfigProvider>
  );
};

export default LocaleProvider;
