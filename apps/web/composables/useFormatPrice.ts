export function useFormatPrice(countryCode = "VE") {
  const localeMap: Record<string, { locale: string; currency: string }> = {
    VE: { locale: "es-VE", currency: "VES" },
    DO: { locale: "es-DO", currency: "DOP" },
    US: { locale: "en-US", currency: "USD" },
  };
  const config = localeMap[countryCode] ?? { locale: "en-US", currency: "USD" };
  const formatter = new Intl.NumberFormat(config.locale, {
    style: "currency",
    currency: config.currency,
    minimumFractionDigits: 2,
  });
  return {
    format: (amount: number) => {
      try {
        return formatter.format(amount);
      } catch {
        return amount.toFixed(2);
      }
    },
  };
}
