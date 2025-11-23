"use client";

import { useEffect, useRef, useState } from "react";
import Hls from "hls.js";
import { Button } from "@heroui/button";
import { Progress } from "@heroui/progress";
import api from "@/lib/api";

interface HlsPlayerProps {
  src: string;
  thumbnail?: string;
  token: string;
  movieTitle: string;
  movieId: string;
}

interface DownloadProgress {
  status: string;
  progress: number;
  stream_ready: boolean;
  quality: string;
}

export default function HlsPlayer({ src, token, thumbnail: _thumbnail, movieTitle, movieId }: HlsPlayerProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [levels, setLevels] = useState<{ index: number; label: string }[]>([]);
  const [currentLevel, setCurrentLevel] = useState<number>(-1);
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    if (Hls.isSupported()) {
      const hlsInstance = new Hls({
        xhrSetup: (xhr, _url) => {
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
        setCurrentLevel(-1);
      });

      hlsInstance.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
        // hls reports the level index (auto resolves to an actual level >=0)
        // We keep -1 to represent "Auto" when ABR is enabled
        const abrEnabled = hlsInstance.autoLevelEnabled;
        setCurrentLevel(abrEnabled ? -1 : data.level);
      });

      hlsRef.current = hlsInstance;
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // Safari native HLS - can't easily add auth headers
      // You may need to append token as query parameter for Safari
      const urlWithToken = token ? `${src}?token=${token}` : src;
      video.src = urlWithToken;
    }

    // Cleanup function
    return () => {
      // Destroy the local instance created in this effect
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, token]);

  // Poll download progress
  useEffect(() => {
    if (!movieId) return;

    let shouldContinuePolling = true;

    const checkDownloadProgress = async () => {
      if (!shouldContinuePolling) return;

      // Try common qualities
      const qualities = ["1080p", "720p", "480p"];
      
      for (const quality of qualities) {
        try {
          const response = await api.get<DownloadProgress>("/torrents/progress", {
            params: { movie_id: movieId, quality },
          });
          
          if (response.data) {
            setDownloadProgress(response.data);
            // If download is completed or streaming, stop polling
            if (response.data.status === "completed" || response.data.stream_ready) {
              shouldContinuePolling = false;
              return;
            }
            break; // Found active download, stop trying other qualities
          }
        } catch (error: any) {
          // If 404, try next quality
          if (error.response?.status === 404) {
            continue;
          }
          // Other errors, stop trying
          break;
        }
      }
    };

    // Check immediately
    checkDownloadProgress();

    // Poll every 2 seconds while download is in progress
    const interval = setInterval(() => {
      if (shouldContinuePolling) {
        checkDownloadProgress();
      } else {
        clearInterval(interval);
      }
    }, 2000);

    return () => {
      shouldContinuePolling = false;
      clearInterval(interval);
    };
  }, [movieId]);

  const handleQualityChange = (level: number) => {
    if (!hlsRef.current) return;
    hlsRef.current.currentLevel = level;
    setCurrentLevel(level);
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "initializing":
        return "Initializing download...";
      case "downloading":
        return "Downloading...";
      case "streaming":
        return "Streaming ready";
      case "completed":
        return "Download completed";
      default:
        return status;
    }
  };

  return (
    <div className="max-w-7xl mx-auto text-center">
      {/* Download Progress */}
      {downloadProgress && downloadProgress.status !== "completed" && !downloadProgress.stream_ready && (
        <div className="mb-4 p-4 rounded-lg bg-content2 dark:bg-slate-800">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-semibold text-foreground-500 dark:text-slate-200">
              {getStatusLabel(downloadProgress.status)}
            </span>
            <span className="text-sm text-foreground-500 dark:text-slate-200">
              {downloadProgress.progress.toFixed(1)}%
            </span>
          </div>
          <Progress
            value={downloadProgress.progress}
            color="primary"
            className="w-full"
            aria-label="Download progress"
          />
          {downloadProgress.quality && (
            <span className="text-xs text-foreground-400 mt-1 block">
              Quality: {downloadProgress.quality}
            </span>
          )}
        </div>
      )}

      {levels.length > 0 && (
        <div className="mb-3 flex items-center justify-between gap-3">
            <span className="text-start font-bold w-full items-start text-xl text-foreground-500 dark:text-slate-200">{movieTitle && movieTitle}</span>
          <div className="flex w-full justify-end">
            <div
              className="inline-flex overflow-x-auto rounded-xl bg-content2 p-1 shadow-sm dark:bg-slate-800"
              role="radiogroup"
              aria-label="Select playback quality"
            >
            <span className="text-center my-auto p-2 text-sm text-foreground-500 dark:text-slate-200">Quality</span>
              {levels.map((lvl) => {
                const isActive = currentLevel === lvl.index;
                const shortLabel = lvl.index === -1 ? "Auto" : `${lvl.label.split(" ")[0]}`;
                return (
                  <Button
                    className="min-w-14 my-auto"
                    color={isActive ? "primary" : "default"}
                    key={lvl.index}
                    radius="lg"
                    role="radio"
                    size="sm"
                    tabIndex={isActive ? 0 : -1}
                    value={lvl.index.toString()}
                    variant={isActive ? "solid" : "light"}
                    aria-checked={isActive}
                    onPress={() => handleQualityChange(lvl.index)}
                  >
                    {shortLabel}
                  </Button>
                );
              })}
            </div>
          </div>
        </div>
      )}
      <video
        ref={videoRef}
        controls
        autoPlay
        style={{ width: "100%", borderRadius: 10 }}
      >
        <track default kind="captions" label="English" src="/subs/en.vtt" srcLang="en" />
      </video>
    </div>
  );
}
