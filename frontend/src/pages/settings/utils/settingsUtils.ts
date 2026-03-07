export function validateUrl(url: string): boolean {
  if (!url) {
    return false;
  }

  try {
    const parsed = new URL(url);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch (err) { console.warn("Operation failed:", err);
    return false;
  }
}

export function formatCacheTtl(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "0 sec";
  }

  if (seconds < 60) {
    return `${seconds} sec`;
  }

  if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    return `${minutes} min`;
  }

  const hours = Math.floor(seconds / 3600);
  return `${hours} ${hours === 1 ? "hour" : "hours"}`;
}

export function formatRouteMethod(method: string): string {
  return method.trim().toUpperCase();
}
