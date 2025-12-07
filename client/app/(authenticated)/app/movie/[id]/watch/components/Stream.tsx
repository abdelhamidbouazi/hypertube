"use client";

import { useEffect, useRef, useState } from "react";
import Hls from "hls.js";
import { Button } from "@heroui/button";
import { Progress } from "@heroui/progress";
import api, { BASE_URL } from "@/lib/api";

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
  message?: string;
  error?: string;
}

export default function HlsPlayer({ src, token, thumbnail: _thumbnail, movieTitle, movieId }: HlsPlayerProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [levels, setLevels] = useState<{ index: number; label: string }[]>([]);
  const [currentLevel, setCurrentLevel] = useState<number>(-1);
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null);
  const [subtitleTracks, setSubtitleTracks] = useState<{ index: number; label: string; lang: string }[]>([]);
  const [currentSubtitle, setCurrentSubtitle] = useState<number>(-1);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handleLoadedMetadata = () => {
      video.currentTime = 0;
      console.log("Video metadata loaded, reset to 0");
    };

    video.addEventListener('loadedmetadata', handleLoadedMetadata);

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

        if (data.subtitleTracks && data.subtitleTracks.length > 0) {
          const subs = data.subtitleTracks.map((track: any, i: number) => ({
            index: i,
            label: track.name || track.lang || `Subtitle ${i + 1}`,
            lang: track.lang || 'unknown'
          }));
          setSubtitleTracks([{ index: -1, label: "Off", lang: "off" }, ...subs]);
          setCurrentSubtitle(-1);
        }
      });

      hlsInstance.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
        // hls reports the level index (auto resolves to an actual level >=0)
        // We keep -1 to represent "Auto" when ABR is enabled
        const abrEnabled = hlsInstance.autoLevelEnabled;
        setCurrentLevel(abrEnabled ? -1 : data.level);
      });

      hlsInstance.on(Hls.Events.SUBTITLE_TRACK_SWITCH, (_, data) => {
        setCurrentSubtitle(data.id);
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
      video.removeEventListener('loadedmetadata', handleLoadedMetadata);
      // Destroy the local instance created in this effect
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, token]);

  // Reload stream when status becomes 'transcoding'
  useEffect(() => {
    if (downloadProgress?.status === 'transcoding' && hlsRef.current) {
      console.log("Status is transcoding, reloading stream...");
      hlsRef.current.loadSource(src);
      if (videoRef.current) {
        videoRef.current.currentTime = 0;
        videoRef.current.play().catch(e => console.error("Error playing:", e));
      }
    }
  }, [downloadProgress?.status, src]);

  useEffect(() => {
    if (!movieId) return;

    let wsUrl = BASE_URL.replace(/^http/, 'ws');
    if (wsUrl.startsWith('//')) {
      wsUrl = `ws:${wsUrl}`;
    }
    if (wsUrl.startsWith('/')) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      wsUrl = `${protocol}//${window.location.host}${wsUrl}`;
    }

    const socketUrl = `${wsUrl}/ws/${movieId}`;

    console.log("Connecting to WebSocket:", socketUrl);
    const ws = new WebSocket(socketUrl);

    ws.onopen = () => {
      console.log("WebSocket connected");
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        const progressUpdate: DownloadProgress = {
          status: data.stage || data.status || 'initializing',
          progress: data.downloadProgress || data.progress || 0,
          stream_ready: data.stage === 'ready' || data.streamReady || false,
          quality: data.quality || 'unknown',
          message: data.message,
          error: data.error
        };

        if (data.downloadProgress !== undefined) {
          progressUpdate.progress = data.downloadProgress;
        }

        if (data.stage === 'ready') {
          progressUpdate.stream_ready = true;
          progressUpdate.status = 'completed';
        }

        setDownloadProgress(progressUpdate);

      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
    };

    return () => {
      ws.close();
    };
  }, [movieId]);

  const handleQualityChange = (level: number) => {
    if (!hlsRef.current) return;
    hlsRef.current.currentLevel = level;
    setCurrentLevel(level);
  };

  const handleSubtitleChange = (trackIndex: number) => {
    if (!hlsRef.current) return;
    hlsRef.current.subtitleTrack = trackIndex;
    setCurrentSubtitle(trackIndex);
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
      case "error":
        return "Error occurred";
      default:
        return status;
    }
  };

  return (
    <div className="max-w-7xl mx-auto text-center">
      {/* Download Progress */}
      {downloadProgress &&
        downloadProgress.status !== "completed" &&
        !downloadProgress.stream_ready && (
          <div className={`mb-4 p-4 rounded-lg ${downloadProgress.status === 'error' ? 'bg-danger-50 dark:bg-danger-900/20' : 'bg-content2 dark:bg-slate-800'}`}>
            <div className="flex items-center justify-between mb-2">
              <span className={`text-sm font-semibold ${downloadProgress.status === 'error' ? 'text-danger' : 'text-foreground-500 dark:text-slate-200'}`}>
                {getStatusLabel(downloadProgress.status)}
              </span>
              {downloadProgress.status !== 'error' && (
                <span className="text-sm text-foreground-500 dark:text-slate-200">
                  {downloadProgress.progress.toFixed(1)}%
                </span>
              )}
            </div>

            {downloadProgress.status === 'error' ? (
              <div className="text-danger text-sm text-left">
                <p className="font-medium">{downloadProgress.message || "An unexpected error occurred."}</p>
                {downloadProgress.error && <p className="mt-1 opacity-80 text-xs">{downloadProgress.error}</p>}
              </div>
            ) : (
              <>
                <Progress
                  value={downloadProgress.progress}
                  color="primary"
                  className="w-full"
                  aria-label="Download progress"
                />
                {downloadProgress.message && (
                  <p className="text-xs text-foreground-400 mt-2 text-left">{downloadProgress.message}</p>
                )}
                {downloadProgress.quality && (
                  <span className="text-xs text-foreground-400 mt-1 block text-left">
                    Quality: {downloadProgress.quality}
                  </span>
                )}
              </>
            )}
          </div>
        )}

      {levels.length > 0 && (
        <div className="mb-3 flex items-center justify-between gap-3">
          <span className="text-start font-bold w-full items-start text-xl text-foreground-500 dark:text-slate-200">
            {movieTitle && movieTitle}
          </span>
          <div className="flex w-full justify-end gap-3">
            {/* Subtitle Selector */}
            {subtitleTracks.length > 0 && (
              <div
                className="inline-flex overflow-x-auto rounded-xl bg-content2 p-1 shadow-sm dark:bg-slate-800"
                role="radiogroup"
                aria-label="Select subtitles"
              >
                <span className="text-center my-auto p-2 text-sm text-foreground-500 dark:text-slate-200">
                  Subtitles
                </span>
                {subtitleTracks.map((track) => {
                  const isActive = currentSubtitle === track.index;
                  return (
                    <Button
                      className="min-w-14 my-auto"
                      color={isActive ? "primary" : "default"}
                      key={track.index}
                      radius="lg"
                      role="radio"
                      size="sm"
                      tabIndex={isActive ? 0 : -1}
                      value={track.index.toString()}
                      variant={isActive ? "solid" : "light"}
                      aria-checked={isActive}
                      onPress={() => handleSubtitleChange(track.index)}
                    >
                      {track.label}
                    </Button>
                  );
                })}
              </div>
            )}

            {/* Quality Selector */}
            <div
              className="inline-flex overflow-x-auto rounded-xl bg-content2 p-1 shadow-sm dark:bg-slate-800"
              role="radiogroup"
              aria-label="Select playback quality"
            >
              <span className="text-center my-auto p-2 text-sm text-foreground-500 dark:text-slate-200">
                Quality
              </span>
              {levels.map((lvl) => {
                const isActive = currentLevel === lvl.index;
                const shortLabel =
                  lvl.index === -1 ? "Auto" : `${lvl.label.split(" ")[0]}`;
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
      ></video>
    </div>
  );
}
