"use client";

import React, { useState, useRef, useEffect } from "react";
import { Card, CardBody } from "@heroui/card";
import { Button } from "@heroui/button";
import { Slider } from "@heroui/slider";
import { Chip } from "@heroui/chip";
import { Spinner } from "@heroui/spinner";
import { Tooltip } from "@heroui/tooltip";
import { 
  Play, 
  Pause, 
  Volume2, 
  VolumeX, 
  Maximize, 
  Minimize,
  Settings,
  ArrowLeft,
  SkipBack,
  SkipForward,
  RotateCcw,
  Clock,
  Calendar,
  Star
} from "lucide-react";
import Link from "next/link";
import { useMovieDetails } from "@/components/movies/useMovieDetails";

interface WatchPageProps {
  params: {
    id: string;
  };
}

export default function WatchPage({ params }: WatchPageProps) {
  const { movie, isLoading, error } = useMovieDetails(params.id);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showControls, setShowControls] = useState(true);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const videoRef = useRef<HTMLVideoElement>(null);
  const controlsTimeoutRef = useRef<NodeJS.Timeout>();

  // Mock video duration (in seconds)
  const mockDuration = movie?.duration ? movie.duration * 60 : 9000;

  useEffect(() => {
    setDuration(mockDuration);
  }, [mockDuration]);

  const togglePlay = () => {
    setIsPlaying(!isPlaying);
  };

  const handleSeek = (value: number | number[]) => {
    const newTime = Array.isArray(value) ? value[0] : value;
    setCurrentTime(newTime);
  };

  const handleVolumeChange = (value: number | number[]) => {
    const newVolume = Array.isArray(value) ? value[0] : value;
    setVolume(newVolume);
    setIsMuted(newVolume === 0);
  };

  const toggleMute = () => {
    setIsMuted(!isMuted);
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  const skipTime = (seconds: number) => {
    const newTime = Math.max(0, Math.min(duration, currentTime + seconds));
    setCurrentTime(newTime);
  };

  const formatTime = (time: number) => {
    const hours = Math.floor(time / 3600);
    const minutes = Math.floor((time % 3600) / 60);
    const seconds = Math.floor(time % 60);
    
    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  const handleMouseMove = () => {
    setShowControls(true);
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current);
    }
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) {
        setShowControls(false);
      }
    }, 3000);
  };

  const handleMouseLeave = () => {
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current);
    }
    if (isPlaying) {
      setShowControls(false);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background to-content2">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-center min-h-[400px]">
            <Spinner size="lg" label="Loading movie..." />
          </div>
        </div>
      </div>
    );
  }

  if (error || !movie) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background to-content2">
        <div className="container mx-auto px-4 py-8">
          <div className="mb-6">
            <Link href="/discover">
              <Button variant="flat" startContent={<ArrowLeft size={16} />}>
                Back to Discover
              </Button>
            </Link>
          </div>
          <Card className="p-8 text-center">
            <CardBody>
              <h1 className="text-2xl font-bold mb-4">Movie not found</h1>
              <p className="text-foreground-500 mb-6">The movie you're looking for doesn't exist.</p>
              <Link href="/discover">
                <Button color="primary" startContent={<ArrowLeft size={16} />}>
                  Back to Discover
                </Button>
              </Link>
            </CardBody>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-content2">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-6">
          <Link href={`/movie/${params.id}`}>
            <Button variant="flat" startContent={<ArrowLeft size={16} />}>
              Back to Movie
            </Button>
          </Link>
        </div>

        {/* Video Player */}
        <Card className="overflow-hidden border border-default-200 dark:border-default-100 shadow-lg">
          <CardBody className="p-0">
            <div 
              className={`relative bg-gradient-to-br from-content1 to-content2 ${isFullscreen ? 'fixed inset-0 z-50' : 'aspect-video'}`}
              onMouseMove={handleMouseMove}
              onMouseLeave={handleMouseLeave}
            >
              {/* Animated Background Pattern */}
              <div className="absolute inset-0 opacity-5">
                <div className="absolute inset-0 bg-gradient-to-r from-primary/20 via-secondary/20 to-primary/20 animate-pulse" />
                <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_50%_50%,rgba(120,119,198,0.1),transparent_50%)]" />
              </div>

              {/* Video Content */}
              <div className="relative w-full h-full flex items-center justify-center">
                <div className="text-center transform transition-transform duration-300 hover:scale-105">
                  <div className="text-8xl mb-6 animate-bounce">🎬</div>
                  <h1 className="text-4xl font-bold mb-3 bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
                    {movie.title}
                  </h1>
                  <p className="text-foreground-500 text-lg mb-2">Video Player Placeholder</p>
                  <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-content1/50 backdrop-blur-sm border border-default-200 dark:border-default-100">
                    <Clock size={16} className="text-primary" />
                    <span className="text-sm text-foreground-600 font-mono">
                      {formatTime(currentTime)} / {formatTime(duration)}
                    </span>
                  </div>
                </div>

                {/* Play/Pause Overlay */}
                {!isPlaying && (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="relative">
                      <div className="absolute inset-0 bg-primary/20 rounded-full animate-ping" />
                      <Button
                        isIconOnly
                        size="lg"
                        color="primary"
                        className="w-24 h-24 bg-content1/90 backdrop-blur-md border border-default-200 dark:border-default-100 shadow-xl hover:scale-110 transition-transform duration-200"
                        onPress={togglePlay}
                      >
                        <Play size={36} className="ml-1" />
                      </Button>
                    </div>
                  </div>
                )}
              </div>

              {/* Video Controls */}
              <div 
                className={`absolute bottom-0 left-0 right-0 bg-gradient-to-t from-content1/90 to-transparent p-4 transition-opacity duration-300 ${
                  showControls ? 'opacity-100' : 'opacity-0'
                }`}
              >
                {/* Progress Bar */}
                <div className="mb-4">
                  <Slider
                    size="sm"
                    step={1}
                    color="primary"
                    value={currentTime}
                    maxValue={duration}
                    onChange={handleSeek}
                    className="w-full"
                  />
                </div>

                {/* Control Buttons */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {/* Play/Pause */}
                    <Tooltip content={isPlaying ? "Pause" : "Play"}>
                      <Button
                        isIconOnly
                        variant="light"
                        color="default"
                        className="hover:bg-primary/20 transition-colors"
                        onPress={togglePlay}
                      >
                        {isPlaying ? <Pause size={20} /> : <Play size={20} />}
                      </Button>
                    </Tooltip>

                    {/* Skip Controls */}
                    <Tooltip content="Skip backward 10s">
                      <Button
                        isIconOnly
                        variant="light"
                        color="default"
                        className="hover:bg-primary/20 transition-colors"
                        onPress={() => skipTime(-10)}
                      >
                        <SkipBack size={20} />
                      </Button>
                    </Tooltip>

                    <Tooltip content="Skip forward 10s">
                      <Button
                        isIconOnly
                        variant="light"
                        color="default"
                        className="hover:bg-primary/20 transition-colors"
                        onPress={() => skipTime(10)}
                      >
                        <SkipForward size={20} />
                      </Button>
                    </Tooltip>

                    {/* Volume */}
                    <div className="flex items-center gap-2">
                      <Tooltip content={isMuted ? "Unmute" : "Mute"}>
                        <Button
                          isIconOnly
                          variant="light"
                          color="default"
                          className="hover:bg-primary/20 transition-colors"
                          onPress={toggleMute}
                        >
                          {isMuted || volume === 0 ? <VolumeX size={20} /> : <Volume2 size={20} />}
                        </Button>
                      </Tooltip>
                      <div className="w-20">
                        <Slider
                          size="sm"
                          step={0.1}
                          color="primary"
                          value={isMuted ? 0 : volume}
                          maxValue={1}
                          onChange={handleVolumeChange}
                          className="hover:opacity-100 transition-opacity"
                        />
                      </div>
                    </div>

                    {/* Time Display */}
                    <div className="text-foreground text-sm font-mono bg-content1/50 px-3 py-1 rounded-full">
                      {formatTime(currentTime)} / {formatTime(duration)}
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    {/* Playback Speed */}
                    <div className="flex items-center gap-2">
                      <Tooltip content="Change playback speed">
                        <Button
                          isIconOnly
                          variant="light"
                          color="default"
                          className="hover:bg-primary/20 transition-colors"
                          onPress={() => {
                            const speeds = [0.5, 0.75, 1, 1.25, 1.5, 2];
                            const currentIndex = speeds.indexOf(playbackSpeed);
                            const nextIndex = (currentIndex + 1) % speeds.length;
                            setPlaybackSpeed(speeds[nextIndex]);
                          }}
                        >
                          <RotateCcw size={16} />
                        </Button>
                      </Tooltip>
                      <span className="text-foreground text-sm bg-content1/50 px-2 py-1 rounded">{playbackSpeed}x</span>
                    </div>

                    {/* Settings */}
                    <Tooltip content="Settings">
                      <Button
                        isIconOnly
                        variant="light"
                        color="default"
                        className="hover:bg-primary/20 transition-colors"
                      >
                        <Settings size={20} />
                      </Button>
                    </Tooltip>

                    {/* Fullscreen */}
                    <Tooltip content={isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}>
                      <Button
                        isIconOnly
                        variant="light"
                        color="default"
                        className="hover:bg-primary/20 transition-colors"
                        onPress={toggleFullscreen}
                      >
                        {isFullscreen ? <Minimize size={20} /> : <Maximize size={20} />}
                      </Button>
                    </Tooltip>
                  </div>
                </div>
              </div>

              {/* Top Info */}
              <div 
                className={`absolute top-0 left-0 right-0 bg-gradient-to-b from-content1/90 to-transparent p-4 transition-opacity duration-300 ${
                  showControls ? 'opacity-100' : 'opacity-0'
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2">
                      <Calendar size={16} className="text-primary" />
                      <Chip variant="flat" color="primary" size="sm" className="font-semibold">
                        {movie.year}
                      </Chip>
                    </div>
                    {movie.rating && (
                      <div className="flex items-center gap-2">
                        <Star size={16} className="text-warning fill-warning" />
                        <Chip variant="flat" color="warning" size="sm" className="font-semibold">
                          {movie.rating.toFixed(1)}
                        </Chip>
                      </div>
                    )}
                    {movie.duration && (
                      <div className="flex items-center gap-2">
                        <Clock size={16} className="text-secondary" />
                        <Chip variant="flat" color="secondary" size="sm" className="font-semibold">
                          {Math.floor(movie.duration / 60)}h {movie.duration % 60}m
                        </Chip>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>

      </div>
    </div>
  );
}