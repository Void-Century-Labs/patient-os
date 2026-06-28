"use client";

import QRCode from "qrcode";
import { useEffect, useState } from "react";

export function HospitalQrCode({ value, size = 200 }: { value: string; size?: number }) {
  const [dataUrl, setDataUrl] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    QRCode.toDataURL(value, { width: size, margin: 1 }).then((url) => {
      if (!cancelled) setDataUrl(url);
    });
    return () => {
      cancelled = true;
    };
  }, [value, size]);

  if (!dataUrl) {
    return (
      <div
        style={{ width: size, height: size }}
        className="flex items-center justify-center rounded-md border text-center text-xs text-muted-foreground"
      >
        Generating QR code…
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-2">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={dataUrl} alt={`QR code linking to ${value}`} width={size} height={size} className="rounded-md border" />
      <a href={dataUrl} download="hospital-qr-code.png" className="text-xs text-muted-foreground underline">
        Download PNG
      </a>
    </div>
  );
}
