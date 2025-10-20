"use client";

import React from "react";
import { Card, CardBody } from "@heroui/card";
import { Image } from "@heroui/image";
import { Button } from "@heroui/button";
import { Chip } from "@heroui/chip";
import { Divider } from "@heroui/divider";
import { Spinner } from "@heroui/spinner";
import { Alert } from "@heroui/alert";
import { 
  ArrowLeft, 
  Play, 
  Clock, 
  Calendar, 
  Globe, 
  DollarSign, 
  Star,
  Users,
  Film,
  Eye,
  EyeOff
} from "lucide-react";
import Link from "next/link";
import { useMovieDetails } from "@/components/movies/useMovieDetails";
import { getErrorMessage } from "@/lib/error-utils";

interface MoviePageProps {
  params: {
    id: string;
  };
}

export default function MoviePage({ params }: MoviePageProps) {
  const { movie, isLoading, error } = useMovieDetails(params.id);

  const toggleWatched = () => {
    console.log("toggle watched for movie:", movie?.id);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background to-content2">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-center min-h-[400px]">
            <Spinner size="lg" label="Loading movie details..." />
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
            <Link href="/app/discover">
              <Button variant="flat" startContent={<ArrowLeft size={16} />}>
                Back to Discover
              </Button>
            </Link>
          </div>
          <Alert
            color="danger"
            variant="flat"
            title="Movie not found"
            description={error ? getErrorMessage(error) : "The movie you're looking for doesn't exist."}
          />
        </div>
      </div>
    );
  }

  const formatCurrency = (amount?: number) => {
    if (!amount) return "N/A";
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatDuration = (minutes?: number) => {
    if (!minutes) return "N/A";
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return hours > 0 ? `${hours}h ${mins}m` : `${mins}m`;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-content2">
      <div className="container mx-auto px-4 py-2">
        

        {/* Movie Header */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
          {/* Poster */}
          <div className="lg:col-span-1">
            <Card className="overflow-hidden">
              <CardBody className="p-0">
                <Image
                  removeWrapper
                  src={movie.poster_path ? `https://image.tmdb.org/t/p/w500${movie.poster_path}` : "/placeholder-poster.jpg"}
                  alt={`${movie.title} poster`}
                  className="w-full h-auto object-cover"
                />
              </CardBody>
            </Card>
          </div>

          {/* Movie Info */}
          <div className="lg:col-span-2 space-y-6">
            {/* Title and Rating */}
            <div>
              <h1 className="text-4xl font-bold mb-2">{movie.title}</h1>
              <div className="flex items-center gap-4 mb-4">
                {movie.release_date && (
                  <span className="text-lg text-foreground-500">{new Date(movie.release_date).getFullYear()}</span>
                )}
                {movie.vote_average && (
                  <div className="flex items-center gap-1">
                    <Star size={20} className="text-warning fill-warning" />
                    <span className="text-lg font-semibold">{movie.vote_average.toFixed(1)}</span>
                    <span className="text-sm text-foreground-500">TMDB</span>
                  </div>
                )}
              </div>
            </div>

            {/* Genres */}
            {movie.genres && movie.genres.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {movie.genres.map((genre: any) => (
                  <Chip key={genre.id} variant="flat" color="primary" size="sm">
                    {genre.name}
                  </Chip>
                ))}
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex gap-3">
            <Link href={`/app/movie/${movie.id}/watch`}>
                <Button
                  color="primary"
                  size="lg"
                  startContent={<Play size={20} />}
                  className="flex-1"
                >
                  Watch Now
                </Button>
              </Link>
              <Button
                variant={movie.watched ? "solid" : "bordered"}
                color={movie.watched ? "success" : "default"}
                size="lg"
                startContent={movie.watched ? <EyeOff size={20} /> : <Eye size={20} />}
                onPress={toggleWatched}
              >
                {movie.watched ? "Watched" : "Mark as Watched"}
              </Button>
            </div>

            {/* Description */}
            {movie.overview && (
              <div>
                <h3 className="text-xl font-semibold mb-3">Plot Summary</h3>
                <p className="text-foreground-600 leading-relaxed">{movie.overview}</p>
              </div>
            )}
          </div>
        </div>

        {/* Movie Details */}
        <Card className="mb-8">
          <CardBody className="p-6">
            <h2 className="text-2xl font-bold mb-6">Movie Details</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {/* Director */}
              {movie.director && movie.director.length > 0 && (
                <div className="flex items-center gap-3">
                  <Film size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Director</p>
                    <p className="font-semibold">{movie.director.map((d: any) => d.name).join(', ')}</p>
                  </div>
                </div>
              )}

              {/* Duration */}
              {movie.runtime && (
                <div className="flex items-center gap-3">
                  <Clock size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Duration</p>
                    <p className="font-semibold">{formatDuration(movie.runtime)}</p>
                  </div>
                </div>
              )}

              {/* Release Date */}
              {movie.release_date && (
                <div className="flex items-center gap-3">
                  <Calendar size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Release Date</p>
                    <p className="font-semibold">{formatDate(movie.release_date)}</p>
                  </div>
                </div>
              )}

              {/* Language */}
              {movie.original_language && (
                <div className="flex items-center gap-3">
                  <Globe size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Language</p>
                    <p className="font-semibold">{movie.original_language.toUpperCase()}</p>
                  </div>
                </div>
              )}
            </div>
          </CardBody>
        </Card>

        {/* Cast */}
        {movie.cast && movie.cast.length > 0 && (
          <Card>
            <CardBody className="p-6">
              <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
                <Users size={24} className="text-primary" />
                Cast
              </h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {movie.cast.map((actor: any) => (
                  <div key={actor.id} className="flex items-center gap-3 p-3 rounded-lg bg-content1">
                    <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                      <span className="text-sm font-semibold text-primary">
                        {actor.name.split(' ').map((n: string) => n[0]).join('')}
                      </span>
                    </div>
                    <div>
                      <span className="font-medium">{actor.name}</span>
                      <p className="text-sm text-foreground-500">{actor.character}</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardBody>
          </Card>
        )}
      </div>
    </div>
  );
}