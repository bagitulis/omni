export function formatAmount(amt: number, cur: string): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: cur || "IDR",
    minimumFractionDigits: 0,
  }).format(amt);
}

export function getCountdown(shipByDate?: number): string {
  if (!shipByDate) return "-";
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return "Overdue";

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  return `${hours}h ${minutes}m`;
}

export function getCountdownColor(shipByDate?: number): string {
  if (!shipByDate) return "#666";
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return "#f5222d";
  if (diff <= 24 * 60 * 60 * 1000) return "#fa8c16";
  return "#666";
}
