export const CURRENCIES = [
  { code: 'CNY', symbol: '¥', name: '人民币' },
  { code: 'USD', symbol: '$', name: '美元' },
  { code: 'EUR', symbol: '€', name: '欧元' },
  { code: 'GBP', symbol: '£', name: '英镑' },
  { code: 'JPY', symbol: '¥', name: '日元' },
  { code: 'HKD', symbol: 'HK$', name: '港币' },
  { code: 'SGD', symbol: 'S$', name: '新加坡元' },
  { code: 'AUD', symbol: 'A$', name: '澳元' },
  { code: 'CAD', symbol: 'C$', name: '加元' },
  { code: 'KRW', symbol: '₩', name: '韩元' },
  { code: 'THB', symbol: '฿', name: '泰铢' },
  { code: 'TWD', symbol: 'NT$', name: '新台币' },
];

export function formatCurrency(amount: number, currency: string): string {
  const c = CURRENCIES.find((c) => c.code === currency);
  const sym = c?.symbol || currency;
  const fixed = ['JPY', 'KRW'].includes(currency) ? 0 : 2;
  return `${sym}${amount.toFixed(fixed)}`;
}

export function getCurrencySymbol(code: string): string {
  return CURRENCIES.find((c) => c.code === code)?.symbol || code;
}
