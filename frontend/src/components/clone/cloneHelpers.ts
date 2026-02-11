export const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
};

export const currencyFormatter = (
  value: number | string | undefined,
): string => {
  if (value === undefined || value === null) return "";
  return `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ".");
};

export const currencyParser = (value: string | undefined): number => {
  if (!value) return 0;
  return Number(value.replace(/Rp\s?|(\.*)/g, ""));
};
