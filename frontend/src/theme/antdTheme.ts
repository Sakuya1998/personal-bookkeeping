import type { ThemeConfig } from 'antd';
import { ui } from './tokens';

export const antdTheme: ThemeConfig = {
  token: {
    borderRadius: ui.radius.base,
    borderRadiusLG: ui.radius.modal,
    fontSize: ui.typography.body.fontSize,
    lineHeight: ui.typography.body.lineHeight / ui.typography.body.fontSize,
  },
  components: {
    Button: {
      borderRadius: ui.radius.base,
    },
    Card: {
      borderRadiusLG: ui.radius.base,
    },
    Input: {
      borderRadius: ui.radius.base,
    },
    Modal: {
      borderRadiusLG: ui.radius.modal,
    },
    Table: {
      cellPaddingBlockSM: ui.space.sm,
      cellPaddingInlineSM: ui.space.md,
    },
  },
};
