function isHttpLabelUrl(value: string): boolean {
  return value.startsWith("http://") || value.startsWith("https://");
}

function downloadBase64Pdf(fileData: string, orderSn: string): void {
  const byteCharacters = atob(fileData);
  const byteNumbers = new Array(byteCharacters.length);
  for (let i = 0; i < byteCharacters.length; i++) {
    byteNumbers[i] = byteCharacters.charCodeAt(i);
  }

  const blob = new Blob([new Uint8Array(byteNumbers)], {
    type: "application/pdf",
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `label_${orderSn}.pdf`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export function downloadOrderLabel(fileData: string, orderSn: string): void {
  if (!fileData) {
    throw new Error(`Empty label data for order ${orderSn}`);
  }

  if (isHttpLabelUrl(fileData)) {
    window.open(fileData, "_blank", "noopener,noreferrer");
    return;
  }

  downloadBase64Pdf(fileData, orderSn);
}
