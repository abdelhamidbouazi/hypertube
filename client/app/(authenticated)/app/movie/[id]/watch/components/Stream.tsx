"use client";

import { useEffect, useRef, useState } from "react";
import Hls from "hls.js";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";
import { Download, AlertCircle, Video } from "lucide-react";
import { BASE_URL } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

interface HlsPlayerProps {
  src: string;
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

// Convert HTTP base URL to WebSocket URL
function getWebSocketUrl(baseUrl: string, movieId: string, token?: string): string {
  let wsUrl = baseUrl.replace(/^http/, "ws");
  
  if (wsUrl.startsWith("//")) {
    wsUrl = `ws:${wsUrl}`;
  } else if (wsUrl.startsWith("/")) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    wsUrl = `${protocol}//${window.location.host}${wsUrl}`;
  }
  
  let url = `${wsUrl}/ws/${movieId}`;
  if (token) {
    url += `?token=${encodeURIComponent(token)}`;
  }
  
  return url;
}

// Get readable status label
function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    initializing: "Initializing...",
    downloading: "Downloading...",
    transcoding: "Transcoding...",
    streaming: "Streaming ready",
    completed: "Completed",
    error: "Error",
  };
  return labels[status] || status;
}

// Parse WebSocket message into DownloadProgress
function parseProgressMessage(data: any): DownloadProgress {
  const isReady = data.stage === "ready" || data.streamReady;
  
  return {
    status: isReady ? "completed" : data.stage || data.status || "initializing",
    progress: data.downloadProgress ?? data.progress ?? 0,
    stream_ready: isReady,
    quality: data.quality || "unknown",
    message: data.message,
    error: data.error,
  };
}

