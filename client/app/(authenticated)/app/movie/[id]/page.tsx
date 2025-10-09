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

interface MoviePageProps {
  params: {
    id: string;
  };
}

export default function MoviePage({ params }: MoviePageProps) {
  const { movie, isLoading, error, toggleWatched } = useMovieDetails(params.id);

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
            <Link href="/discover">
              <Button variant="flat" startContent={<ArrowLeft size={16} />}>
                Back to Discover
              </Button>
            </Link>
          </div>
          <Alert
            color="danger"
            variant="flat"
            title="Movie not found"
            description={error || "The movie you're looking for doesn't exist."}
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
      <div className="container mx-auto px-4 py-8">
        {/* Back Button */}
        <div className="mb-6">
          <Link href="/discover">
            <Button variant="flat" startContent={<ArrowLeft size={16} />}>
              Back to Discover
            </Button>
          </Link>
        </div>

        {/* Movie Header */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
          {/* Poster */}
          <div className="lg:col-span-1">
            <Card className="overflow-hidden">
              <CardBody className="p-0">
                <Image
                  removeWrapper
                  src={movie.posterUrl || "/placeholder-poster.jpg"}
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
                {movie.year && (
                  <span className="text-lg text-foreground-500">{movie.year}</span>
                )}
                {movie.rating && (
                  <div className="flex items-center gap-1">
                    <Star size={20} className="text-warning fill-warning" />
                    <span className="text-lg font-semibold">{movie.rating.toFixed(1)}</span>
                    <span className="text-sm text-foreground-500">IMDb</span>
                  </div>
                )}
              </div>
            </div>

            {/* Genres */}
            {movie.genres && movie.genres.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {movie.genres.map((genre) => (
                  <Chip key={genre} variant="flat" color="primary" size="sm">
                    {genre}
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
            {movie.description && (
              <div>
                <h3 className="text-xl font-semibold mb-3">Plot Summary</h3>
                <p className="text-foreground-600 leading-relaxed">{movie.description}</p>
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
              {movie.director && (
                <div className="flex items-center gap-3">
                  <Film size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Director</p>
                    <p className="font-semibold">{movie.director}</p>
                  </div>
                </div>
              )}

              {/* Duration */}
              {movie.duration && (
                <div className="flex items-center gap-3">
                  <Clock size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Duration</p>
                    <p className="font-semibold">{formatDuration(movie.duration)}</p>
                  </div>
                </div>
              )}

              {/* Release Date */}
              {movie.releaseDate && (
                <div className="flex items-center gap-3">
                  <Calendar size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Release Date</p>
                    <p className="font-semibold">{formatDate(movie.releaseDate)}</p>
                  </div>
                </div>
              )}

              {/* Language */}
              {movie.language && (
                <div className="flex items-center gap-3">
                  <Globe size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Language</p>
                    <p className="font-semibold">{movie.language}</p>
                  </div>
                </div>
              )}

              {/* Budget */}
              {movie.budget && (
                <div className="flex items-center gap-3">
                  <DollarSign size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Budget</p>
                    <p className="font-semibold">{formatCurrency(movie.budget)}</p>
                  </div>
                </div>
              )}

              {/* Revenue */}
              {movie.revenue && (
                <div className="flex items-center gap-3">
                  <DollarSign size={20} className="text-primary" />
                  <div>
                    <p className="text-sm text-foreground-500">Revenue</p>
                    <p className="font-semibold">{formatCurrency(movie.revenue)}</p>
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
                {movie.cast.map((actor, index) => (
                  <div key={index} className="flex items-center gap-3 p-3 rounded-lg bg-content1">
                    <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                      <span className="text-sm font-semibold text-primary">
                        {actor.split(' ').map(n => n[0]).join('')}
                      </span>
                    </div>
                    <span className="font-medium">{actor}</span>
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