export interface CurrencyInfo {
  code: string;
  symbol: string;
  name: string;
  group: string;
}

const groupOrder: string[] = [
  'major', 'asia', 'europe', 'americas', 'middle-east',
  'africa', 'oceania',
];

export const groupLabels: Record<string, string> = {
  major: 'major',
  asia: 'asia',
  europe: 'europe',
  americas: 'americas',
  'middle-east': 'middleEast',
  africa: 'africa',
  oceania: 'oceania',
};

export const CURRENCIES: CurrencyInfo[] = [
  // ── Major ──
  { code: 'CNY', symbol: '¥',   name: 'Chinese Yuan Renminbi',       group: 'major' },
  { code: 'USD', symbol: '$',   name: 'US Dollar',                   group: 'major' },
  { code: 'EUR', symbol: '€',   name: 'Euro',                        group: 'major' },
  { code: 'GBP', symbol: '£',   name: 'British Pound',               group: 'major' },
  { code: 'JPY', symbol: '¥',   name: 'Japanese Yen',                group: 'major' },
  { code: 'CHF', symbol: 'Fr',  name: 'Swiss Franc',                 group: 'major' },
  { code: 'CAD', symbol: 'C$',  name: 'Canadian Dollar',             group: 'major' },
  { code: 'AUD', symbol: 'A$',  name: 'Australian Dollar',           group: 'major' },

  // ── Asia ──
  { code: 'HKD', symbol: 'HK$', name: 'Hong Kong Dollar',            group: 'asia' },
  { code: 'SGD', symbol: 'S$',  name: 'Singapore Dollar',            group: 'asia' },
  { code: 'KRW', symbol: '₩',   name: 'South Korean Won',            group: 'asia' },
  { code: 'TWD', symbol: 'NT$', name: 'New Taiwan Dollar',           group: 'asia' },
  { code: 'INR', symbol: '₹',   name: 'Indian Rupee',                group: 'asia' },
  { code: 'IDR', symbol: 'Rp',  name: 'Indonesian Rupiah',           group: 'asia' },
  { code: 'MYR', symbol: 'RM',  name: 'Malaysian Ringgit',           group: 'asia' },
  { code: 'PHP', symbol: '₱',   name: 'Philippine Peso',             group: 'asia' },
  { code: 'THB', symbol: '฿',   name: 'Thai Baht',                   group: 'asia' },
  { code: 'VND', symbol: '₫',   name: 'Vietnamese Dong',             group: 'asia' },
  { code: 'BDT', symbol: '৳',   name: 'Bangladeshi Taka',            group: 'asia' },
  { code: 'PKR', symbol: '₨',   name: 'Pakistani Rupee',             group: 'asia' },
  { code: 'LKR', symbol: 'Rs',  name: 'Sri Lankan Rupee',            group: 'asia' },
  { code: 'NPR', symbol: 'Rs',  name: 'Nepalese Rupee',              group: 'asia' },
  { code: 'MMK', symbol: 'K',   name: 'Myanmar Kyat',                group: 'asia' },
  { code: 'KHR', symbol: '៛',   name: 'Cambodian Riel',              group: 'asia' },
  { code: 'LAK', symbol: '₭',   name: 'Lao Kip',                     group: 'asia' },
  { code: 'MNT', symbol: '₮',   name: 'Mongolian Tugrik',            group: 'asia' },
  { code: 'MOP', symbol: 'MOP$',name: 'Macanese Pataca',             group: 'asia' },
  { code: 'BND', symbol: 'B$',  name: 'Brunei Dollar',               group: 'asia' },
  { code: 'MVR', symbol: 'Rf',  name: 'Maldivian Rufiyaa',           group: 'asia' },

  // ── Europe ──
  { code: 'SEK', symbol: 'kr',  name: 'Swedish Krona',               group: 'europe' },
  { code: 'NOK', symbol: 'kr',  name: 'Norwegian Krone',             group: 'europe' },
  { code: 'DKK', symbol: 'kr',  name: 'Danish Krone',                group: 'europe' },
  { code: 'ISK', symbol: 'kr',  name: 'Icelandic Krona',             group: 'europe' },
  { code: 'PLN', symbol: 'zł',  name: 'Polish Zloty',                group: 'europe' },
  { code: 'CZK', symbol: 'Kč',  name: 'Czech Koruna',                group: 'europe' },
  { code: 'HUF', symbol: 'Ft',  name: 'Hungarian Forint',            group: 'europe' },
  { code: 'RON', symbol: 'lei', name: 'Romanian Leu',                group: 'europe' },
  { code: 'BGN', symbol: 'лв',  name: 'Bulgarian Lev',               group: 'europe' },
  { code: 'RSD', symbol: 'дин', name: 'Serbian Dinar',               group: 'europe' },
  { code: 'HRK', symbol: 'kn',  name: 'Croatian Kuna',               group: 'europe' },
  { code: 'TRY', symbol: '₺',   name: 'Turkish Lira',                group: 'europe' },
  { code: 'RUB', symbol: '₽',   name: 'Russian Ruble',               group: 'europe' },
  { code: 'UAH', symbol: '₴',   name: 'Ukrainian Hryvnia',           group: 'europe' },
  { code: 'BYN', symbol: 'Br',  name: 'Belarusian Ruble',            group: 'europe' },
  { code: 'MDL', symbol: 'L',   name: 'Moldovan Leu',                group: 'europe' },
  { code: 'GEL', symbol: '₾',   name: 'Georgian Lari',               group: 'europe' },
  { code: 'AMD', symbol: '֏',   name: 'Armenian Dram',               group: 'europe' },
  { code: 'AZN', symbol: '₼',   name: 'Azerbaijani Manat',           group: 'europe' },
  { code: 'ALL', symbol: 'L',   name: 'Albanian Lek',                group: 'europe' },
  { code: 'MKD', symbol: 'ден', name: 'Macedonian Denar',            group: 'europe' },
  { code: 'BAM', symbol: 'KM',  name: 'Bosnia-Herzegovina Mark',     group: 'europe' },

  // ── Americas ──
  { code: 'MXN', symbol: 'Mex$',name: 'Mexican Peso',                group: 'americas' },
  { code: 'BRL', symbol: 'R$',  name: 'Brazilian Real',              group: 'americas' },
  { code: 'ARS', symbol: 'AR$', name: 'Argentine Peso',              group: 'americas' },
  { code: 'CLP', symbol: 'CLP$',name: 'Chilean Peso',                group: 'americas' },
  { code: 'COP', symbol: 'COL$',name: 'Colombian Peso',              group: 'americas' },
  { code: 'PEN', symbol: 'S/',  name: 'Peruvian Sol',                group: 'americas' },
  { code: 'UYU', symbol: '$U',  name: 'Uruguayan Peso',              group: 'americas' },
  { code: 'PYG', symbol: '₲',   name: 'Paraguayan Guarani',          group: 'americas' },
  { code: 'BOB', symbol: 'Bs',  name: 'Bolivian Boliviano',          group: 'americas' },
  { code: 'CRC', symbol: '₡',   name: 'Costa Rican Colon',           group: 'americas' },
  { code: 'DOP', symbol: 'RD$', name: 'Dominican Peso',              group: 'americas' },
  { code: 'GTQ', symbol: 'Q',   name: 'Guatemalan Quetzal',          group: 'americas' },
  { code: 'HNL', symbol: 'L',   name: 'Honduran Lempira',            group: 'americas' },
  { code: 'NIO', symbol: 'C$',  name: 'Nicaraguan Cordoba',          group: 'americas' },
  { code: 'PAB', symbol: 'B/.', name: 'Panamanian Balboa',           group: 'americas' },
  { code: 'TTD', symbol: 'TT$', name: 'Trinidad & Tobago Dollar',    group: 'americas' },
  { code: 'JMD', symbol: 'J$',  name: 'Jamaican Dollar',             group: 'americas' },
  { code: 'BSD', symbol: 'B$',  name: 'Bahamian Dollar',             group: 'americas' },
  { code: 'BBD', symbol: 'Bds$',name: 'Barbadian Dollar',            group: 'americas' },
  { code: 'BMD', symbol: 'BD$', name: 'Bermudian Dollar',            group: 'americas' },
  { code: 'KYD', symbol: 'CI$', name: 'Cayman Islands Dollar',       group: 'americas' },

  // ── Middle East / Central Asia ──
  { code: 'AED', symbol: 'د.إ', name: 'UAE Dirham',                 group: 'middle-east' },
  { code: 'SAR', symbol: '﷼',   name: 'Saudi Riyal',                 group: 'middle-east' },
  { code: 'QAR', symbol: '﷼',   name: 'Qatari Riyal',                group: 'middle-east' },
  { code: 'KWD', symbol: 'د.ك', name: 'Kuwaiti Dinar',              group: 'middle-east' },
  { code: 'BHD', symbol: 'د.ب', name: 'Bahraini Dinar',             group: 'middle-east' },
  { code: 'OMR', symbol: '﷼',   name: 'Omani Rial',                  group: 'middle-east' },
  { code: 'JOD', symbol: 'د.ا', name: 'Jordanian Dinar',            group: 'middle-east' },
  { code: 'LBP', symbol: 'ل.ل', name: 'Lebanese Pound',             group: 'middle-east' },
  { code: 'ILS', symbol: '₪',   name: 'Israeli Shekel',              group: 'middle-east' },
  { code: 'IRR', symbol: '﷼',   name: 'Iranian Rial',                group: 'middle-east' },
  { code: 'IQD', symbol: 'د.ع', name: 'Iraqi Dinar',                group: 'middle-east' },
  { code: 'SYP', symbol: '£S',  name: 'Syrian Pound',                group: 'middle-east' },
  { code: 'YER', symbol: '﷼',   name: 'Yemeni Rial',                 group: 'middle-east' },
  { code: 'AFN', symbol: '؋',   name: 'Afghan Afghani',              group: 'middle-east' },
  { code: 'KZT', symbol: '₸',   name: 'Kazakhstani Tenge',           group: 'middle-east' },
  { code: 'UZS', symbol: 'лв',  name: 'Uzbekistani Som',             group: 'middle-east' },
  { code: 'TMT', symbol: 'm',   name: 'Turkmenistani Manat',         group: 'middle-east' },
  { code: 'KGS', symbol: 'сом', name: 'Kyrgyzstani Som',             group: 'middle-east' },
  { code: 'TJS', symbol: 'SM',  name: 'Tajikistani Somoni',          group: 'middle-east' },

  // ── Africa ──
  { code: 'ZAR', symbol: 'R',   name: 'South African Rand',          group: 'africa' },
  { code: 'NGN', symbol: '₦',   name: 'Nigerian Naira',              group: 'africa' },
  { code: 'EGP', symbol: 'E£',  name: 'Egyptian Pound',              group: 'africa' },
  { code: 'KES', symbol: 'KSh', name: 'Kenyan Shilling',             group: 'africa' },
  { code: 'GHS', symbol: 'GH₵', name: 'Ghanaian Cedi',               group: 'africa' },
  { code: 'TZS', symbol: 'TSh', name: 'Tanzanian Shilling',          group: 'africa' },
  { code: 'UGX', symbol: 'USh', name: 'Ugandan Shilling',            group: 'africa' },
  { code: 'RWF', symbol: 'FRw', name: 'Rwandan Franc',               group: 'africa' },
  { code: 'ETB', symbol: 'Br',  name: 'Ethiopian Birr',              group: 'africa' },
  { code: 'DZD', symbol: 'دج',  name: 'Algerian Dinar',              group: 'africa' },
  { code: 'MAD', symbol: 'د.م.',name: 'Moroccan Dirham',             group: 'africa' },
  { code: 'TND', symbol: 'د.ت', name: 'Tunisian Dinar',              group: 'africa' },
  { code: 'SDG', symbol: 'ج.س', name: 'Sudanese Pound',              group: 'africa' },
  { code: 'ZMW', symbol: 'ZK',  name: 'Zambian Kwacha',              group: 'africa' },
  { code: 'MUR', symbol: 'Rs',  name: 'Mauritian Rupee',             group: 'africa' },
  { code: 'MWK', symbol: 'MK',  name: 'Malawian Kwacha',             group: 'africa' },
  { code: 'AOA', symbol: 'Kz',  name: 'Angolan Kwanza',              group: 'africa' },
  { code: 'MZN', symbol: 'MT',  name: 'Mozambican Metical',          group: 'africa' },
  { code: 'XOF', symbol: 'CFA', name: 'West African CFA Franc',      group: 'africa' },
  { code: 'XAF', symbol: 'FCFA',name: 'Central African CFA Franc',   group: 'africa' },
  { code: 'CDF', symbol: 'FC',  name: 'Congolese Franc',             group: 'africa' },
  { code: 'BWP', symbol: 'P',   name: 'Botswana Pula',               group: 'africa' },

  // ── Oceania ──
  { code: 'NZD', symbol: 'NZ$', name: 'New Zealand Dollar',          group: 'oceania' },
  { code: 'FJD', symbol: 'FJ$', name: 'Fijian Dollar',               group: 'oceania' },
  { code: 'PGK', symbol: 'K',   name: 'Papua New Guinea Kina',       group: 'oceania' },
  { code: 'SBD', symbol: 'SI$', name: 'Solomon Islands Dollar',      group: 'oceania' },
  { code: 'TOP', symbol: 'T$',  name: 'Tongan Paʻanga',              group: 'oceania' },
  { code: 'VUV', symbol: 'VT',  name: 'Vanuatu Vatu',                group: 'oceania' },
  { code: 'WST', symbol: 'WS$', name: 'Samoan Tala',                 group: 'oceania' },
  { code: 'XPF', symbol: 'F',   name: 'CFP Franc',                   group: 'oceania' },
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

/** Group currencies by region, preserving groupOrder */
export function getGroupedCurrencies(): { label: string; title: string; options: { label: string; value: string }[] }[] {
  const map = new Map<string, { label: string; value: string }[]>();
  for (const c of CURRENCIES) {
    if (!map.has(c.group)) map.set(c.group, []);
    map.get(c.group)!.push({ label: `${c.symbol} ${c.code} — ${c.name}`, value: c.code });
  }
  // Return in groupOrder, then alphabetical for any unknown groups
  const groups = new Set(CURRENCIES.map(c => c.group));
  const ordered = [...groupOrder.filter(g => groups.has(g)), ...Array.from(groups).filter(g => !groupOrder.includes(g))];
  return ordered.map(g => ({
    label: g,
    title: groupLabels[g] || g,
    options: map.get(g) || [],
  }));
}

/** Get list of all currency codes */
export function getAllCurrencyCodes(): string[] {
  return CURRENCIES.map(c => c.code);
}
