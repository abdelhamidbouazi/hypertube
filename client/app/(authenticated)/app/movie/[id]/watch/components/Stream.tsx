"use client";

import { useEffect, useRef, useState } from "react";
import Hls from "hls.js";
import { Button } from "@heroui/button";
import { ListFilterPlus } from "lucide-react";

interface HlsPlayerProps {
  src: string;
  thumbnail?: string;
  token: string;
}

export default function HlsPlayer({ src, token, thumbnail }: HlsPlayerProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const [hls, setHls] = useState<Hls | null>(null);
  const [levels, setLevels] = useState<{ index: number; label: string }[]>([]);
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    let hls: Hls | null = null;

    if (Hls.isSupported()) {
      const hlsInstance = new Hls({
        xhrSetup: (xhr, url) => {
          if (token) {
            xhr.setRequestHeader('Authorization', `Bearer ${token}`);
          }
        },
      });
      hlsInstance.loadSource(src);
      hlsInstance.attachMedia(video);

      hlsInstance.on(Hls.Events.MANIFEST_PARSED, (_, data) => {
        const lvls = data.levels.map((lvl: any, i: number) => ({
          index: i,
          label: `${lvl.height}p (${Math.round(lvl.bitrate / 1000)} kbps)`,
        }));
        setLevels([{ index: -1, label: "Auto" }, ...lvls]);
      });

      setHls(hlsInstance);
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // Safari native HLS - can't easily add auth headers
      // You may need to append token as query parameter for Safari
      const urlWithToken = token ? `${src}?token=${token}` : src;
      video.src = urlWithToken;
    }

    // Cleanup function
    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  }, [src, token]);
  const handleQualityChange = (level: number) => {
    if (!hls) return;
    hls.currentLevel = level;
  };

  return (
    <div className="max-w-7xl mx-auto text-center">
      {levels.length > 0 && (
        <div className="flex gap-2 justify-between">
          {levels.map((lvl) => (
            <Button
              variant="flat"
              color="secondary"
              key={lvl.index}
              value={lvl.index.toString()}
              onPress={() => {
                handleQualityChange(lvl.index);
              }}
            >
              {`${lvl.label.split(" ")[0]}`}
            </Button>
          ))}
        </div>
      )}
      <video
        ref={videoRef}
        controls
        autoPlay
        style={{ width: "100%", borderRadius: 10 }}
      >
        {levels.length > 0 && (
          <select
            onChange={(e) => handleQualityChange(Number(e.target.value))}
            style={{ marginTop: 10 }}
          >
            {levels.map((lvl) => (
              <option key={lvl.index} value={lvl.index}>
                {lvl.label}
              </option>
            ))}
          </select>
        )}
        <track
          src="/subs/en.vtt"
          kind="subtitles"
          srcLang="en"
          label="English"
          default
        />
      </video>
    </div>
  );
}
