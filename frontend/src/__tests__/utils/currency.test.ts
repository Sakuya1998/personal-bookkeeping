import { describe, it, expect } from 'vitest';
import { formatCurrency, getCurrencySymbol } from '../../utils/currency';

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
