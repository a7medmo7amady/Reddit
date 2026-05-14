"use client";

import { useEffect, useRef } from "react";

interface Props {
  src: string;
  className?: string;
  downloadFilename?: string;
}

export default function VideoPlayer({ src, className, downloadFilename }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    if (src.endsWith(".m3u8")) {
      if (video.canPlayType("application/vnd.apple.mpegurl")) {
        // Safari — native HLS
        video.src = src;
      } else {
        // Chrome / Firefox — use hls.js
        import("hls.js").then(({ default: Hls }) => {
          if (!Hls.isSupported()) return;
          const hls = new Hls();
          hls.loadSource(src);
          hls.attachMedia(video);
          return () => hls.destroy();
        });
      }
    } else {
      video.src = src;
    }
  }, [src]);

  return (
    <div>
      <video
        ref={videoRef}
        controls
        playsInline
        className={className}
        style={{ width: "100%", maxHeight: 480, display: "block", background: "#000" }}
      />
      <a
        href={src}
        download={downloadFilename ?? true}
        style={{
          display: "inline-block",
          marginTop: 8,
          padding: "5px 14px",
          background: "#272729",
          color: "#d7dadc",
          border: "1px solid #343536",
          borderRadius: 4,
          fontSize: 12,
          textDecoration: "none",
        }}
      >
        ↓ Download video
      </a>
    </div>
  );
}
