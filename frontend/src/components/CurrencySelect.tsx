import React from 'react';
import { Select, SelectProps } from 'antd';
import { useTranslation } from 'react-i18next';
import { getGroupedCurrencies, groupLabels } from '../utils/currency';

interface CurrencySelectProps extends Omit<SelectProps, 'options'> {
  /** Show only major currencies (default: show all) */
  majorOnly?: boolean;
  /** Show code only in labels (no name) for compact display */
  compact?: boolean;
}

const groupOrder = ['major', 'asia', 'europe', 'americas', 'middle-east', 'africa', 'oceania'];

const CurrencySelect: React.FC<CurrencySelectProps> = ({ majorOnly, compact, ...rest }) => {
  const { t } = useTranslation();

  const groupedOptions = React.useMemo(() => {
    const raw = getGroupedCurrencies();
    let groups = raw;
    if (majorOnly) {
      groups = raw.filter(g => g.label === 'major');
    }
    // Sort by groupOrder
    const groupMap = new Map(groups.map(g => [g.label, g]));
    const keys = groupOrder.filter(k => groupMap.has(k));
    // Append any groups not in order
    for (const k of groupMap.keys()) {
      if (!keys.includes(k)) keys.push(k);
    }

    return keys.map(gk => {
      const g = groupMap.get(gk)!;
      return {
        label: t(`currencyGroup.${groupLabels[gk] || gk}`),
        options: compact
          ? g.options.map(o => ({ ...o, label: o.label.split(' — ')[0] }))
          : g.options,
      };
    });
  }, [majorOnly, compact, t]);

  return (
    <Select
      showSearch
      allowClear
      placeholder={t('common.search')}
      optionFilterProp="label"
      filterOption={(input, option) =>
        (option?.label as string)?.toLowerCase().includes(input.toLowerCase()) ?? false
      }
      {...rest}
      options={groupedOptions}
    />
  );
};

export default CurrencySelect;
