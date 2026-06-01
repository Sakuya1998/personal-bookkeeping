export const ui = {
  radius: {
    base: 10,
    modal: 12,
  },
  space: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    xxl: 32,
  },
  typography: {
    pageTitle: { fontSize: 20, lineHeight: 28, fontWeight: 600 },
    sectionTitle: { fontSize: 16, lineHeight: 24, fontWeight: 600 },
    body: { fontSize: 14, lineHeight: 22, fontWeight: 400 },
    help: { fontSize: 12, lineHeight: 20, fontWeight: 400 },
  },
  layout: {
    pageMaxWidth: 1200,
    pagePadding: 24,
    pagePaddingMobile: 16,
  },
} as const;
