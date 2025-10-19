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
import StreamPlayer from "./components/Stream";
import HlsPlayer from "./components/Stream";
import Image from "next/image";

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

  return (
    <div className="min-h-screen ">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-6">
          <Link href={`/movie/${params.id}`}>
            <Button variant="flat" startContent={<ArrowLeft size={16} />}>
              Back to Movie
            </Button>
          </Link>
        </div>
        </div>
        <HlsPlayer src='https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8' token={localStorage.getItem('token') || ''} thumbnail={`https://image.tmdb.org/t/p/w500${movie.backdrop_path}` || ''} />
        {/* <HlsPlayer src={`${process.env.NEXT_PUBLIC_API_URL}/stream/${params.id}`} token={localStorage.getItem('token') || ''} thumbnail={`https://image.tmdb.org/t/p/w500${movie.backdrop_path}` || ''} /> */}

        {/* {movie ? `https://image.tmdb.org/t/p/w500${movie.backdrop_path}` : 'loading movies'} */}
        {/* <Image src={`https://image.tmdb.org/t/p/w500${movie.backdrop_path}` || ''} alt={movie?.title || ''} width={1000} height={1000} /> */}
    </div>
  );
}