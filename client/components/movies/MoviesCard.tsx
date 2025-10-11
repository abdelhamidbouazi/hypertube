"use client";

import React from "react";
import { Card, CardBody, CardFooter } from "@heroui/card";
import { Image } from "@heroui/image";
import { Chip } from "@heroui/chip";
import { Tooltip } from "@heroui/tooltip";
import Link from "next/link";

export type Movie = {
  id: number;
  title: string;
  release_date?: string;
  poster_path?: string;
  overview?: string;
  original_language?: string;
  watched?: boolean;
};

export default function MovieCard({ movie }: { movie: Movie }) {
  const { id, title, release_date, poster_path, watched } = movie;
  const year = release_date ? new Date(release_date).getFullYear() : undefined;

  return (
    <Link href={`/app/movie/${id}`} className="group">
      <Card className="h-full overflow-hidden border border-default-100 bg-content1 transition-transform group-hover:-translate-y-1">
        <CardBody className="p-0">
          <div className="relative">
            <Image
              removeWrapper
              src={poster_path ? `https://image.tmdb.org/t/p/w500${poster_path}` : "/placeholder-poster.jpg"}
              alt={`${title} poster`}
              className="h-64 w-full object-cover"
            />

            {/* Watched badge */}
            <div className="absolute right-2 top-2">
              <Tooltip content={watched ? "Watched" : "Unwatched"}>
                <Chip
                  size="sm"
                  variant={watched ? "solid" : "bordered"}
                  color={watched ? "primary" : "default"}
                >
                  {watched ? "Watched" : "New"}
                </Chip>
              </Tooltip>
            </div>
          </div>
        </CardBody>

        <CardFooter className="flex flex-col items-start gap-1">
          <p className="line-clamp-1 text-small font-medium">{title}</p>
          <p className="text-tiny text-foreground-500">{year ?? "—"}</p>
        </CardFooter>
      </Card>
    </Link>
  );
}
