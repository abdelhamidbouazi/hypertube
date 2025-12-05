"use client";

import React, { useState, useRef, useEffect, use } from "react";
import { Card, CardBody } from "@heroui/card";
import { Button } from "@heroui/button";
import { Slider } from "@heroui/slider";
import { Chip } from "@heroui/chip";
import { Spinner } from "@heroui/spinner";
import { Tooltip } from "@heroui/tooltip";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { useMovieDetails } from "@/components/movies/useMovieDetails";
// import StreamPlayer from "./components/Stream";
import HlsPlayer from "./components/Stream";
// import Image from "next/image";
import { BASE_URL } from "@/lib/api";

interface WatchPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default function WatchPage({ params }: WatchPageProps) {
  const { id } = use(params);
  const { movie, isLoading, error } = useMovieDetails(id);
  useEffect(() => {
    if (movie) {
      console.log("movie", movie);
    }
  }, [movie])


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
      <div className="container mx-auto  py-2">
        {/* Header */}
        <div className="mb-2">
          <Link href={`/app/movie/${id}`}>
            <Button startContent={<ArrowLeft size={16} />} variant="solid" color="primary" >
              Back to Movie Details
            </Button>
          </Link>
        </div>
      </div>
      {/* {
        movie.isAvailable ? (
          <HlsPlayer
            src={BASE_URL + `/stream/${id}/master.m3u8`}
            token={localStorage.getItem("token") || ""}
            movieTitle={movie.title}
            thumbnail={
              `https://image.tmdb.org/t/p/w500${movie.backdrop_path}` || ""
            }
          />
        ) : (
          
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-center min-h-[400px]">
            <Spinner size="lg" label={`Downloading ${movie.title}...`} />
            
          </div>
        </div>
        )
      } */}
      <HlsPlayer
        src={BASE_URL + `/stream/${id}/master.m3u8`}
        token={localStorage.getItem("token") || ""}
        movieTitle={movie.title}
        movieId={id}
        thumbnail={
          `https://image.tmdb.org/t/p/w500${movie.poster_path}` || ""
        }
      />
    </div>
  );
}
