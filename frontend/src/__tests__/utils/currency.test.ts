import { describe, it, expect } from 'vitest';
import { formatCurrency, getCurrencySymbol, getGroupedCurrencies, getAllCurrencyCodes } from '../../utils/currency';

describe('formatCurrency', () => {
  it('formats CNY with 2 decimal places', () => {
    expect(formatCurrency(29.9, 'CNY')).toBe('¥29.90');
  });

  it('formats JPY with 0 decimal places', () => {
    expect(formatCurrency(1000, 'JPY')).toBe('¥1000');
  });

  it('formats KRW with 0 decimal places', () => {
    expect(formatCurrency(50000, 'KRW')).toBe('₩50000');
  });

  it('formats USD with 2 decimal places', () => {
    expect(formatCurrency(0, 'USD')).toBe('$0.00');
  });

  it('rounds EUR to 2 decimal places', () => {
    expect(formatCurrency(99.999, 'EUR')).toBe('€100.00');
  });

  it('handles negative amounts', () => {
    expect(formatCurrency(-50.5, 'USD')).toBe('$-50.50');
  });
});

describe('getCurrencySymbol', () => {
  it('returns ¥ for CNY', () => {
    expect(getCurrencySymbol('CNY')).toBe('¥');
  });

  it('returns $ for USD', () => {
    expect(getCurrencySymbol('USD')).toBe('$');
  });

  it('returns € for EUR', () => {
    expect(getCurrencySymbol('EUR')).toBe('€');
  });

  it('returns £ for GBP', () => {
    expect(getCurrencySymbol('GBP')).toBe('£');
  });

  it('returns the code itself for unknown currencies', () => {
    expect(getCurrencySymbol('XYZ')).toBe('XYZ');
  });

  it('returns the code itself for empty string', () => {
    expect(getCurrencySymbol('')).toBe('');
  });
});

describe('getAllCurrencyCodes', () => {
  it('returns all currency codes', () => {
    const codes = getAllCurrencyCodes();
    expect(codes).toContain('CNY');
    expect(codes).toContain('USD');
    expect(codes).toContain('EUR');
    expect(codes.length).toBeGreaterThan(100);
  });
});

describe('getGroupedCurrencies', () => {
  it('returns all groups with currencies', () => {
    const groups = getGroupedCurrencies();
    const groupLabels = groups.map(g => g.label);
    expect(groupLabels).toContain('major');
    expect(groupLabels).toContain('asia');
    expect(groupLabels).toContain('europe');
    expect(groups.length).toBeGreaterThanOrEqual(7);
  });

  it('each group has at least one option', () => {
    const groups = getGroupedCurrencies();
    for (const g of groups) {
      expect(g.options.length).toBeGreaterThan(0);
    }
  });

  it('major group contains CNY, USD, EUR', () => {
    const groups = getGroupedCurrencies();
    const major = groups.find(g => g.label === 'major');
    expect(major).toBeDefined();
    const codes = major!.options.map(o => o.value);
    expect(codes).toContain('CNY');
    expect(codes).toContain('USD');
    expect(codes).toContain('EUR');
  });
});