export default function HlsPlayer({ src, token, movieTitle, movieId }: HlsPlayerProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [levels, setLevels] = useState<{ index: number; label: string }[]>([]);
  const [currentLevel, setCurrentLevel] = useState<number>(-1);
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null);
  const [subtitleTracks, setSubtitleTracks] = useState<{ index: number; label: string; lang: string }[]>([]);
  const [currentSubtitle, setCurrentSubtitle] = useState<number>(-1);

  const preferredLanguage = useAuthStore((state) => state.user?.preferred_language);

  // Initialize HLS player and setup quality/subtitle tracks
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handleLoadedMetadata = () => {
      video.currentTime = 0;
    };

    video.addEventListener("loadedmetadata", handleLoadedMetadata);

    if (Hls.isSupported()) {
      const hlsInstance = new Hls({
        xhrSetup: (xhr) => {
          if (token) {
            xhr.setRequestHeader("Authorization", `Bearer ${token}`);
          }
        },
      });

      hlsInstance.loadSource(src);
      hlsInstance.attachMedia(video);

      // Setup quality levels from manifest
      hlsInstance.on(Hls.Events.MANIFEST_PARSED, (_, data: any) => {
        const qualityLevels = data.levels.map((lvl: any, i: number) => ({
          index: i,
          label: `${lvl.height}p (${Math.round(lvl.bitrate / 1000)} kbps)`,
        }));
        setLevels([{ index: -1, label: "Auto" }, ...qualityLevels]);
        setCurrentLevel(-1);

        // Setup subtitle tracks
        if (data.subtitleTracks?.length > 0) {
          // Map HLS subtitle tracks directly 
          const hlsSubtitleTracks = data.subtitleTracks.map((track: any, i: number) => ({
            index: i, // HLS track index 
            label: track.name || track.lang || `Subtitle ${i + 1}`,
            lang: track.lang || "en",
          }));

          // Add "Off" option at index -1 (HLS.js uses -1 to disable subtitles)
          setSubtitleTracks([{ index: -1, label: "Off", lang: "off" }, ...hlsSubtitleTracks]);

          // Auto-select preferred language if available
          // Find the HLS track index (0-based) for the preferred language
          let hlsSubtitleIndex = -1;
          
          if (preferredLanguage) {
            // Try to find the user's preferred language
            const preferredHlsIndex = hlsSubtitleTracks.findIndex((sub: { lang: string }) => 
              sub.lang?.toLowerCase() === preferredLanguage.toLowerCase()
            );
            if (preferredHlsIndex >= 0) {
              hlsSubtitleIndex = preferredHlsIndex;
            }
          }
          else {
            hlsSubtitleIndex = 0;
          }
          
          // If preferred language not found, default to English
          if (hlsSubtitleIndex === -1) {
            const englishIndex = hlsSubtitleTracks.findIndex((sub: { lang: string }) => 
              sub.lang?.toLowerCase() === "en"
            );
            if (englishIndex >= 0) {
              hlsSubtitleIndex = englishIndex;
            } else if (hlsSubtitleTracks.length > 0) {
              // If English not available, use the first available subtitle track
              hlsSubtitleIndex = 0;
            }
            // If no subtitles available, hlsSubtitleIndex remains -1 (Off)
          }
          
          hlsInstance.subtitleTrack = hlsSubtitleIndex;
          setCurrentSubtitle(hlsSubtitleIndex);
        } else {
          setSubtitleTracks([]);
          setCurrentSubtitle(-1);
        }
      });

      // Track quality level changes
      hlsInstance.on(Hls.Events.LEVEL_SWITCHED, (_, data: any) => {
        setCurrentLevel(hlsInstance.autoLevelEnabled ? -1 : data.level);
      });

      // Track subtitle changes
      // data.id is the HLS track index (-1 for Off, 0, 1, 2... for tracks)
      hlsInstance.on(Hls.Events.SUBTITLE_TRACK_SWITCH, (_, data: any) => {
        setCurrentSubtitle(data.id); // data.id is already the HLS track index
      });

      hlsRef.current = hlsInstance;
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // Safari native HLS support (token as query param)
      video.src = token ? `${src}?token=${token}` : src;
    }

    return () => {
      video.removeEventListener("loadedmetadata", handleLoadedMetadata);
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, token, preferredLanguage]);

  // Reload stream when transcoding completes
  useEffect(() => {
    if (downloadProgress?.status === "transcoding" && hlsRef.current && videoRef.current) {
      hlsRef.current.loadSource(src);
      videoRef.current.currentTime = 0;
      videoRef.current.play().catch(() => {
        addToast({
          title: "Playback error",
          description: "Failed to play video. Please try again.",
          severity: "warning",
          timeout: 3000,
        });
      });
    }
  }, [downloadProgress?.status, src]);

  
  // WebSocket connection for download/transcoding progress
  useEffect(() => {
    if (!movieId) return;

    const ws = new WebSocket(getWebSocketUrl(BASE_URL, movieId, token));

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setDownloadProgress(parseProgressMessage(data));
      } catch (err) {
        addToast({
          title: "Connection error",
          description: "Failed to parse server message. Please refresh the page.",
          severity: "warning",
          timeout: 4000,
        });
      }
    };

    ws.onerror = () => {
      addToast({
        title: "Connection error",
        description: "Lost connection to server. Please refresh the page.",
        severity: "warning",
        timeout: 4000,
      });
    };

    return () => ws.close();
  }, [movieId, token]);

  const handleQualityChange = (level: number) => {
    if (hlsRef.current) {
      hlsRef.current.currentLevel = level;
      setCurrentLevel(level);
    }
  };

  const handleSubtitleChange = (hlsTrackIndex: number) => {
    if (hlsRef.current) {
      // Use HLS track index directly: -1 for "Off", 0, 1, 2... for subtitle tracks
      hlsRef.current.subtitleTrack = hlsTrackIndex;
      setCurrentSubtitle(hlsTrackIndex);
    }
  };

  const showProgress = downloadProgress && 
    downloadProgress.status !== "completed" && 
    !downloadProgress.stream_ready;

  return (
    <div className="max-w-7xl mx-auto text-center">
      {levels.length > 0 && (
        <div className="mb-3 flex items-center justify-between gap-3">
          <span className="text-start font-bold w-full items-start text-xl text-foreground-500 dark:text-slate-200">
            {movieTitle}
          </span>
          <div className="flex w-full justify-end gap-3">
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
                const shortLabel = lvl.index === -1 ? "Auto" : lvl.label.split(" ")[0];
                return (
                  <Button
                    key={lvl.index}
                    className="min-w-14 my-auto"
                    color={isActive ? "primary" : "default"}
                    radius="lg"
                    role="radio"
                    size="sm"
                    tabIndex={isActive ? 0 : -1}
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
      />

      {showProgress && (
        <div
          className={`mt-4 p-4 rounded-lg border ${
            downloadProgress.status === "error"
              ? "bg-danger-50 dark:bg-danger-950/30 border-danger-200 dark:border-danger-800"
              : "bg-content2 dark:bg-content2 border-default-200 dark:border-default-100"
          }`}
        >
          <div className="flex items-start gap-3">
            <div
              className={`flex-shrink-0 ${
                downloadProgress.status === "error"
                  ? "text-danger dark:text-danger-400"
                  : "text-primary dark:text-primary-400"
              }`}
            >
              {downloadProgress.status === "error" ? (
                <AlertCircle className="w-5 h-5" />
              ) : downloadProgress.status === "transcoding" ? (
                <Video className="w-5 h-5 animate-pulse" />
              ) : (
                <Download className="w-5 h-5 animate-pulse" />
              )}
            </div>

            <div className="flex-1 space-y-1.5">
              <div className="flex items-center gap-2 flex-wrap">
                <span
                  className={`text-sm font-medium ${
                    downloadProgress.status === "error"
                      ? "text-danger dark:text-danger-400"
                      : "text-foreground dark:text-foreground"
                  }`}
                >
                  {getStatusLabel(downloadProgress.status)}
                </span>
                {downloadProgress.message && (
                  <p
                    className={`text-sm ${
                      downloadProgress.status === "error"
                        ? "text-danger-600 dark:text-danger-400"
                        : "text-foreground-600 dark:text-foreground-400"
                    }`}
                  >
                    {downloadProgress.message}
                  </p>
                )}
                {downloadProgress.quality && downloadProgress.quality !== "unknown" && (
                  <span className="text-xs text-foreground-500 dark:text-foreground-400">
                    • {downloadProgress.quality}
                  </span>
                )}
              </div>

            </div>
          </div>
        </div>
      )}
    </div>
  );
}
